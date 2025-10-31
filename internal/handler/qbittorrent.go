package handler

import (
	"github.com/carcheky/keepercheky/internal/config"
	"github.com/carcheky/keepercheky/internal/service/clients"
	"github.com/carcheky/keepercheky/pkg/logger"
	"github.com/gofiber/fiber/v2"
)

type QBittorrentHandler struct {
	config *config.Config
	logger *logger.Logger
	client *clients.QBittorrentClient
}

func NewQBittorrentHandler(cfg *config.Config, appLogger *logger.Logger) *QBittorrentHandler {
	var qbClient *clients.QBittorrentClient

	if cfg.Clients.QBittorrent.Enabled {
		zapLogger := appLogger.Desugar()
		qbClient = clients.NewQBittorrentClient(
			cfg.Clients.QBittorrent.URL,
			cfg.Clients.QBittorrent.Username,
			cfg.Clients.QBittorrent.Password,
			zapLogger,
		)
	}

	return &QBittorrentHandler{
		config: cfg,
		logger: appLogger,
		client: qbClient,
	}
}

// GetTransferInfo retrieves global transfer statistics from qBittorrent.
func (h *QBittorrentHandler) GetTransferInfo(c *fiber.Ctx) error {
	if h.client == nil {
		return c.Status(503).JSON(fiber.Map{
			"error": "qBittorrent client not configured",
		})
	}

	info, err := h.client.GetTransferInfo(c.Context())
	if err != nil {
		h.logger.Error("Failed to get qBittorrent transfer info", "error", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to retrieve transfer info: " + err.Error(),
		})
	}

	return c.JSON(info)
}

// GetServerState retrieves server state from qBittorrent.
func (h *QBittorrentHandler) GetServerState(c *fiber.Ctx) error {
	if h.client == nil {
		return c.Status(503).JSON(fiber.Map{
			"error": "qBittorrent client not configured",
		})
	}

	state, err := h.client.GetServerState(c.Context())
	if err != nil {
		h.logger.Error("Failed to get qBittorrent server state", "error", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to retrieve server state: " + err.Error(),
		})
	}

	return c.JSON(state)
}

// GetActiveTorrents retrieves currently active (downloading or seeding) torrents.
func (h *QBittorrentHandler) GetActiveTorrents(c *fiber.Ctx) error {
	if h.client == nil {
		return c.Status(503).JSON(fiber.Map{
			"error": "qBittorrent client not configured",
		})
	}

	torrentsMap, err := h.client.GetAllTorrentsMap(c.Context())
	if err != nil {
		h.logger.Error("Failed to get qBittorrent torrents", "error", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to retrieve torrents: " + err.Error(),
		})
	}

	// Convert map to slice and filter for active torrents
	activeTorrents := make([]interface{}, 0)
	seen := make(map[string]bool)

	for _, torrent := range torrentsMap {
		// Skip duplicates (same torrent indexed by different paths)
		if seen[torrent.Hash] {
			continue
		}
		seen[torrent.Hash] = true

		// Include downloading or seeding torrents with activity
		if torrent.DlSpeed > 0 || torrent.UpSpeed > 0 || torrent.State == clients.TorrentStateDownloading {
			activeTorrents = append(activeTorrents, fiber.Map{
				"hash":         torrent.Hash,
				"name":         torrent.Name,
				"state":        torrent.State,
				"progress":     torrent.Progress,
				"ratio":        torrent.Ratio,
				"size":         torrent.Size,
				"upspeed":      torrent.UpSpeed,
				"dlspeed":      torrent.DlSpeed,
				"eta":          torrent.ETA,
				"num_seeds":    torrent.NumSeeds,
				"num_peers":    torrent.NumPeers,
				"category":     torrent.Category,
				"is_seeding":   torrent.IsSeeding,
				"is_complete":  torrent.IsComplete,
				"added_on":     torrent.AddedOn,
				"completed_on": torrent.CompletedOn,
			})
		}
	}

	return c.JSON(fiber.Map{
		"total": len(activeTorrents),
		"items": activeTorrents,
	})
}

// GetTorrentProperties retrieves detailed properties for a specific torrent.
func (h *QBittorrentHandler) GetTorrentProperties(c *fiber.Ctx) error {
	if h.client == nil {
		return c.Status(503).JSON(fiber.Map{
			"error": "qBittorrent client not configured",
		})
	}

	hash := c.Params("hash")
	if hash == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "torrent hash required",
		})
	}

	props, err := h.client.GetTorrentProperties(c.Context(), hash)
	if err != nil {
		h.logger.Error("Failed to get torrent properties", "error", err, "hash", hash)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to retrieve torrent properties: " + err.Error(),
		})
	}

	return c.JSON(props)
}

// GetTorrentTrackers retrieves tracker information for a specific torrent.
func (h *QBittorrentHandler) GetTorrentTrackers(c *fiber.Ctx) error {
	if h.client == nil {
		return c.Status(503).JSON(fiber.Map{
			"error": "qBittorrent client not configured",
		})
	}

	hash := c.Params("hash")
	if hash == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "torrent hash required",
		})
	}

	trackers, err := h.client.GetTorrentTrackers(c.Context(), hash)
	if err != nil {
		h.logger.Error("Failed to get torrent trackers", "error", err, "hash", hash)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to retrieve torrent trackers: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"total":    len(trackers),
		"trackers": trackers,
	})
}
