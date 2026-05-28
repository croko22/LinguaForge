//go:build wails

package main

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log"
	"net/http"
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
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:     "LinguaForge",
		Width:     1200,
		Height:    800,
		MinWidth:  800,
		MinHeight: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup:  app.startup,
		OnShutdown: app.shutdown,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		log.Fatalf("Desktop app failed: %v", err)
	}
}

type App struct {
	ctx    context.Context
	server *http.Server
}

func NewApp() *App { return &App{} }

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := sql.Open("sqlite", cfg.DatabasePath)
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	if err := repository.RunMigrations(db); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	docRepo := repository.NewDocumentRepository(db)
	chRepo := repository.NewChapterRepository(db)
	fileStorage, err := storage.NewLocalFileStorage(cfg.UploadDir)
	if err != nil {
		log.Fatalf("storage: %v", err)
	}
	docService := service.NewDocumentService(docRepo, chRepo, fileStorage, parser.NewEpubParser())
	docHandler := handler.NewDocumentHandler(docService)
	transProvider := translator.NewProvider(translator.DefaultSettings())
	transHandler := handler.NewTranslateHandler(transProvider)
	settingsHandler := handler.NewSettingsHandler(transProvider)

	r := chi.NewRouter()
	r.Use(middleware.Logger, middleware.Recoverer, middleware.Timeout(30*time.Second))
	docHandler.RegisterRoutes(r)
	r.Post("/api/translate", transHandler.Translate)
	r.Get("/api/settings", settingsHandler.GetSettings)
	r.Put("/api/settings", settingsHandler.UpdateSettings)
	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte(`{"status":"ok"}`)) })

	port := 8080
	if cfg.Port != 0 {
		port = cfg.Port
	}
	a.server = &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: r, ReadHeaderTimeout: 10 * time.Second}
	go func() {
		log.Printf("API on :%d", port)
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("server: %v", err)
		}
	}()
}

func (a *App) shutdown(ctx context.Context) {
	if a.server != nil {
		c, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		a.server.Shutdown(c)
	}
}
