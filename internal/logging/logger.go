package logging

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

// Entry represents a single request/response pair log entry
type Entry struct {
	Timestamp     time.Time
	RequestBody   interface{}
	Response      interface{}
	Duration      time.Duration
	Path          string
	Method        string
	CorrelationID string // Azure APIM correlation ID for linking with diagnostic logs
}

// Logger interface defines logging behavior
type Logger interface {
	LogRequest(entry Entry)
	Close()
}

// FileLogger implements logging to a file
type FileLogger struct {
	file *os.File
}

// NewFileLogger creates a new file logger
func NewFileLogger(filename string) (*FileLogger, error) {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	return &FileLogger{file: file}, nil
}

// LogRequest logs a request and response to the file
func (l *FileLogger) LogRequest(entry Entry) {
	enc := json.NewEncoder(l.file)
	enc.SetEscapeHTML(false) // prevents HTML escaping for cleaner logs
	if err := enc.Encode(entry); err != nil {
		log.Printf("Error encoding log entry: %v", err)
	} else {
		// Log a brief confirmation to stdout
		log.Printf("Logged %s %s - %v", entry.Method, entry.Path, entry.Duration)
	}
}

// Close closes the log file
func (l *FileLogger) Close() {
	if l.file != nil {
		if err := l.file.Close(); err != nil {
			log.Printf("Error closing log file: %v", err)
		}
	}
}
