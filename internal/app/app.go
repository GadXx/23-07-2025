package app

import (
	"os"
	"zipcollector/internal/config"
	"zipcollector/internal/handler"
	"zipcollector/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type App struct {
	Router *chi.Mux
}

func NewApp() *App {

	SESSION_DIR := os.Getenv("SESSION_DIR")
	ARCHIVE_DIR := os.Getenv("ARCHIVE_DIR")
	_ = os.MkdirAll(SESSION_DIR, 0777)
	_ = os.MkdirAll(ARCHIVE_DIR, 0777)

	cfg := config.DefaultConfig()
	downloader := service.NewLoaderService(config.GetEnvInt("QUEUE_SIZE", 100))
	archive := service.NewArchiveService(cfg, downloader)

	downloader.StartDownloader(config.GetEnvInt("DOWNLOADER_WORKERS", 10))
	go archive.DownloadDistributor()

	collector := handler.NewCollectorHandler(archive)

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Post("/start", collector.HandleStartTask)
	router.Post("/add/{taskID}", collector.AddedLink)
	router.Get("/status/{taskID}", collector.HandleGetStatus)
	router.Get("/zip/{taskID}", collector.GetZip)
	router.Delete("/remove/{taskID}", collector.HandleRemoveTask)
	router.Get("/archive/{taskID}.zip", collector.ServeArchive)

	return &App{
		Router: router,
	}
}
