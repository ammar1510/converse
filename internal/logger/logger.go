package logger

import (
	"fmt"
	"log"
	"os"
)

// Log levels
const (
	LevelDebug = iota
	LevelInfo
	LevelWarn
	LevelError
)

var (
	levelNames = map[int]string{
		LevelDebug: "DEBUG",
		LevelInfo:  "INFO",
		LevelWarn:  "WARN",
		LevelError: "ERROR",
	}

	// Default to INFO in production, DEBUG in development
	minLevel = LevelInfo
)

// Logger wraps the standard logger with levels
type Logger struct {
	component string
}

func init() {
	// Set log level based on environment
	if os.Getenv("ENV") == "development" {
		minLevel = LevelDebug
	}

	// Configure standard logger format
	log.SetFlags(log.Ldate | log.Ltime)
}

// New creates a new logger for a specific component
func New(component string) *Logger {
	return &Logger{component: component}
}

// SetMinLevel allows changing the minimum log level at runtime
func SetMinLevel(level int) {
	minLevel = level
}

// logf logs a message at the specified level
func (l *Logger) logf(level int, format string, args ...interface{}) {
	if level < minLevel {
		return
	}

	prefix := fmt.Sprintf("[%s][%s] ", levelNames[level], l.component)
	log.Printf(prefix+format, args...)
}

// Debug logs debug information
func (l *Logger) Debug(format string, args ...interface{}) {
	l.logf(LevelDebug, format, args...)
}

// Info logs information messages
func (l *Logger) Info(format string, args ...interface{}) {
	l.logf(LevelInfo, format, args...)
}

// Warn logs warning messages
func (l *Logger) Warn(format string, args ...interface{}) {
	l.logf(LevelWarn, format, args...)
}

// Error logs error messages
func (l *Logger) Error(format string, args ...interface{}) {
	l.logf(LevelError, format, args...)
}

// GetAppEnv returns the current application environment
func GetAppEnv() string {
	env := os.Getenv("ENV")
	if env == "" {
		return "development" // Default to development
	}
	return env
}

// IsDevelopment returns true if the current environment is development
func IsDevelopment() bool {
	return GetAppEnv() == "development"
}
