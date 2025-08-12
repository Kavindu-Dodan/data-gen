package generators

import (
	"fmt"

	"github.com/google/uuid"
	"go.elastic.co/ecszap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LogGenerator generate logs in Elastic log format based on ZAP
type LogGenerator struct {
	buf    trackedBuffer
	logger *zap.Logger
	writer *writer
	shChan chan struct{}
}

func NewLogGenerator() *LogGenerator {
	w := writer{}
	shutdown := make(chan struct{})

	encoderConfig := ecszap.NewDefaultEncoderConfig()
	core := ecszap.NewCore(encoderConfig, zapcore.AddSync(&w), zap.DebugLevel)
	logger := zap.New(core, zap.AddCaller())

	return &LogGenerator{
		buf:    newTrackedBuffer(),
		logger: logger,
		writer: &w,
		shChan: shutdown,
	}
}

func (l *LogGenerator) Generate() (int64, error) {
	l.logger.Info(fmt.Sprintf("log entry: %s", uuid.NewString()))
	err := l.buf.write(l.writer.data)

	return l.buf.size(), err
}

func (l *LogGenerator) GetAndReset() []byte {
	return l.buf.getAndRest()
}

// writer helps to extract logs and emit through com chan
type writer struct {
	data []byte
}

func (w *writer) Write(p []byte) (n int, err error) {
	w.data = p
	return len(p), nil
}
