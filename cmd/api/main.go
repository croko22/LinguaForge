package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
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
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// 2. Open SQLite database
	db, err := sql.Open("sqlite", cfg.DatabasePath)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// 3. Run migrations
	if err := repository.RunMigrations(db); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	// 4. Create dependencies
	docRepo := repository.NewDocumentRepository(db)
	chRepo := repository.NewChapterRepository(db)
	fileStorage, err := storage.NewLocalFileStorage(cfg.UploadDir)
	if err != nil {
		slog.Error("failed to create file storage", "error", err)
		os.Exit(1)
	}
	epubParser := parser.NewEpubParser()
	docService := service.NewDocumentService(docRepo, chRepo, fileStorage, []parser.Parser{epubParser})
	docHandler := handler.NewDocumentHandler(docService)
	wordRepo := repository.NewWordRepository(db)
	wordService := service.NewWordService(wordRepo)
	wordHandler := handler.NewWordHandler(wordService)
	transSettings := translator.DefaultSettings()
	transProvider := translator.NewProvider(transSettings)
	transHandler := handler.NewTranslateHandler(translator.NewCachedTranslator(transProvider))
	settingsHandler := handler.NewSettingsHandler(transProvider)

	// 5. Setup chi router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	docHandler.RegisterRoutes(r)
	r.Post("/api/translate", transHandler.Translate)
	r.Get("/api/settings", settingsHandler.GetSettings)
	r.Put("/api/settings", settingsHandler.UpdateSettings)
	r.Post("/api/words", wordHandler.SaveWord)
	r.Get("/api/words", wordHandler.ListWords)
	r.Delete("/api/words/{id}", wordHandler.DeleteWord)

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
		slog.Info("shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}()

	slog.Info("server starting", "port", cfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
	slog.Info("server stopped")
}
