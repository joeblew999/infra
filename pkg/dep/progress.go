package dep

import (
	"fmt"
	"io"
	"strings"
	"time"
)

// ProgressWriter wraps an io.Writer to show download progress
type ProgressWriter struct {
	Total       int64
	Downloaded  int64
	Name        string
	StartTime   time.Time
	LastUpdate  time.Time
}

// NewProgressWriter creates a new progress writer
func NewProgressWriter(total int64, name string) *ProgressWriter {
	return &ProgressWriter{
		Total:     total,
		Name:      name,
		StartTime: time.Now(),
	}
}

// Write implements io.Writer and updates progress
func (pw *ProgressWriter) Write(p []byte) (n int, err error) {
	n = len(p)
	pw.Downloaded += int64(n)
	
	// Update progress every 100ms to avoid too frequent updates
	now := time.Now()
	if now.Sub(pw.LastUpdate) > 100*time.Millisecond || pw.Downloaded == pw.Total {
		pw.ShowProgress()
		pw.LastUpdate = now
	}
	
	return n, nil
}

// ShowProgress displays the current progress
func (pw *ProgressWriter) ShowProgress() {
	if pw.Total <= 0 {
		// Unknown size, just show downloaded amount
		fmt.Printf("\r⬇️  %s: %s downloaded", pw.Name, formatBytes(pw.Downloaded))
		return
	}
	
	percent := float64(pw.Downloaded) / float64(pw.Total) * 100
	elapsed := time.Since(pw.StartTime)
	
	// Calculate speed
	speed := float64(pw.Downloaded) / elapsed.Seconds()
	
	// Estimate remaining time
	var eta string
	if speed > 0 && pw.Downloaded < pw.Total {
		remaining := time.Duration(float64(pw.Total-pw.Downloaded) / speed * float64(time.Second))
		eta = fmt.Sprintf(" ETA: %s", formatDuration(remaining))
	}
	
	// Create progress bar
	barWidth := 30
	filledWidth := int(float64(barWidth) * percent / 100)
	bar := strings.Repeat("█", filledWidth) + strings.Repeat("░", barWidth-filledWidth)
	
	fmt.Printf("\r⬇️  %s: [%s] %.1f%% (%s/%s) %s/s%s",
		pw.Name,
		bar,
		percent,
		formatBytes(pw.Downloaded),
		formatBytes(pw.Total),
		formatBytes(int64(speed)),
		eta,
	)
}

// Finish completes the progress display
func (pw *ProgressWriter) Finish() {
	pw.ShowProgress()
	elapsed := time.Since(pw.StartTime)
	avgSpeed := float64(pw.Downloaded) / elapsed.Seconds()
	fmt.Printf("\n✅ %s downloaded (%s in %s, avg %s/s)\n",
		pw.Name,
		formatBytes(pw.Downloaded),
		formatDuration(elapsed),
		formatBytes(int64(avgSpeed)),
	)
}

// ProgressReader wraps an io.Reader to track progress
type ProgressReader struct {
	io.Reader
	progress *ProgressWriter
}

// NewProgressReader creates a progress-tracking reader
func NewProgressReader(reader io.Reader, total int64, name string) *ProgressReader {
	return &ProgressReader{
		Reader:   reader,
		progress: NewProgressWriter(total, name),
	}
}

// Read implements io.Reader and tracks progress
func (pr *ProgressReader) Read(p []byte) (n int, err error) {
	n, err = pr.Reader.Read(p)
	if n > 0 {
		pr.progress.Write(p[:n])
	}
	return n, err
}

// Finish completes the progress display
func (pr *ProgressReader) Finish() {
	pr.progress.Finish()
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.0fm%.0fs", d.Minutes(), d.Seconds()-60*d.Minutes())
	}
	return fmt.Sprintf("%.0fh%.0fm", d.Hours(), d.Minutes()-60*d.Hours())
}