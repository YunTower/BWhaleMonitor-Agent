package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	White  = "\033[37m"
)

type Logger struct {
	fileLogger *log.Logger
	console    *log.Logger
	mu         sync.Mutex
}

func NewLogger(logDir string) (*Logger, error) {
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	date := time.Now().Format("2006-01-02")
	filePath := filepath.Join(logDir, date+".txt")

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return &Logger{
		fileLogger: log.New(file, "", log.LstdFlags),
		console:    log.New(os.Stdout, "", log.LstdFlags),
	}, nil
}

func (l *Logger) Info(format string, v ...interface{}) {
	l.log(White, "INFO", format, v...) // 将颜色改为白色
}

func (l *Logger) Warn(format string, v ...interface{}) {
	l.log(Yellow, "WARN", format, v...)
}

func (l *Logger) Error(format string, v ...interface{}) {
	l.log(Red, "ERROR", format, v...)
}

func (l *Logger) Success(format string, v ...interface{}) { // 新增 Success 方法
	l.log(Green, "SUCCESS", format, v...)
}

func (l *Logger) log(color, level, format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	msg := fmt.Sprintf(format, v...)
	l.fileLogger.Printf("[%s] %s", level, msg)
	l.console.Printf("%s[%s] %s%s", color, level, msg, Reset)
}
