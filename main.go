package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/natefinch/lumberjack"
	"go.elastic.co/ecszap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	processors       = 2
	delay            = 1
	location         = "./logs"
	maxLogFileSizeMb = 10

	currentConfig = config{
		Processors:     processors,
		LogDelay:       delay,
		LogLocation:    location,
		MaxLogFileSize: maxLogFileSizeMb,
	}
)

type config struct {
	Processors     int    `json:"processors,omitempty"`
	LogDelay       int    `json:"log_delay,omitempty"`
	LogLocation    string `json:"log_location,omitempty"`
	MaxLogFileSize int    `json:"max_log_file_size,omitempty"`
}

func main() {
	marshal, err := json.Marshal(currentConfig)
	if err != nil {
		slog.Error("Error marshalling config", err)
		return
	}

	slog.Info(fmt.Sprintf("Starting with configurations: %v", string(marshal)))

	// logs to file with rotation
	lumberJack := zapcore.AddSync(&lumberjack.Logger{
		Filename: location,
		MaxSize:  maxLogFileSizeMb,
	})

	encoderConfig := ecszap.NewDefaultEncoderConfig()
	core := ecszap.NewCore(encoderConfig, lumberJack, zap.DebugLevel)
	logger := zap.New(core, zap.AddCaller())

	runner := newRunner(logger)
	runner.start(processors, time.Duration(delay)*time.Second)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	select {
	case <-sigs:
		runner.end()
		<-time.After(1 * time.Second)
	}

	slog.Info("Exiting")
}
