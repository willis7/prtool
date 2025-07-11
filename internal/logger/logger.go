package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

var (
	instance *Logger
	once     sync.Once
)

type Logger struct {
	verbose    bool
	ci         bool
	logFile    io.WriteCloser
	fileLogger *log.Logger
	mu         sync.Mutex
}

// Init initializes the global logger
func Init(verbose bool, ci bool, logFilePath string) error {
	var err error
	once.Do(func() {
		instance = &Logger{
			verbose: verbose,
			ci:      ci,
		}

		if logFilePath != "" {
			instance.logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				return
			}
			instance.fileLogger = log.New(instance.logFile, "", log.LstdFlags)
		}
	})
	return err
}

// Close closes the log file if open
func Close() error {
	if instance != nil && instance.logFile != nil {
		return instance.logFile.Close()
	}
	return nil
}

// Verbose logs a message only if verbose mode is enabled
func Verbose(format string, args ...interface{}) {
	if instance == nil || !instance.verbose {
		return
	}

	msg := fmt.Sprintf(format, args...)
	instance.mu.Lock()
	defer instance.mu.Unlock()

	if instance.ci {
		// In CI mode, prefix verbose messages
		fmt.Fprintf(os.Stderr, "[VERBOSE] %s\n", msg)
	} else {
		fmt.Fprintf(os.Stderr, "%s\n", msg)
	}

	if instance.fileLogger != nil {
		instance.fileLogger.Printf("[VERBOSE] %s", msg)
	}
}

// Info logs an informational message
func Info(format string, args ...interface{}) {
	if instance == nil {
		fmt.Fprintf(os.Stdout, format+"\n", args...)
		return
	}

	msg := fmt.Sprintf(format, args...)
	instance.mu.Lock()
	defer instance.mu.Unlock()

	if instance.ci {
		// In CI mode, use structured output
		fmt.Fprintf(os.Stdout, "[INFO] %s\n", msg)
	} else {
		fmt.Fprintf(os.Stdout, "%s\n", msg)
	}

	if instance.fileLogger != nil {
		instance.fileLogger.Printf("[INFO] %s", msg)
	}
}

// Warn logs a warning message
func Warn(format string, args ...interface{}) {
	if instance == nil {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
		return
	}

	msg := fmt.Sprintf(format, args...)
	instance.mu.Lock()
	defer instance.mu.Unlock()

	if instance.ci {
		fmt.Fprintf(os.Stderr, "[WARN] %s\n", msg)
	} else {
		fmt.Fprintf(os.Stderr, "Warning: %s\n", msg)
	}

	if instance.fileLogger != nil {
		instance.fileLogger.Printf("[WARN] %s", msg)
	}
}

// Error logs an error message
func Error(format string, args ...interface{}) {
	if instance == nil {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
		return
	}

	msg := fmt.Sprintf(format, args...)
	instance.mu.Lock()
	defer instance.mu.Unlock()

	if instance.ci {
		fmt.Fprintf(os.Stderr, "[ERROR] %s\n", msg)
	} else {
		fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
	}

	if instance.fileLogger != nil {
		instance.fileLogger.Printf("[ERROR] %s", msg)
	}
}

// Progress shows a progress message (suppressed in CI mode)
func Progress(format string, args ...interface{}) {
	if instance != nil && instance.ci {
		// Suppress progress in CI mode
		return
	}

	msg := fmt.Sprintf(format, args...)
	if instance != nil {
		instance.mu.Lock()
		defer instance.mu.Unlock()

		if instance.fileLogger != nil {
			instance.fileLogger.Printf("[PROGRESS] %s", msg)
		}
	}

	// Always show progress to stderr when not in CI mode
	fmt.Fprintf(os.Stderr, "%s\n", msg)
}

// IsCI returns whether CI mode is enabled
func IsCI() bool {
	return instance != nil && instance.ci
}

// IsVerbose returns whether verbose mode is enabled
func IsVerbose() bool {
	return instance != nil && instance.verbose
}