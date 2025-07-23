package app

import (
	"os"
	"zipcollector/internal/config"
	"zipcollector/internal/handler"
	"zipcollector/internal/service"

	_ "zipcollector/docs"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
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

	router.Use(chiMiddleware.Logger)

	router.Get("/docs/*", httpSwagger.WrapHandler)

	router.Post("/start", collector.HandleStartTask)
	router.Post("/add/{taskID}", collector.HandleAddedLink)
	router.Get("/status/{taskID}", collector.HandleGetStatus)
	router.Get("/zip/{taskID}", collector.HandleGetZip)
	router.Delete("/remove/{taskID}", collector.HandleRemoveTask)
	router.Get("/archive/{taskID}.zip", collector.HandleServeArchive)

	return &App{
		Router: router,
	}
}
