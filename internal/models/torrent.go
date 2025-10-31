package models

// TorrentInfo represents torrent information for media
type TorrentInfo struct {
	Hash            string  `json:"hash"`
	Name            string  `json:"name"`
	Size            int64   `json:"size"`
	Progress        float64 `json:"progress"`
	State           string  `json:"state"`
	Ratio           float64 `json:"ratio"`
	UpSpeed         int64   `json:"upspeed"`
	DlSpeed         int64   `json:"dlspeed"`
	SeedingTime     int64   `json:"seeding_time"`
	Category        string  `json:"category"`
	Tags            string  `json:"tags"`
	SavePath        string  `json:"save_path"`
	Seeders         int     `json:"seeders"`
	Leechers        int     `json:"leechers"`
	IsSeeding       bool    `json:"is_seeding"`
	IsComplete      bool    `json:"is_complete"`
	AddedOn         int64   `json:"added_on,omitempty"`         // Timestamp when torrent was added
	CompletedOn     int64   `json:"completed_on,omitempty"`     // Timestamp when torrent completed
	ETA             int64   `json:"eta,omitempty"`              // Estimated time to completion (seconds)
	TotalUploaded   int64   `json:"total_uploaded,omitempty"`   // Total bytes uploaded
	TotalDownloaded int64   `json:"total_downloaded,omitempty"` // Total bytes downloaded
	NumSeeds        int     `json:"num_seeds,omitempty"`        // Number of seeds connected
	NumPeers        int     `json:"num_peers,omitempty"`        // Number of peers connected
}

// QBittorrentTransferInfo represents global transfer statistics
type QBittorrentTransferInfo struct {
	DLInfoSpeed      int64  `json:"dl_info_speed"`     // Global download rate (bytes/s)
	DLInfoData       int64  `json:"dl_info_data"`      // Data downloaded this session (bytes)
	UPInfoSpeed      int64  `json:"up_info_speed"`     // Global upload rate (bytes/s)
	UPInfoData       int64  `json:"up_info_data"`      // Data uploaded this session (bytes)
	DHTNodes         int    `json:"dht_nodes"`         // DHT nodes connected to
	ConnectionStatus string `json:"connection_status"` // Connection status (connected, firewalled, disconnected)
}

// QBittorrentServerState represents server state from sync/maindata
type QBittorrentServerState struct {
	DLInfoSpeed      int64  `json:"dl_info_speed"`
	UPInfoSpeed      int64  `json:"up_info_speed"`
	DLInfoData       int64  `json:"dl_info_data"`
	UPInfoData       int64  `json:"up_info_data"`
	DHTNodes         int    `json:"dht_nodes"`
	ConnectionStatus string `json:"connection_status"`
	FreeSpaceOnDisk  int64  `json:"free_space_on_disk"`
}

// QBittorrentTorrentProperties represents detailed torrent properties
type QBittorrentTorrentProperties struct {
	AdditionDate    int64  `json:"addition_date"`
	CompletionDate  int64  `json:"completion_date"`
	TotalUploaded   int64  `json:"total_uploaded"`
	TotalDownloaded int64  `json:"total_downloaded"`
	NbConnections   int    `json:"nb_connections"`
	Seeds           int    `json:"seeds"`
	SeedsTotal      int    `json:"seeds_total"`
	Peers           int    `json:"peers"`
	PeersTotal      int    `json:"peers_total"`
	ETA             int64  `json:"eta"`
	SavePath        string `json:"save_path"`
	CreationDate    int64  `json:"creation_date"`
	PieceSize       int64  `json:"piece_size"`
	Comment         string `json:"comment"`
	TotalWasted     int64  `json:"total_wasted"`
	TotalSize       int64  `json:"total_size"`
	UpLimit         int64  `json:"up_limit"`
	DlLimit         int64  `json:"dl_limit"`
}

// QBittorrentTracker represents a tracker for a torrent
type QBittorrentTracker struct {
	URL      string `json:"url"`
	Status   int    `json:"status"` // 0=disabled, 1=not contacted, 2=working, 3=updating, 4=not working
	NumPeers int    `json:"num_peers"`
	NumSeeds int    `json:"num_seeds"`
	Msg      string `json:"msg"` // Tracker message
}
