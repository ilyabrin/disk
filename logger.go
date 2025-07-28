package disk

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

// LogLevel represents the severity level of a log message
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	SILENT // No logging
)

// String returns the string representation of a LogLevel
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case SILENT:
		return "SILENT"
	default:
		return "UNKNOWN"
	}
}

// LoggerConfig holds configuration for the logger
type LoggerConfig struct {
	Level        LogLevel  // Minimum log level to output
	Output       io.Writer // Where to write logs (default: os.Stdout)
	Prefix       string    // Prefix for log messages
	TimeFormat   string    // Time format for timestamps
	Structured   bool      // Enable structured logging
	Verbose      bool      // Enable verbose mode (includes DEBUG level)
	SanitizeAuth bool      // Sanitize authorization headers in logs
}

// DefaultLoggerConfig returns a LoggerConfig with sensible defaults
func DefaultLoggerConfig() *LoggerConfig {
	return &LoggerConfig{
		Level:        INFO,
		Output:       os.Stdout,
		Prefix:       "[disk] ",
		TimeFormat:   "2006-01-02 15:04:05",
		Structured:   true,
		Verbose:      false,
		SanitizeAuth: true,
	}
}

// DiskLogger provides structured logging with multiple levels
type DiskLogger struct {
	config *LoggerConfig
	logger *log.Logger
}

// NewLogger creates a new DiskLogger with the given configuration
func NewLogger(config *LoggerConfig) *DiskLogger {
	if config == nil {
		config = DefaultLoggerConfig()
	}

	// Set DEBUG level if verbose mode is enabled
	if config.Verbose && config.Level > DEBUG {
		config.Level = DEBUG
	}

	return &DiskLogger{
		config: config,
		logger: log.New(config.Output, config.Prefix, 0), // We'll handle timestamps ourselves
	}
}

// shouldLog checks if a message at the given level should be logged
func (l *DiskLogger) shouldLog(level LogLevel) bool {
	return level >= l.config.Level && l.config.Level != SILENT
}

// formatMessage formats a log message with timestamp and level
func (l *DiskLogger) formatMessage(level LogLevel, format string, args ...interface{}) string {
	timestamp := time.Now().Format(l.config.TimeFormat)
	message := fmt.Sprintf(format, args...)

	if l.config.Structured {
		return fmt.Sprintf("[%s] %s: %s", timestamp, level.String(), message)
	}
	return fmt.Sprintf("[%s] %s", timestamp, message)
}

// Debug logs a debug message
func (l *DiskLogger) Debug(format string, args ...interface{}) {
	if l.shouldLog(DEBUG) {
		l.logger.Print(l.formatMessage(DEBUG, format, args...))
	}
}

// Info logs an info message
func (l *DiskLogger) Info(format string, args ...interface{}) {
	if l.shouldLog(INFO) {
		l.logger.Print(l.formatMessage(INFO, format, args...))
	}
}

// Warn logs a warning message
func (l *DiskLogger) Warn(format string, args ...interface{}) {
	if l.shouldLog(WARN) {
		l.logger.Print(l.formatMessage(WARN, format, args...))
	}
}

// Error logs an error message
func (l *DiskLogger) Error(format string, args ...interface{}) {
	if l.shouldLog(ERROR) {
		l.logger.Print(l.formatMessage(ERROR, format, args...))
	}
}

// SetLevel updates the minimum log level
func (l *DiskLogger) SetLevel(level LogLevel) {
	l.config.Level = level
}

// SetVerbose enables or disables verbose mode
func (l *DiskLogger) SetVerbose(verbose bool) {
	l.config.Verbose = verbose
	if verbose && l.config.Level > DEBUG {
		l.config.Level = DEBUG
	}
}

// SetOutput changes the output destination for logs
func (l *DiskLogger) SetOutput(output io.Writer) {
	l.config.Output = output
	l.logger.SetOutput(output)
}

// SanitizeValue sanitizes sensitive information for logging
func (l *DiskLogger) SanitizeValue(key, value string) string {
	if !l.config.SanitizeAuth {
		return value
	}

	lowerKey := strings.ToLower(key)
	if strings.Contains(lowerKey, "auth") ||
		strings.Contains(lowerKey, "token") ||
		strings.Contains(lowerKey, "key") ||
		strings.Contains(lowerKey, "secret") {
		if len(value) <= 8 {
			return "***"
		}
		return value[:4] + "***" + value[len(value)-2:]
	}
	return value
}

// LogRequest logs HTTP request details
func (l *DiskLogger) LogRequest(method, url string, headers map[string]string) {
	if !l.shouldLog(DEBUG) {
		return
	}

	l.Debug("HTTP Request: %s %s", method, url)

	if l.config.Verbose {
		for key, value := range headers {
			sanitizedValue := l.SanitizeValue(key, value)
			if sanitizedValue != value {
				l.Debug("  Header: %s: %s", key, sanitizedValue)
			} else {
				l.Debug("  Header: %s: [sanitized]", key)
			}
		}
	}
}

// LogResponse logs HTTP response details
func (l *DiskLogger) LogResponse(statusCode int, contentLength int64, duration time.Duration) {
	if l.shouldLog(DEBUG) {
		l.Debug("HTTP Response: %d (Content-Length: %d, Duration: %v)",
			statusCode, contentLength, duration)
	} else if l.shouldLog(INFO) && statusCode >= 400 {
		l.Info("HTTP Error Response: %d (Duration: %v)", statusCode, duration)
	}
}

// LogError logs an error with context
func (l *DiskLogger) LogError(operation string, err error) {
	if l.shouldLog(ERROR) {
		l.Error("Operation '%s' failed: %v", operation, err)
	}
}
