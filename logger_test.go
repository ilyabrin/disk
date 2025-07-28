package disk

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"
)

func TestLogger(t *testing.T) {
	t.Run("Logger levels work correctly", func(t *testing.T) {
		var buf bytes.Buffer
		config := &LoggerConfig{
			Level:      WARN,
			Output:     &buf,
			Prefix:     "[test] ",
			Structured: true,
		}

		logger := NewLogger(config)

		// These should not appear in output
		logger.Debug("debug message")
		logger.Info("info message")

		// These should appear
		logger.Warn("warn message")
		logger.Error("error message")

		output := buf.String()
		if strings.Contains(output, "debug message") {
			t.Error("Debug message should not appear with WARN level")
		}
		if strings.Contains(output, "info message") {
			t.Error("Info message should not appear with WARN level")
		}
		if !strings.Contains(output, "warn message") {
			t.Error("Warn message should appear with WARN level")
		}
		if !strings.Contains(output, "error message") {
			t.Error("Error message should appear with WARN level")
		}
	})

	t.Run("Verbose mode enables DEBUG level", func(t *testing.T) {
		var buf bytes.Buffer
		config := &LoggerConfig{
			Level:   INFO,
			Output:  &buf,
			Verbose: true,
		}

		logger := NewLogger(config)
		logger.Debug("debug message")

		output := buf.String()
		if !strings.Contains(output, "debug message") {
			t.Error("Debug message should appear in verbose mode")
		}
	})

	t.Run("Sanitization works correctly", func(t *testing.T) {
		var buf bytes.Buffer
		config := &LoggerConfig{
			Level:        DEBUG,
			Output:       &buf,
			SanitizeAuth: true,
		}

		logger := NewLogger(config)

		// Test various sensitive keys
		sanitized := logger.SanitizeValue("Authorization", "Bearer very-secret-token-here")
		if !strings.Contains(sanitized, "***") {
			t.Error("Authorization header should be sanitized")
		}

		sanitized = logger.SanitizeValue("Content-Type", "application/json")
		if strings.Contains(sanitized, "***") {
			t.Error("Content-Type header should not be sanitized")
		}

		// Test short tokens
		sanitized = logger.SanitizeValue("token", "short")
		if sanitized != "***" {
			t.Error("Short tokens should be completely hidden")
		}
	})

	t.Run("Silent mode logs nothing", func(t *testing.T) {
		var buf bytes.Buffer
		config := &LoggerConfig{
			Level:  SILENT,
			Output: &buf,
		}

		logger := NewLogger(config)
		logger.Error("error message")

		if buf.Len() > 0 {
			t.Error("Silent mode should log nothing")
		}
	})
}

func TestClientLogging(t *testing.T) {
	t.Run("Client with custom logging config", func(t *testing.T) {
		var buf bytes.Buffer
		config := &ClientConfig{
			DefaultTimeout: 30 * time.Second,
			Logger: &LoggerConfig{
				Level:   DEBUG,
				Output:  &buf,
				Verbose: true,
			},
		}

		client, err := NewWithConfig(config, "test-token")
		if err != nil {
			t.Fatal("Failed to create client:", err)
		}

		if client.Logger == nil {
			t.Fatal("Client logger should not be nil")
		}

		// Test log level setting
		client.SetLogLevel(ERROR)
		client.Logger.Info("info message")
		client.Logger.Error("error message")

		output := buf.String()
		if strings.Contains(output, "info message") {
			t.Error("Info message should not appear with ERROR level")
		}
		if !strings.Contains(output, "error message") {
			t.Error("Error message should appear with ERROR level")
		}
	})

	t.Run("Log output can be changed", func(t *testing.T) {
		client, _ := New("test-token")

		var buf bytes.Buffer
		client.SetLogOutput(&buf)

		client.Logger.Info("test message")

		if !strings.Contains(buf.String(), "test message") {
			t.Error("Message should appear in custom output")
		}
	})

	t.Run("Verbose mode can be toggled", func(t *testing.T) {
		var buf bytes.Buffer
		client, _ := New("test-token")
		client.SetLogOutput(&buf)

		client.SetVerbose(true)
		client.Logger.Debug("debug message")

		if !strings.Contains(buf.String(), "debug message") {
			t.Error("Debug message should appear in verbose mode")
		}
	})
}

func TestRequestResponseLogging(t *testing.T) {
	t.Run("Logger sanitizes authorization headers", func(t *testing.T) {
		var buf bytes.Buffer
		config := &LoggerConfig{
			Level:        DEBUG,
			Output:       &buf,
			Verbose:      true,
			SanitizeAuth: true,
		}

		logger := NewLogger(config)

		// Test that logger sanitizes headers correctly
		headers := map[string]string{
			"Authorization": "OAuth very-secret-token-here",
			"Content-Type":  "application/json",
		}

		logger.LogRequest("GET", "https://example.com", headers)

		output := buf.String()
		if !strings.Contains(output, "HTTP Request") {
			t.Error("Should log HTTP request")
		}

		// In verbose mode, headers should be logged
		if strings.Contains(output, "very-secret-token-here") {
			t.Error("Full token should not appear in logs")
		}
		if !strings.Contains(output, "***") {
			t.Error("Should contain sanitized authorization")
		}
	})

	t.Run("Response logging works correctly", func(t *testing.T) {
		var buf bytes.Buffer
		config := &LoggerConfig{
			Level:  DEBUG,
			Output: &buf,
		}

		logger := NewLogger(config)
		logger.LogResponse(200, 1024, 150*time.Millisecond)

		output := buf.String()
		if !strings.Contains(output, "HTTP Response") {
			t.Error("Should log HTTP response")
		}
		if !strings.Contains(output, "200") {
			t.Error("Should include status code")
		}
		if !strings.Contains(output, "1024") {
			t.Error("Should include content length")
		}
	})
}

func TestFileLogging(t *testing.T) {
	t.Run("Can log to file", func(t *testing.T) {
		// Create a temporary file for testing
		tmpFile, err := os.CreateTemp("", "disklog_test_*.log")
		if err != nil {
			t.Fatal("Failed to create temp file:", err)
		}
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		config := &ClientConfig{
			Logger: &LoggerConfig{
				Level:  INFO,
				Output: tmpFile,
			},
		}

		client, _ := NewWithConfig(config, "test-token")
		client.Logger.Info("test file logging")

		// Read file contents
		tmpFile.Seek(0, 0)
		buf := make([]byte, 1024)
		n, _ := tmpFile.Read(buf)
		content := string(buf[:n])

		if !strings.Contains(content, "test file logging") {
			t.Error("Log should be written to file")
		}
	})
}
