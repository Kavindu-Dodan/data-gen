package generators

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"go.elastic.co/ecszap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LogGenerator generate logs in Elastic log format based on ZAP
type LogGenerator struct {
	logger *zap.Logger
	writer writer
	shChan chan struct{}
}

func newLogGenerator() *LogGenerator {
	stream := make(chan []byte, 1)
	w := writer{stream}

	shutdown := make(chan struct{})

	encoderConfig := ecszap.NewDefaultEncoderConfig()
	core := ecszap.NewCore(encoderConfig, zapcore.AddSync(&w), zap.DebugLevel)

	logger := zap.New(core, zap.AddCaller())

	return &LogGenerator{
		logger: logger,
		writer: w,
		shChan: shutdown,
	}
}

func (l LogGenerator) Start(delay time.Duration, errChan chan<- error) <-chan []byte {
	go func() {
		for {
			select {
			case <-time.After(delay):
				l.logger.Info(fmt.Sprintf("log entry: %s", uuid.NewString()))
			case _ = <-l.shChan:
				slog.Info("shutting down log generator")
				return
			}
		}
	}()

	return l.writer.emitter
}

func (l LogGenerator) Stop() {
	close(l.shChan)
}

// writer helps to extract logs and emit through com chan
type writer struct {
	emitter chan []byte
}

func (w *writer) Emitter() chan []byte {
	return w.emitter
}

func (w *writer) Write(p []byte) (n int, err error) {
	w.emitter <- p

	return len(p), nil
}
