package main

import (
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
	processors = 2
	delay      = time.Second * 1
	location   = "./logs"
)

func main() {
	// logs to file with rotation
	lumberJack := zapcore.AddSync(&lumberjack.Logger{
		Filename: location,
		MaxSize:  10,
	})

	encoderConfig := ecszap.NewDefaultEncoderConfig()
	core := ecszap.NewCore(encoderConfig, lumberJack, zap.DebugLevel)
	logger := zap.New(core, zap.AddCaller())

	runner := newRunner(logger)
	runner.start(processors, delay)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	select {
	case <-sigs:
		runner.end()
		<-time.After(5 * time.Second)
	}
}
