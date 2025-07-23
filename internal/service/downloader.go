package service

import (
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type DownloaderService struct {
	in  chan *LoadTask
	out chan *LoadTask
}

func NewLoaderService(size int) *DownloaderService {
	return &DownloaderService{
		in:  make(chan *LoadTask, size),
		out: make(chan *LoadTask, size),
	}
}

type LoadTask struct {
	ID    string
	Link  string
	Error string
}

func (s *DownloaderService) StartDownloader(workers int) {
	for i := 0; i < workers; i++ {
		go func() {
			for task := range s.in {
				if s.Download(task.ID, task.Link) {
					s.out <- &LoadTask{
						ID:    task.ID,
						Link:  task.Link,
						Error: "ok",
					}
				} else {
					s.out <- &LoadTask{
						ID:    task.ID,
						Link:  task.Link,
						Error: "failed to download file",
					}
				}
			}
		}()
	}
}

func (s *DownloaderService) Download(session string, url string) bool {
	sessionDir := os.Getenv("SESSION_DIR")

	head, err := http.Head(url)
	if err != nil {
		return false
	}
	switch head.Header["Content-Type"][0] {
	case "image/jpeg":
	case "image/png":
	case "application/pdf":
	default:
		return false
	}

	resp, err := http.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	splitUrl := strings.Split(url, "/")
	fileName := splitUrl[len(splitUrl)-1]

	outPath := filepath.Join(sessionDir, session, fileName)
	out, err := os.Create(outPath)
	if err != nil {
		slog.Error("Failed to create file", "error", err)
		return false
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)

	return err == nil
}

func (s *DownloaderService) AddTask(task *LoadTask) {
	s.in <- task
}

func (s *DownloaderService) Out() <-chan *LoadTask {
	return s.out
}
