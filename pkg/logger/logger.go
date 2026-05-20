package logger

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Level = zapcore.Level

const (
	DebugLevel Level = zap.DebugLevel
	InfoLevel  Level = zap.InfoLevel
	WarnLevel  Level = zap.WarnLevel
	ErrorLevel Level = zap.ErrorLevel
	PanicLevel Level = zap.PanicLevel
	FatalLevel Level = zap.FatalLevel
)

func WithLevel(level string) Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn", "warning":
		return WarnLevel
	case "error":
		return ErrorLevel
	case "panic":
		return PanicLevel
	case "fatal":
		return FatalLevel
	default:
		return InfoLevel
	}
}

type Logger struct {
	logger *zap.Logger
	sugar  *zap.SugaredLogger
	level  Level
}

type RotateOptions struct {
	FileName   string
	MaxSize    int
	MaxAge     int
	MaxBackups int
	Compress   bool
}

type Option = zap.Option

var (
	AddCaller     = zap.AddCaller
	AddStacktrace = zap.AddStacktrace
	AddCallerSkip = zap.AddCallerSkip
)

func newEncoderConfig() zapcore.EncoderConfig {
	encCfg := zap.NewProductionEncoderConfig()
	encCfg.TimeKey = "time"
	encCfg.EncodeTime = func(t time.Time, pae zapcore.PrimitiveArrayEncoder) {
		pae.AppendString(t.Format("2006-01-02 15:04:05"))
	}
	return encCfg
}

func NewBootstrapLogger(options ...Option) *Logger {
	base := zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(newEncoderConfig()),
			zapcore.AddSync(os.Stdout),
			InfoLevel,
		),
		options...,
	)
	return &Logger{
		logger: base,
		sugar:  base.Sugar(),
		level:  InfoLevel,
	}
}

func NewLogger(level zapcore.Level, ropt RotateOptions, options ...Option) *Logger {
	if ropt.FileName != "" {
		_ = os.MkdirAll(filepath.Dir(ropt.FileName), 0o755)
	}

	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   ropt.FileName,
		MaxSize:    ropt.MaxSize,
		MaxBackups: ropt.MaxBackups,
		MaxAge:     ropt.MaxAge,
		Compress:   ropt.Compress,
	})
	consoleWriter := zapcore.AddSync(os.Stdout)
	ws := zapcore.NewMultiWriteSyncer(fileWriter, consoleWriter)

	base := zap.New(zapcore.NewCore(zapcore.NewJSONEncoder(newEncoderConfig()), ws, level), options...)
	return &Logger{
		logger: base,
		sugar:  base.Sugar(),
		level:  level,
	}
}

func (l *Logger) StdLogger() *log.Logger { return zap.NewStdLog(l.logger) }

func (l *Logger) Debug(msg string, fields ...zap.Field) { l.logger.Debug(msg, fields...) }
func (l *Logger) Info(msg string, fields ...zap.Field)  { l.logger.Info(msg, fields...) }
func (l *Logger) Warn(msg string, fields ...zap.Field)  { l.logger.Warn(msg, fields...) }
func (l *Logger) Error(msg string, fields ...zap.Field) { l.logger.Error(msg, fields...) }
func (l *Logger) Fatal(msg string, fields ...zap.Field) { l.logger.Fatal(msg, fields...) }

func (l *Logger) Debugf(format string, args ...any) { l.sugar.Debugf(format, args...) }
func (l *Logger) Infof(format string, args ...any)  { l.sugar.Infof(format, args...) }
func (l *Logger) Warnf(format string, args ...any)  { l.sugar.Warnf(format, args...) }
func (l *Logger) Errorf(format string, args ...any) { l.sugar.Errorf(format, args...) }
func (l *Logger) Fatalf(format string, args ...any) { l.sugar.Fatalf(format, args...) }

func (l *Logger) Sync() error {
	if l == nil || l.logger == nil {
		return nil
	}
	if err := l.logger.Sync(); err != nil {
		if runtime.GOOS == "windows" && strings.Contains(strings.ToLower(err.Error()), "invalid argument") {
			return nil
		}
		return err
	}
	return nil
}
