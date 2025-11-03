package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

// LogLevel represents the logging level
type LogLevel string

const (
	ErrorLevel LogLevel = "error"
	WarnLevel  LogLevel = "warn"
	InfoLevel  LogLevel = "info"
	DebugLevel LogLevel = "debug"
	TraceLevel LogLevel = "trace"
)

// Initialize sets up the logger with the specified level and file
func Initialize(level string, logFile string) error {
	log = logrus.New()

	// Set log level
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	}
	log.SetLevel(logLevel)

	// Create logs directory if it doesn't exist
	logDir := filepath.Dir(logFile)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open log file
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Create a multi-writer to write to both console and file
	multiWriter := io.MultiWriter(os.Stdout, file)
	log.SetOutput(multiWriter)

	// Set custom formatter with timestamps and colors for console
	log.SetFormatter(&CustomFormatter{
		TimestampFormat: "2006-01-02 15:04:05.000",
		ForceColors:     true,
	})

	return nil
}

// CustomFormatter is a custom logrus formatter with color support
type CustomFormatter struct {
	TimestampFormat string
	ForceColors     bool
}

// Format renders a single log entry
func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := entry.Time.Format(f.TimestampFormat)
	level := strings.ToUpper(entry.Level.String())

	// Color codes for different log levels
	var levelColor string
	if f.ForceColors {
		switch entry.Level {
		case logrus.TraceLevel:
			levelColor = "\033[37m" // Gray
		case logrus.DebugLevel:
			levelColor = "\033[36m" // Cyan
		case logrus.InfoLevel:
			levelColor = "\033[32m" // Green
		case logrus.WarnLevel:
			levelColor = "\033[33m" // Yellow
		case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
			levelColor = "\033[31m" // Red
		}
	}

	resetColor := "\033[0m"
	if !f.ForceColors {
		levelColor = ""
		resetColor = ""
	}

	// Format: timestamp LEVEL [component] message
	component := entry.Data["component"]
	if component == nil {
		component = "Main"
	}

	msg := fmt.Sprintf("%s %s%-5s%s [%s] %s",
		timestamp,
		levelColor,
		level,
		resetColor,
		component,
		entry.Message,
	)

	// Add additional fields
	for key, value := range entry.Data {
		if key != "component" {
			msg += fmt.Sprintf(", %s=%v", key, value)
		}
	}

	return []byte(msg + "\n"), nil
}

// WithComponent returns a logger with a component field
func WithComponent(component string) *logrus.Entry {
	if log == nil {
		// Initialize with defaults if not already initialized
		_ = Initialize("info", "./logs/ldap-test.log")
	}
	return log.WithField("component", component)
}

// Error logs an error message
func Error(component string, message string, fields ...interface{}) {
	entry := WithComponent(component)
	if len(fields) > 0 {
		entry = entry.WithFields(parseFields(fields...))
	}
	entry.Error(message)
}

// Warn logs a warning message
func Warn(component string, message string, fields ...interface{}) {
	entry := WithComponent(component)
	if len(fields) > 0 {
		entry = entry.WithFields(parseFields(fields...))
	}
	entry.Warn(message)
}

// Info logs an info message
func Info(component string, message string, fields ...interface{}) {
	entry := WithComponent(component)
	if len(fields) > 0 {
		entry = entry.WithFields(parseFields(fields...))
	}
	entry.Info(message)
}

// Debug logs a debug message
func Debug(component string, message string, fields ...interface{}) {
	entry := WithComponent(component)
	if len(fields) > 0 {
		entry = entry.WithFields(parseFields(fields...))
	}
	entry.Debug(message)
}

// Trace logs a trace message
func Trace(component string, message string, fields ...interface{}) {
	entry := WithComponent(component)
	if len(fields) > 0 {
		entry = entry.WithFields(parseFields(fields...))
	}
	entry.Trace(message)
}

// parseFields converts variadic interface{} to logrus.Fields
// Expected format: key1, value1, key2, value2, ...
func parseFields(fields ...interface{}) logrus.Fields {
	result := logrus.Fields{}
	for i := 0; i < len(fields)-1; i += 2 {
		key, ok := fields[i].(string)
		if !ok {
			continue
		}
		result[key] = fields[i+1]
	}
	return result
}

// LogLDAPOperation logs detailed LDAP operation information at TRACE level
func LogLDAPOperation(component, operation, dn string, attributes map[string][]string) {
	entry := WithComponent(component)
	entry = entry.WithFields(logrus.Fields{
		"operation":  operation,
		"dn":         dn,
		"attributes": attributes,
	})
	entry.Trace(fmt.Sprintf("Operation: %s", operation))
}

// LogLDAPResult logs the result of an LDAP operation
func LogLDAPResult(component, operation string, success bool, code int, message string, duration time.Duration) {
	entry := WithComponent(component)
	entry = entry.WithFields(logrus.Fields{
		"operation": operation,
		"success":   success,
		"code":      code,
		"message":   message,
		"duration":  fmt.Sprintf("%dms", duration.Milliseconds()),
	})

	if success {
		entry.Trace(fmt.Sprintf("Response: Success (code: %d), Duration: %dms", code, duration.Milliseconds()))
	} else {
		entry.Error(fmt.Sprintf("Response: Failed (code: %d), Message: %s, Duration: %dms", code, message, duration.Milliseconds()))
	}
}

// LogSearchOperation logs a search operation with filter details
func LogSearchOperation(component, baseDN, filter, scope string, attributes []string) {
	entry := WithComponent(component)
	entry = entry.WithFields(logrus.Fields{
		"base_dn":    baseDN,
		"filter":     filter,
		"scope":      scope,
		"attributes": attributes,
	})
	entry.Trace("Operation: Search")
}

// LogSearchResult logs the result of a search operation
func LogSearchResult(component string, entriesFound int, duration time.Duration) {
	entry := WithComponent(component)
	entry = entry.WithFields(logrus.Fields{
		"entries_found": entriesFound,
		"duration":      fmt.Sprintf("%dms", duration.Milliseconds()),
	})
	entry.Trace(fmt.Sprintf("Found %d entries, Duration: %dms", entriesFound, duration.Milliseconds()))
}
