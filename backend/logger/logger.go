package logger

import (
	"io"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.Logger
}

func NewLogger(output io.Writer) *Logger {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "log_timestamp"
	encoderConfig.LevelKey = "log_level"
	encoderConfig.MessageKey = "log_message"
	encoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder

	encoder := zapcore.NewJSONEncoder(encoderConfig)
	core := zapcore.NewCore(encoder, zapcore.AddSync(output), zapcore.InfoLevel)
	logger := zap.New(core)
	defer logger.Sync()

	return &Logger{Logger: logger}
}

func (l Logger) Error(msg string) {
	l.Logger.Error(msg)
}

func (l Logger) Info(msg string) {
	l.Logger.Info(msg)
}

func (l Logger) Warn(msg string) {
	l.Logger.Warn(msg)
}
