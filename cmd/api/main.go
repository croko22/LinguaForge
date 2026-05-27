package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "modernc.org/sqlite"

	"github.com/croko/language-app/internal/config"
	"github.com/croko/language-app/internal/handler"
	"github.com/croko/language-app/internal/parser"
	"github.com/croko/language-app/internal/repository"
	"github.com/croko/language-app/internal/service"
	"github.com/croko/language-app/internal/storage"
	"github.com/croko/language-app/internal/translator"
)

func main() {
	// 1. Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// 2. Open SQLite database
	db, err := sql.Open("sqlite", cfg.DatabasePath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// 3. Run migrations
	if err := repository.RunMigrations(db); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	// 4. Create dependencies
	docRepo := repository.NewDocumentRepository(db)
	chRepo := repository.NewChapterRepository(db)
	fileStorage, err := storage.NewLocalFileStorage(cfg.UploadDir)
	if err != nil {
		log.Fatalf("failed to create file storage: %v", err)
	}
	epubParser := parser.NewEpubParser()
	docService := service.NewDocumentService(docRepo, chRepo, fileStorage, epubParser)
	docHandler := handler.NewDocumentHandler(docService)
	trans := translator.NewMockTranslator()
	transHandler := handler.NewTranslateHandler(trans)

	// 5. Setup chi router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	docHandler.RegisterRoutes(r)
	r.Post("/api/translate", transHandler.Translate)

	// 6. Health check
	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// 7. Start server with graceful shutdown
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Graceful shutdown goroutine
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}()

	log.Printf("language-app api listening on :%d", cfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
	log.Println("server stopped")
}
