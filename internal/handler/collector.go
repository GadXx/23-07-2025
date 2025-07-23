package handler

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"zipcollector/internal/config"
	"zipcollector/internal/model"
	"zipcollector/internal/service"

	"errors"

	"github.com/go-chi/chi/v5"
)

type CollectorHandler struct {
	archiveService *service.ArchiveService
}

func NewCollectorHandler(archiveService *service.ArchiveService) *CollectorHandler {
	return &CollectorHandler{
		archiveService: archiveService,
	}
}

func (h *CollectorHandler) HandleStartTask(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Links []string `json:"links"`
	}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		slog.Error("Failed to decode request body", "error", err)
		WriteError(w, errors.New("failed to decode request body"))
		return
	}

	taskID := rand.Text()

	h.archiveService.NewTask(&model.Task{
		ID:         taskID,
		Links:      req.Links,
		Downloaded: make(map[string]string),
	})
	slog.Info("New task created", "taskID", taskID, "links", req.Links)

	err = h.archiveService.SendForLoadTasks(taskID)
	if err != nil {
		slog.Error("Failed to send tasks for loading", "taskID", taskID, "error", err)
		WriteError(w, err)
		return
	}

	WriteSuccess(w, map[string]string{"taskID": taskID})
}

func (h *CollectorHandler) AddedLink(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")
	var req struct {
		Link string `json:"link"`
	}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		slog.Error("Failed to decode request body for AddedLink", "error", err)
		WriteError(w, errors.New("failed to decode request body"))
		return
	}

	err = h.archiveService.AddLink(taskID, req.Link)
	if err != nil {
		slog.Error("Failed to add link to task", "taskID", taskID, "link", req.Link, "error", err)
		WriteError(w, err)
		return
	}

	err = h.archiveService.SendForLoadTasks(taskID)
	if err != nil {
		slog.Error("Failed to send tasks for loading after AddedLink", "taskID", taskID, "error", err)
		WriteError(w, err)
		return
	}

	slog.Info("Link added to task", "taskID", taskID, "link", req.Link)
	WriteSuccess(w, map[string]string{"taskID": taskID})
}

func (h *CollectorHandler) HandleGetStatus(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")
	status := h.archiveService.GetStatus(taskID)
	if status == nil {
		slog.Error("Task not found in HandleGetStatus", "taskID", taskID)
		WriteError(w, errors.New("task not found"))
		return
	}

	count := 0
	for _, err := range status {
		if err == "ok" {
			count++
		}
	}
	if count >= config.GetEnvInt("MAX_FILES_PER_TASK", 10) {
		archPath := h.archiveService.ZipArchive(taskID)
		if archPath == "" {
			slog.Error("Failed to create archive in HandleGetStatus", "taskID", taskID)
			WriteError(w, errors.New("failed to create archive"))
			return
		}
		slog.Info("Archive created and sent in HandleGetStatus", "taskID", taskID, "archivePath", archPath)
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", "attachment; filename="+taskID+".zip")
		w.WriteHeader(http.StatusOK)
		http.ServeFile(w, r, archPath)
		return
	}

	slog.Info("Status requested for task", "taskID", taskID, "status", status)
	WriteSuccess(w, status)
}

func (h *CollectorHandler) HandleRemoveTask(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")
	h.archiveService.RemoveTask(taskID)
	slog.Info("Task removed", "taskID", taskID)
	WriteSuccess(w, map[string]string{"taskID": taskID})
}

func (h *CollectorHandler) GetZip(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")
	archPath := h.archiveService.ZipArchive(taskID)
	if archPath == "" {
		slog.Error("Failed to create archive in GetZip", "taskID", taskID)
		WriteError(w, errors.New("failed to create archive"))
		return
	}
	slog.Info("Archive sent in GetZip", "taskID", taskID, "archivePath", archPath)
	archiveUrl := fmt.Sprintf("http://%s/archive/%s.zip", r.Host, taskID)
	WriteSuccess(w, map[string]string{"archiveUrl": archiveUrl})
}

func (h *CollectorHandler) ServeArchive(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")
	archPath := filepath.Join("archive", taskID+".zip")

	if _, err := os.Stat(archPath); err != nil {
		slog.Error("Archive file not found", "taskID", taskID, "archivePath", archPath)
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+taskID+".zip\"")
	http.ServeFile(w, r, archPath)
}
