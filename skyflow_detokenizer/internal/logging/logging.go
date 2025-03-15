// internal/logging/logging.go
package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

// LogLevel represents the level of logging
type LogLevel int

const (
	// LevelInfo includes errors and informational messages
	LevelInfo LogLevel = iota
	// LevelDebug includes all messages (info, errors, and debug)
	LevelDebug
)

var (
	// Default logger settings
	level      = LevelInfo
	logger     *log.Logger
	logFile    *os.File
	loggerInit = false
)

// InitLogger initializes the logger with the specified configuration
func InitLogger(logLevel string, logToFile bool, logFilePath string) error {
	// Set log level
	switch logLevel {
	case "debug":
		level = LevelDebug
	case "info":
		level = LevelInfo
	default:
		level = LevelInfo
	}

	// Set log output destination
	var output io.Writer
	if logToFile && logFilePath != "" {
		// Create directory if it doesn't exist
		dir := filepath.Dir(logFilePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %v", err)
		}

		// Open log file
		var err error
		logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %v", err)
		}
		output = logFile
	} else {
		output = os.Stdout
	}

	// Create logger
	logger = log.New(output, "", log.LstdFlags)
	loggerInit = true
	
	Info("Logging initialized with level: %s, output: %s", logLevel, getOutputDescription(logToFile, logFilePath))
	return nil
}

// getOutputDescription returns a description of the log output destination
func getOutputDescription(logToFile bool, logFilePath string) string {
	if logToFile {
		return "file: " + logFilePath
	}
	return "console"
}

// Close closes any open resources
func Close() {
	if logFile != nil {
		logFile.Close()
	}
}

// Debug logs a debug message
func Debug(format string, v ...interface{}) {
	if !loggerInit {
		// Default to standard log if not initialized
		log.Printf("[DEBUG] "+format, v...)
		return
	}
	
	if level >= LevelDebug {
		logger.Printf("[DEBUG] "+format, v...)
	}
}

// Info logs an informational message
func Info(format string, v ...interface{}) {
	if !loggerInit {
		// Default to standard log if not initialized
		log.Printf("[INFO] "+format, v...)
		return
	}
	
	if level >= LevelInfo {
		logger.Printf("[INFO] "+format, v...)
	}
}

// Error logs an error message
func Error(format string, v ...interface{}) {
	if !loggerInit {
		// Default to standard log if not initialized
		log.Printf("[ERROR] "+format, v...)
		return
	}
	
	// Always log errors regardless of level
	logger.Printf("[ERROR] "+format, v...)
}
