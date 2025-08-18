package collection

import (
	"context"
	"time"
)

// Collector handles collecting binaries from original sources for all platforms
type Collector interface {
	// CollectBinary downloads a binary for all configured platforms
	CollectBinary(ctx context.Context, name, version string) (*CollectionResult, error)
	
	// CollectAll downloads all configured binaries for all platforms
	CollectAll(ctx context.Context) (*BatchCollectionResult, error)
	
	// GetCollectionStatus returns the current collection status
	GetCollectionStatus(name, version string) (*CollectionStatus, error)
	
	// ListCollected returns all collected binaries
	ListCollected() ([]CollectedBinary, error)
}

// Publisher handles uploading collected binaries to managed releases
type Publisher interface {
	// PublishBinary uploads a collected binary to managed release
	PublishBinary(ctx context.Context, name, version string) (*PublishResult, error)
	
	// PublishAll uploads all collected binaries to managed releases
	PublishAll(ctx context.Context) (*BatchPublishResult, error)
	
	// GetPublishStatus returns the publish status of a binary
	GetPublishStatus(name, version string) (*PublishStatus, error)
	
	// ListPublished returns all published binaries
	ListPublished() ([]PublishedBinary, error)
}

// ManagedDownloader handles downloading from managed releases with fallback
type ManagedDownloader interface {
	// Download attempts to download from managed release, falls back to original
	Download(ctx context.Context, name, version, destPath string) (*DownloadResult, error)
	
	// GetDownloadStrategy returns the preferred download strategy
	GetDownloadStrategy(name, version string) (*DownloadStrategy, error)
	
	// CheckAvailability checks if binary is available in managed releases
	CheckAvailability(name, version string) (*AvailabilityStatus, error)
}

// CollectionResult represents the result of collecting a binary
type CollectionResult struct {
	Binary      string                      `json:"binary"`
	Version     string                      `json:"version"`
	Platforms   map[string]*PlatformResult `json:"platforms"`
	Manifest    *BinaryManifest            `json:"manifest"`
	CollectedAt time.Time                  `json:"collected_at"`
	Success     bool                       `json:"success"`
	Errors      []string                   `json:"errors,omitempty"`
}

// PlatformResult represents collection result for a single platform
type PlatformResult struct {
	Platform   string    `json:"platform"`
	Filename   string    `json:"filename"`
	Size       int64     `json:"size"`
	SHA256     string    `json:"sha256"`
	LocalPath  string    `json:"local_path"`
	SourceURL  string    `json:"source_url"`
	Success    bool      `json:"success"`
	Error      string    `json:"error,omitempty"`
	Duration   time.Duration `json:"duration"`
}

// BatchCollectionResult represents results from collecting multiple binaries
type BatchCollectionResult struct {
	Results     map[string]*CollectionResult `json:"results"`
	TotalCount  int                         `json:"total_count"`
	SuccessCount int                        `json:"success_count"`
	FailureCount int                        `json:"failure_count"`
	Duration    time.Duration               `json:"duration"`
	StartedAt   time.Time                  `json:"started_at"`
	CompletedAt time.Time                  `json:"completed_at"`
}

// CollectionStatus represents the current status of a collected binary
type CollectionStatus struct {
	Binary        string             `json:"binary"`
	Version       string             `json:"version"`
	Collected     bool              `json:"collected"`
	PlatformCount int               `json:"platform_count"`
	Platforms     map[string]bool   `json:"platforms"`
	CollectedAt   *time.Time        `json:"collected_at,omitempty"`
	ManifestPath  string            `json:"manifest_path,omitempty"`
}

// CollectedBinary represents a collected binary with metadata
type CollectedBinary struct {
	Name        string      `json:"name"`
	Version     string      `json:"version"`
	Platforms   []string    `json:"platforms"`
	Size        int64       `json:"total_size"`
	CollectedAt time.Time   `json:"collected_at"`
	ManifestPath string     `json:"manifest_path"`
}

// PublishResult represents the result of publishing a binary
type PublishResult struct {
	Binary       string                       `json:"binary"`
	Version      string                       `json:"version"`
	ReleaseTag   string                       `json:"release_tag"`
	ReleaseURL   string                       `json:"release_url"`
	Assets       map[string]*AssetUploadResult `json:"assets"`
	ManifestURL  string                       `json:"manifest_url"`
	Success      bool                         `json:"success"`
	PublishedAt  time.Time                    `json:"published_at"`
	Errors       []string                     `json:"errors,omitempty"`
}

// AssetUploadResult represents the result of uploading a single asset
type AssetUploadResult struct {
	Platform    string        `json:"platform"`
	Filename    string        `json:"filename"`
	Size        int64         `json:"size"`
	UploadURL   string        `json:"upload_url"`
	Success     bool          `json:"success"`
	Error       string        `json:"error,omitempty"`
	Duration    time.Duration `json:"duration"`
}

