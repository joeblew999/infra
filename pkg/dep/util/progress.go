package util

import (
	"fmt"
	"io"
	"strings"
	"time"
)

// ProgressWriter shows download progress with a visual progress bar
type ProgressWriter struct {
	total       int64
	downloaded  int64
	name        string
	startTime   time.Time
	lastUpdate  time.Time
	updateEvery int64 // Update every N bytes
}

// NewProgressWriter creates a new progress writer
func NewProgressWriter(total int64, name string) *ProgressWriter {
	updateEvery := int64(1048576) // Default: 1MB
	if total > 100*1048576 {      // For files > 100MB, update every 5MB
		updateEvery = 5 * 1048576
	}
	
	return &ProgressWriter{
		total:       total,
		name:        name,
		startTime:   time.Now(),
		updateEvery: updateEvery,
	}
}

// Write implements io.Writer and updates progress
func (pw *ProgressWriter) Write(p []byte) (n int, err error) {
	n = len(p)
	pw.downloaded += int64(n)
	
	// Update progress at intervals or at completion
	now := time.Now()
	shouldUpdate := pw.downloaded%pw.updateEvery == 0 || 
		pw.downloaded == pw.total ||
		now.Sub(pw.lastUpdate) > 500*time.Millisecond
	
	if shouldUpdate {
		pw.ShowProgress()
		pw.lastUpdate = now
	}
	
	return n, nil
}

// ShowProgress displays the current progress with a visual bar
func (pw *ProgressWriter) ShowProgress() {
	if pw.total <= 0 {
		// Unknown size, just show downloaded amount
		fmt.Printf("\r⬇️  %s: %s downloaded", pw.name, FormatBytes(pw.downloaded))
		return
	}
	
	percent := float64(pw.downloaded) / float64(pw.total) * 100
	elapsed := time.Since(pw.startTime)
	
	// Calculate speed
	speed := float64(pw.downloaded) / elapsed.Seconds()
	
	// Estimate remaining time
	var eta string
	if speed > 0 && pw.downloaded < pw.total {
		remaining := time.Duration(float64(pw.total-pw.downloaded) / speed * float64(time.Second))
		eta = fmt.Sprintf(" ETA: %s", FormatDuration(remaining))
	}
	
	// Create progress bar (adaptive width)
	barWidth := 25
	filledWidth := int(float64(barWidth) * percent / 100)
	bar := strings.Repeat("█", filledWidth) + strings.Repeat("░", barWidth-filledWidth)
	
	fmt.Printf("\r⬇️  %s: [%s] %.1f%% (%s/%s) %s/s%s",
		pw.name,
		bar,
		percent,
		FormatBytes(pw.downloaded),
		FormatBytes(pw.total),
		FormatBytes(int64(speed)),
		eta,
	)
}

// Finish completes the progress display
func (pw *ProgressWriter) Finish() {
	pw.ShowProgress()
	elapsed := time.Since(pw.startTime)
	avgSpeed := float64(pw.downloaded) / elapsed.Seconds()
	fmt.Printf("\n✅ %s downloaded (%s in %s, avg %s/s)\n",
		pw.name,
		FormatBytes(pw.downloaded),
		FormatDuration(elapsed),
		FormatBytes(int64(avgSpeed)),
	)
}

// ProgressReader wraps an io.Reader to track download progress
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

// SimpleProgressWriter provides basic progress updates without visual bars
type SimpleProgressWriter struct {
	downloaded int64
	name       string
	updateEvery int64
}

// NewSimpleProgressWriter creates a simple progress writer
func NewSimpleProgressWriter(name string) *SimpleProgressWriter {
	return &SimpleProgressWriter{
		name:        name,
		updateEvery: 1048576, // Update every 1MB
	}
}

// Write implements io.Writer with simple progress updates
func (spw *SimpleProgressWriter) Write(p []byte) (n int, err error) {
	n = len(p)
	spw.downloaded += int64(n)
	
	if spw.downloaded%spw.updateEvery == 0 {
		fmt.Printf("\r⬇️  %s: %s downloaded...", spw.name, FormatBytes(spw.downloaded))
	}
	
	return n, nil
}

// Finish completes the simple progress display
func (spw *SimpleProgressWriter) Finish() {
	fmt.Printf("\r✅ %s downloaded (%s total)\n", spw.name, FormatBytes(spw.downloaded))
}