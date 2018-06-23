package log

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	config zap.Config
	root   *zap.Logger
	logger *zap.SugaredLogger
)

func init() {
	var err error
	config = zap.NewProductionConfig()
	config.EncoderConfig = zap.NewProductionEncoderConfig()
	config.EncoderConfig.TimeKey = "time"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder // "2006-01-02T15:04:05.000Z0700"
	root, err = config.Build(zap.AddCallerSkip(2))
	if err != nil {
		panic(err)
	}
	logger = root.Sugar()
}

// rfc3339TimeEncoder
func rfc3339TimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(time.RFC3339Nano))
}

// Named returns a named logger
func Named(name string) *zap.SugaredLogger {
	return logger.Named(name)
}

// With returns a logger with the given structured context
func With(keysAndValues ...interface{}) *zap.SugaredLogger {
	return logger.With(keysAndValues...)
}

// SetDebug log level
func SetDebug() {
	config.Level.SetLevel(zap.DebugLevel)
}

// SetInfo log level
func SetInfo() {
	config.Level.SetLevel(zap.InfoLevel)
}

// SetWarn log level
func SetWarn() {
	config.Level.SetLevel(zap.WarnLevel)
}

// SetError log level
func SetError() {
	config.Level.SetLevel(zap.ErrorLevel)
}

// Info logs a message with some additional context with variadic key-value pairs.
func Info(msg string, keysAndValues ...interface{}) {
	logger.Infow(msg, keysAndValues...)
}

// Warn logs a message with some additional context with variadic key-value pairs.
func Warn(msg string, keysAndValues ...interface{}) {
	logger.Warnw(msg, keysAndValues...)
}

// Error logs a message with some additional context with variadic key-value pairs.
func Error(msg string, keysAndValues ...interface{}) {
	logger.Errorw(msg, keysAndValues...)
}

// Debug logs a message with some additional context with variadic key-value pairs.
func Debug(msg string, keysAndValues ...interface{}) {
	logger.Debugw(msg, keysAndValues...)
}
