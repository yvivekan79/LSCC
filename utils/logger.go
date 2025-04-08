package utils

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
	"time"
)

// LogLevel is the verbosity level for logging
type LogLevel int

const (
	// LogLevelError only logs errors
	LogLevelError LogLevel = iota
	// LogLevelWarn logs errors and warnings
	LogLevelWarn
	// LogLevelInfo logs errors, warnings, and info
	LogLevelInfo
	// LogLevelDebug logs errors, warnings, info, and debug
	LogLevelDebug
	// LogLevelTrace logs everything
	LogLevelTrace
)

// Logger is a simple logging utility
type Logger struct {
	level  LogLevel
	logger *log.Logger
	mu     sync.Mutex
}

var (
	instance *Logger
	once     sync.Once
)

// GetLogger returns the singleton logger instance
func GetLogger() *Logger {
	once.Do(func() {
		instance = &Logger{
			level:  LogLevelInfo,
			logger: log.New(os.Stdout, "", log.LstdFlags),
		}
	})
	return instance
}

// InitLogger initializes the logger with the specified level
func InitLogger(verbosity int) {
	logger := GetLogger()
	level := LogLevel(verbosity)
	if level < LogLevelError {
		level = LogLevelError
	}
	if level > LogLevelTrace {
		level = LogLevelTrace
	}
	logger.level = level
}

// formatLog formats a log message with additional context
func (l *Logger) formatLog(level, msg string, keyvals ...interface{}) string {
	// Get caller info
	_, file, line, ok := runtime.Caller(2)
	caller := "unknown"
	if ok {
		// Extract just the filename, not the full path
		for i := len(file) - 1; i >= 0; i-- {
			if file[i] == '/' {
				file = file[i+1:]
				break
			}
		}
		caller = fmt.Sprintf("%s:%d", file, line)
	}
	
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	logMsg := fmt.Sprintf("[%s] [%s] [%s] %s", timestamp, level, caller, msg)
	
	// Add key-value pairs if any
	if len(keyvals) > 0 {
		if len(keyvals)%2 != 0 {
			// Ensure even number of elements by adding empty string
			keyvals = append(keyvals, "")
		}
		
		for i := 0; i < len(keyvals); i += 2 {
			key, val := keyvals[i], keyvals[i+1]
			logMsg += fmt.Sprintf(" %v=%v", key, val)
		}
	}
	
	return logMsg
}

// log logs a message with the given level
func (l *Logger) log(logLevel LogLevel, levelName, msg string, keyvals ...interface{}) {
	if logLevel > l.level {
		return
	}
	
	l.mu.Lock()
	defer l.mu.Unlock()
	
	formatted := l.formatLog(levelName, msg, keyvals...)
	l.logger.Println(formatted)
}

// Error logs an error message
func (l *Logger) Error(msg string, keyvals ...interface{}) {
	l.log(LogLevelError, "ERROR", msg, keyvals...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, keyvals ...interface{}) {
	l.log(LogLevelWarn, "WARN", msg, keyvals...)
}

// Info logs an info message
func (l *Logger) Info(msg string, keyvals ...interface{}) {
	l.log(LogLevelInfo, "INFO", msg, keyvals...)
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, keyvals ...interface{}) {
	l.log(LogLevelDebug, "DEBUG", msg, keyvals...)
}

// Trace logs a trace message
func (l *Logger) Trace(msg string, keyvals ...interface{}) {
	l.log(LogLevelTrace, "TRACE", msg, keyvals...)
}

// NewError creates a formatted error
func NewError(msg string, keyvals ...interface{}) error {
	errMsg := msg
	if len(keyvals) > 0 {
		if len(keyvals)%2 != 0 {
			keyvals = append(keyvals, "")
		}
		
		for i := 0; i < len(keyvals); i += 2 {
			key, val := keyvals[i], keyvals[i+1]
			errMsg += fmt.Sprintf(" %v=%v", key, val)
		}
	}
	return fmt.Errorf(errMsg)
}