// BatchPublishResult represents results from publishing multiple binaries
type BatchPublishResult struct {
	Results      map[string]*PublishResult `json:"results"`
	TotalCount   int                      `json:"total_count"`
	SuccessCount int                      `json:"success_count"`
	FailureCount int                      `json:"failure_count"`
	Duration     time.Duration            `json:"duration"`
	StartedAt    time.Time               `json:"started_at"`
	CompletedAt  time.Time               `json:"completed_at"`
}

// PublishStatus represents the current publish status of a binary
type PublishStatus struct {
	Binary      string     `json:"binary"`
	Version     string     `json:"version"`
	Published   bool       `json:"published"`
	ReleaseTag  string     `json:"release_tag,omitempty"`
	ReleaseURL  string     `json:"release_url,omitempty"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	AssetCount  int        `json:"asset_count"`
}

// PublishedBinary represents a published binary with metadata
type PublishedBinary struct {
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	ReleaseTag  string    `json:"release_tag"`
	ReleaseURL  string    `json:"release_url"`
	AssetCount  int       `json:"asset_count"`
	PublishedAt time.Time `json:"published_at"`
}

// DownloadResult represents the result of downloading a binary
type DownloadResult struct {
	Binary     string             `json:"binary"`
	Version    string             `json:"version"`
	Platform   string             `json:"platform"`
	Source     DownloadSource     `json:"source"`
	LocalPath  string             `json:"local_path"`
	Size       int64              `json:"size"`
	SHA256     string             `json:"sha256"`
	Success    bool               `json:"success"`
	Error      string             `json:"error,omitempty"`
	Duration   time.Duration      `json:"duration"`
	Fallbacks  []DownloadAttempt  `json:"fallbacks,omitempty"`
}

// DownloadSource indicates where the binary was downloaded from
type DownloadSource string

const (
	SourceManaged  DownloadSource = "managed"
	SourceOriginal DownloadSource = "original"
	SourceCache    DownloadSource = "cache"
)

// DownloadAttempt represents a single download attempt
type DownloadAttempt struct {
	Source   DownloadSource `json:"source"`
	URL      string         `json:"url"`
	Success  bool           `json:"success"`
	Error    string         `json:"error,omitempty"`
	Duration time.Duration  `json:"duration"`
}

// DownloadStrategy represents the preferred download strategy
type DownloadStrategy struct {
	Binary         string           `json:"binary"`
	Version        string           `json:"version"`
	PreferredSource DownloadSource  `json:"preferred_source"`
	FallbackChain  []DownloadSource `json:"fallback_chain"`
	ManagedAvailable bool           `json:"managed_available"`
	OriginalAvailable bool          `json:"original_available"`
	CacheAvailable   bool           `json:"cache_available"`
}

// AvailabilityStatus represents availability across different sources
type AvailabilityStatus struct {
	Binary    string `json:"binary"`
	Version   string `json:"version"`
	Platform  string `json:"platform"`
	Managed   *SourceAvailability `json:"managed"`
	Original  *SourceAvailability `json:"original"`
	Cache     *SourceAvailability `json:"cache"`
}

// SourceAvailability represents availability from a specific source
type SourceAvailability struct {
	Available bool      `json:"available"`
	URL       string    `json:"url,omitempty"`
	Size      int64     `json:"size,omitempty"`
	SHA256    string    `json:"sha256,omitempty"`
	CheckedAt time.Time `json:"checked_at"`
	Error     string    `json:"error,omitempty"`
}

// BinaryManifest represents the manifest for a collected binary
type BinaryManifest struct {
	Binary         string                    `json:"binary"`
	Version        string                    `json:"version"`
	CollectionDate time.Time                 `json:"collection_date"`
	Source         *SourceInfo               `json:"source"`
	Platforms      map[string]*PlatformInfo  `json:"platforms"`
	ManagedRelease *ManagedReleaseInfo       `json:"managed_release,omitempty"`
}

// SourceInfo represents information about the original source
type SourceInfo struct {
	Type       string `json:"type"` // "github-release", "claude-release", etc.
	Repo       string `json:"repo,omitempty"`
	ReleaseURL string `json:"release_url,omitempty"`
	Version    string `json:"version"`
}

// PlatformInfo represents information about a platform-specific binary
type PlatformInfo struct {
	Filename   string `json:"filename"`
	Size       int64  `json:"size"`
	SHA256     string `json:"sha256"`
	Executable bool   `json:"executable"`
	LocalPath  string `json:"local_path,omitempty"`
}

// ManagedReleaseInfo represents information about the managed release
type ManagedReleaseInfo struct {
	Repo      string    `json:"repo"`
	Tag       string    `json:"tag"`
	Published bool      `json:"published"`
	URL       string    `json:"url,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
}