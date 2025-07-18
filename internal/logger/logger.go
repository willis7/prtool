package logger

import (
	"fmt"
	"io"
	"log"
	"os"
)

// Logger provides structured logging for prtool
type Logger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
	verbose     bool
	ci          bool
}

// New creates a new logger instance
func New(verbose, ci bool, logFile string) (*Logger, error) {
	var logWriter io.Writer = os.Stderr

	// If log file is specified, write to file instead
	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		logWriter = file
	}

	// In CI mode, disable timestamps and prefixes for cleaner output
	flags := log.LstdFlags
	if ci {
		flags = 0
	}

	return &Logger{
		infoLogger:  log.New(logWriter, "", flags),
		errorLogger: log.New(os.Stderr, "", flags),
		verbose:     verbose,
		ci:          ci,
	}, nil
}

// Info logs an informational message (only if verbose is enabled)
func (l *Logger) Info(format string, args ...interface{}) {
	if l.verbose {
		l.infoLogger.Printf(format, args...)
	}
}

// Error logs an error message (always shown)
func (l *Logger) Error(format string, args ...interface{}) {
	l.errorLogger.Printf(format, args...)
}

// Progress logs a progress message (suppressed in CI mode)
func (l *Logger) Progress(format string, args ...interface{}) {
	if !l.ci {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	}
}

// Output writes to stdout (for actual output, not logging)
func (l *Logger) Output(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}
