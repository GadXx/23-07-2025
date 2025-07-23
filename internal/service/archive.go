package service

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"zipcollector/internal/model"

	"log/slog"
	"zipcollector/internal/config"
)

type ArchiveService struct {
	tasks      map[string]*model.Task
	mu         sync.Mutex
	sem        chan struct{}
	config     *config.Config
	downloader *DownloaderService
}

func NewArchiveService(cfg *config.Config, downloader *DownloaderService) *ArchiveService {
	return &ArchiveService{
		tasks:      make(map[string]*model.Task),
		sem:        make(chan struct{}, cfg.MaxActiveTasks),
		config:     cfg,
		downloader: downloader,
	}
}

func (s *ArchiveService) NewTask(task *model.Task) *model.Task {
	s.sem <- struct{}{}
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tasks[task.ID] = task
	err := os.Mkdir(filepath.Join(os.Getenv("SESSION_DIR"), task.ID), 0755)
	if err != nil {
		slog.Error("Failed to create session directory for new task", "taskID", task.ID, "error", err)
	} else {
		slog.Info("Session directory created for new task", "taskID", task.ID)
	}

	return s.tasks[task.ID]
}

func (s *ArchiveService) AddLink(taskID string, link string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[taskID]
	if !ok {
		slog.Error("No such task in AddLink", "taskID", taskID)
		return fmt.Errorf("no such task: %s", taskID)
	}
	task.Links[link] = struct{}{}
	slog.Info("Link added to task", "taskID", taskID, "link", link)
	return nil
}

func (s *ArchiveService) SendForLoadTasks(taskID string) error {
	s.mu.Lock()
	task, ok := s.tasks[taskID]
	s.mu.Unlock()
	if !ok {
		slog.Error("SendForLoadTasks: task not found", "taskID", taskID)
		return fmt.Errorf("task not found: %s", taskID)
	}
	for link := range task.Links {
		slog.Info("Sending link for download", "taskID", taskID, "link", link)
		s.downloader.AddTask(&LoadTask{ID: taskID, Link: link})
	}
	return nil
}

func (s *ArchiveService) DownloadDistributor() {
	for task := range s.downloader.Out() {
		s.mu.Lock()
		s.tasks[task.ID].Downloaded[task.Link] = task.Error
		slog.Info("Download result received", "taskID", task.ID, "link", task.Link, "result", task.Error)
		s.mu.Unlock()
	}
}

func (s *ArchiveService) GetStatus(taskID string) map[string]string {
	s.mu.Lock()
	task := s.tasks[taskID]
	s.mu.Unlock()
	if task == nil {
		slog.Error("GetStatus: task not found", "taskID", taskID)
		return nil
	}
	slog.Info("GetStatus called", "taskID", taskID, "downloaded", task.Downloaded)
	return task.Downloaded
}

func (s *ArchiveService) RemoveTask(taskID string) {
	s.mu.Lock()
	delete(s.tasks, taskID)
	s.mu.Unlock()
	<-s.sem
}

func (s *ArchiveService) GetTask(taskID string) *model.Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	task := s.tasks[taskID]
	if task == nil {
		slog.Error("GetTask: task not found", "taskID", taskID)
	}
	return task
}

func (s *ArchiveService) ZipArchive(taskID string) string {
	sourceDir := filepath.Join(os.Getenv("SESSION_DIR"), taskID)
	archiveDir := "./archive"
	err := os.MkdirAll(archiveDir, 0755)
	if err != nil {
		slog.Error("Failed to create archive directory", "error", err)
		return ""
	}

	archivePath := filepath.Join(archiveDir, taskID+".zip")

	zipFile, err := os.Create(archivePath)
	if err != nil {
		slog.Error("Failed to create archive file", "archivePath", archivePath, "error", err)
		return ""
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	walkErr := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			slog.Error("Walk error in ZipArchive", "path", path, "error", err)
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			slog.Error("Failed to get relative path in ZipArchive", "error", err)
			return err
		}

		f, err := os.Open(path)
		if err != nil {
			slog.Error("Failed to open file for zipping", "path", path, "error", err)
			return err
		}
		defer f.Close()

		w, err := zipWriter.Create(relPath)
		if err != nil {
			slog.Error("Failed to create entry in zip", "relPath", relPath, "error", err)
			return err
		}
		_, err = io.Copy(w, f)
		if err != nil {
			slog.Error("Failed to copy file to zip", "relPath", relPath, "error", err)
		}
		return err
	})

	if walkErr != nil {
		slog.Error("Failed to walk source directory for zip", "sourceDir", sourceDir, "error", walkErr)
		return ""
	}
	return archivePath
}
