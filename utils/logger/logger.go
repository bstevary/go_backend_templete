package logger

import (
	"io"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

type LogConfig struct {
	LogFilePath string
	Environment string // e.g., "debug", "staging", "production"
}

// SetupLogger configures the global zerolog logger with file rotation and console output.
func SetupLogger(config LogConfig) {
	// Determine the directory for logs.
	logDir := filepath.Dir(config.LogFilePath)
	if logDir != "." { // Only attempt to create directory if it's not the current directory
		if _, err := os.Stat(logDir); os.IsNotExist(err) {
			err := os.MkdirAll(logDir, 0755) // Create directory with read/write/execute permissions for owner, read/execute for others
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to create log directory")
			}
		}
	}

	// Configure Lumberjack for log rotation
	fileLogger := &lumberjack.Logger{
		Filename:   config.LogFilePath,
		MaxSize:    10,   // megabytes, e.g., 10 for 10MB
		MaxBackups: 0,    // 0 means no old log files are removed based on count
		MaxAge:     0,    // 0 means no old log files are removed based on age
		Compress:   true, // compress old log files with gzip
	}

	var writers []io.Writer
	// Always write to the file logger
	writers = append(writers, fileLogger)

	// Set global Zerolog level based on environment
	switch config.Environment {
	case "debug":
		// In debug mode, also write to console for easier development
		consoleWriter := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "2006-01-02 15:04:05"}
		writers = append(writers, consoleWriter)
		zerolog.SetGlobalLevel(zerolog.DebugLevel) // Set to DebugLevel for debug
	case "test":
		// In test mode, set to DebugLevel and only log to file
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "production":
		// In production mode, set to InfoLevel and only log to file
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	default:
		log.Fatal().Msg("Invalid environment specified")
	}

	// Create a multi-writer to send logs to both file and console (if debug)
	multiWriter := io.MultiWriter(writers...)

	// Configure Zerolog global logger to use the multi-writer
	// Add Timestamp() and Caller() to include time and file/line number in all logs
	log.Logger = zerolog.New(multiWriter).With().Timestamp().Caller().Logger()

	// If in debug mode, enhance the console writer with PID and set specific logger for console
	// This ensures console output has the PID, while file output uses the multiWriter.
	if config.Environment == "debug" {
		log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "2006-01-02 15:04:05"}).
			Level(zerolog.DebugLevel).
			With().
			Timestamp().
			Caller().
			Int("pid", os.Getpid()).
			Logger()
	}

}
