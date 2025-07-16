package utils

import (
    "log"
    "os"
)

type Logger struct {
    logger *log.Logger
    level  string
}

func InitLoggerLevel(level string) *Logger {
    return &Logger{
        logger: log.New(os.Stdout, "", log.LstdFlags),
        level:  level,
    }
}

func (l *Logger) logf(prefix string, msg string, args ...interface{}) {
    if l.level == "debug" || (l.level == "info" && prefix != "[DEBUG]") {
        l.logger.Printf("%s %s", prefix, msg)
    }
}

func (l *Logger) Debug(msg string, args ...interface{}) { l.logf("[DEBUG]", msg, args...) }
func (l *Logger) Info(msg string, args ...interface{})  { l.logf("[INFO]", msg, args...) }
func (l *Logger) Error(msg string, args ...interface{}) { l.logf("[ERROR]", msg, args...) }
func (l *Logger) Warn(msg string, args ...interface{})  { l.logf("[WARN]", msg, args...) }  
func (l *Logger) Fatal(msg string, args ...interface{}) {
    l.logf("[FATAL]", msg, args...)
    os.Exit(1)
}   
func (l *Logger) Panic(msg string, args ...interface{}) {
    l.logf("[PANIC]", msg, args...)
    panic(msg)
}   
func (l *Logger) SetLevel(level string) {
    l.level = level
}   
func (l *Logger) GetLevel() string {
    return l.level
}   
func (l *Logger) Close() {
    // No resources to close in this simple logger
}       
func (l *Logger) Println(msg string) {
    l.logger.Println(msg)
}
func (l *Logger) Printf(format string, args ...interface{}) {
    l.logger.Printf(format, args...)
}
func (l *Logger) Print(args ...interface{}) {
    l.logger.Print(args...)
}
