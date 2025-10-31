package handler

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/carcheky/keepercheky/internal/service"
	"github.com/carcheky/keepercheky/pkg/logger"
	"github.com/gofiber/fiber/v2"
)

// SyncService interface defines the contract for sync services
type SyncService interface {
	SyncAllWithProgress(ctx context.Context, progressChan chan<- service.SyncProgress) error
}

type SyncHandler struct {
	syncService SyncService
	logger      *logger.Logger
}

func NewSyncHandler(syncService SyncService, logger *logger.Logger) *SyncHandler {
	return &SyncHandler{
		syncService: syncService,
		logger:      logger,
	}
}

// Sync triggers a full synchronization from all services with real-time progress updates.
func (h *SyncHandler) Sync(c *fiber.Ctx) error {
	h.logger.Info("Manual sync triggered")

	// Set headers for Server-Sent Events
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no") // Disable nginx buffering

	// Channel for progress updates
	progressChan := make(chan service.SyncProgress, 100)

	// Create context that won't be cancelled by defer
	// It will be cancelled when the sync finishes or times out
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)

	// Run sync in goroutine
	go func() {
		defer close(progressChan)
		defer cancel() // Cancel context when sync completes

		progressChan <- service.SyncProgress{
			Step:    "init",
			Message: "🔄 Iniciando sincronización completa...",
			Status:  "processing",
		}

		// Execute sync with progress reporting
		if err := h.syncService.SyncAllWithProgress(ctx, progressChan); err != nil {
			progressChan <- service.SyncProgress{
				Step:    "error",
				Message: fmt.Sprintf("❌ Error durante sincronización: %v", err),
				Status:  "error",
			}
			return
		}

		// Completion message is now sent by the service itself
	}()

	// Stream progress updates to client
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		for progress := range progressChan {
			// Marshal to JSON manually for better control
			jsonData, err := json.Marshal(progress)
			if err != nil {
				h.logger.Error("Failed to marshal JSON", "error", err)
				return
			}

			// Write SSE format: data: {json}\n\n
			if _, err := fmt.Fprintf(w, "data: %s\n\n", jsonData); err != nil {
				h.logger.Error("Failed to write SSE", "error", err)
				return
			}

			if err := w.Flush(); err != nil {
				h.logger.Error("Failed to flush", "error", err)
				return
			}

			// Log what we're sending
			h.logger.Info("Sent SSE event",
				"step", progress.Step,
				"message", progress.Message,
				"status", progress.Status,
			)

			// If error or complete, close stream after a small delay
			// to ensure the message is received
			if progress.Status == "error" || progress.Step == "complete" {
				time.Sleep(100 * time.Millisecond)
				return
			}
		}
	})

	return nil
}

// SyncFiles triggers a filesystem-first synchronization with real-time progress updates.
// This is designed for the Files view and shows detailed file discovery and enrichment.
func (h *SyncHandler) SyncFiles(c *fiber.Ctx) error {
	h.logger.Info("Filesystem sync triggered from Files view")

	// Set headers for Server-Sent Events
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no") // Disable nginx buffering

	// Channel for progress updates
	progressChan := make(chan service.SyncProgress, 100)

	// Create context that won't be cancelled by defer
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)

	// Run sync in goroutine
	go func() {
		// DON'T close the channel here - the service will close it
		defer cancel()

		progressChan <- service.SyncProgress{
			Step:    "init",
			Message: "🗂️  Iniciando sincronización de archivos...",
			Status:  "processing",
		}

		// Check if we have a FilesystemSyncService
		fsSync, ok := h.syncService.(*service.FilesystemSyncService)
		if !ok {
			progressChan <- service.SyncProgress{
				Step:    "error",
				Message: "❌ Servicio de filesystem no disponible. Usando sincronización estándar.",
				Status:  "error",
			}
			close(progressChan) // Close only if we're not calling the service
			return
		}

		// Execute filesystem-first sync with progress reporting
		// The service will close the channel when done
		if err := fsSync.SyncAllWithProgress(ctx, progressChan); err != nil {
			// Error already sent by the service
			return
		}
	}()

	// Stream progress updates to client
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		for progress := range progressChan {
			// Marshal to JSON
			jsonData, err := json.Marshal(progress)
			if err != nil {
				h.logger.Error("Failed to marshal JSON", "error", err)
				return
			}

			// Write SSE format
			if _, err := fmt.Fprintf(w, "data: %s\n\n", jsonData); err != nil {
				h.logger.Error("Failed to write SSE", "error", err)
				return
			}

			if err := w.Flush(); err != nil {
				h.logger.Error("Failed to flush", "error", err)
				return
			}

			h.logger.Info("Sent SSE event",
				"step", progress.Step,
				"message", progress.Message,
				"status", progress.Status,
			)

			// If error or complete, close stream after a delay
			if progress.Status == "error" || progress.Step == "complete" {
				time.Sleep(100 * time.Millisecond)
				return
			}
		}
	})

	return nil
}
