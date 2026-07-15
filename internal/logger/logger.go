package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	DefaultLogDir      = ".sloper/logs"
	DefaultLogFileName = "sloper.log"
	DefaultMaxSizeMB   = 50
	DefaultMaxBackups  = 5
	DefaultMaxAgeDays  = 30
)

var (
	globalLogger atomic.Pointer[zap.Logger]
	initOnce     sync.Once
)

func Default() *zap.Logger {
	l := globalLogger.Load()
	if l == nil {
		initOnce.Do(func() {
			l2 := newConsoleLogger(zapcore.InfoLevel)
			globalLogger.Store(l2)
		})
		l = globalLogger.Load()
	}
	return l
}

func Init(logDir string) error {
	if logDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("logger: determine home dir: %w", err)
		}
		logDir = filepath.Join(homeDir, DefaultLogDir)
	}

	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return fmt.Errorf("logger: create log dir %s: %w", logDir, err)
	}

	logFile := filepath.Join(logDir, DefaultLogFileName)

	fileEncoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})

	consoleEncoder := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05"),
		EncodeDuration: zapcore.StringDurationEncoder,
	})

	fileWriter := &rotatingWriter{
		path:       logFile,
		maxSize:    DefaultMaxSizeMB * 1024 * 1024,
		maxBackups: DefaultMaxBackups,
	}

	fileCore := zapcore.NewCore(fileEncoder, zapcore.AddSync(fileWriter), zapcore.DebugLevel)
	consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stderr), zapcore.InfoLevel)

	core := zapcore.NewTee(fileCore, consoleCore)
	l := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(0))

	globalLogger.Store(l)

	// Replace the standard library's log package output so existing
	// log.Printf calls also go through zap until they are migrated.
	zap.ReplaceGlobals(l)

	return nil
}

func Sync() {
	if l := globalLogger.Load(); l != nil {
		_ = (*l).Sync()
	}
}

func WithIssue(number int64) zap.Field {
	return zap.Int64("issue", number)
}

func WithPR(number int64) zap.Field {
	return zap.Int64("pr", number)
}

func WithStage(stage string) zap.Field {
	return zap.String("stage", stage)
}

func newConsoleLogger(level zapcore.Level) *zap.Logger {
	encoderCfg := zap.NewDevelopmentEncoderConfig()
	encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encoderCfg.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05")
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderCfg),
		zapcore.AddSync(os.Stderr),
		level,
	)
	return zap.New(core)
}

type rotatingWriter struct {
	mu          sync.Mutex
	path        string
	maxSize     int
	maxBackups  int
	currentSize int64
}

func (w *rotatingWriter) Write(data []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	info, err := os.Stat(w.path)
	if err == nil && info.Size() > int64(w.maxSize) {
		w.rotate(info.Size())
	}

	f, err := os.OpenFile(w.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return 0, fmt.Errorf("logger: open %s: %w", w.path, err)
	}
	defer f.Close()

	n, err := f.Write(data)
	w.currentSize += int64(n)
	return n, err
}

func (w *rotatingWriter) rotate(currentSize int64) {
	for i := w.maxBackups - 1; i >= 1; i-- {
		oldPath := fmt.Sprintf("%s.%d", w.path, i)
		newPath := fmt.Sprintf("%s.%d", w.path, i+1)
		_ = os.Rename(oldPath, newPath)
	}

	_ = os.Rename(w.path, w.path+".1")
	w.currentSize = 0
}
