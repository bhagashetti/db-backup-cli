package logs

import (
	"log"
	"os"
	"time"
)

var Logger *log.Logger

const maxLogSizeBytes = 5 * 1024 * 1024 // 5 MB

// Init sets up the global logger that writes to the given file path.
func Init(logPath string) (*os.File, error) {
	// If file exists and is too big, rotate it.
	if info, err := os.Stat(logPath); err == nil {
		if info.Size() > maxLogSizeBytes {
			timestamp := time.Now().Format("20060102-150405")
			rotated := logPath + "." + timestamp
			_ = os.Rename(logPath, rotated)
		}
	}

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	Logger = log.New(f, "", log.LstdFlags)
	return f, nil
}

func Info(format string, v ...any) {
	if Logger != nil {
		Logger.Printf("[INFO] "+format, v...)
	}
}

func Error(format string, v ...any) {
	if Logger != nil {
		Logger.Printf("[ERROR] "+format, v...)
	}
}
