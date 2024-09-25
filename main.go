package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/firehose"
	"github.com/aws/aws-sdk-go-v2/service/firehose/types"
	"github.com/natefinch/lumberjack"
	"go.elastic.co/ecszap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	region             = "us-east-1"
	firehoseStreamName = "kavindu-firehose"
	outFirehose        = "FIREHOSE"
	outFile            = "FILE"
	typeLogs           = "LOGS"
	typeMetrics        = "METRICS"
)

type Cfg struct {
	Type           string `json:"type"`
	LogOutput      string `json:"output,logOutput"`
	Processors     int    `json:"processors,omitempty"`
	Delay          int    `json:"delay,omitempty"`
	LogLocation    string `json:"log_location,omitempty"`
	MaxLogFileSize int    `json:"max_log_file_size,omitempty"`
}

// NOTE -  current configurations
var (
	output           = outFirehose
	processors       = 1
	delay            = 1
	location         = "./logs"
	maxLogFileSizeMb = 10

	currentConfig = Cfg{
		LogOutput:      output,
		Processors:     processors,
		Delay:          delay,
		LogLocation:    location,
		MaxLogFileSize: maxLogFileSizeMb,
	}
)

func main() {
	marshal, err := json.Marshal(currentConfig)
	if err != nil {
		slog.Error("Error marshalling config", err)
		return
	}
	slog.Info(fmt.Sprintf("Starting with configurations: %v", string(marshal)))

	shutdownChan := make(chan interface{})
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	// choose log writer
	var sync zapcore.WriteSyncer
	switch currentConfig.LogOutput {
	case outFile:
		// logs to file with rotation
		sync = zapcore.AddSync(&lumberjack.Logger{
			Filename: location,
			MaxSize:  maxLogFileSizeMb,
		})
	case outFirehose:
		// firehose emitter
		gCtx := context.Background()

		//cfg, err := config.LoadDefaultConfig(gCtx, config.WithSharedConfigProfile("ecdev"), config.WithRegion(region))
		cfg, err := config.LoadDefaultConfig(gCtx)

		if err != nil {
			slog.Error("error loading default configurations: ", err)
			return
		}

		fhClient := firehose.New(firehose.Options{
			Credentials: cfg.Credentials,
			Region:      region,
		})
		comChan := make(chan []byte, 1)
		writer := newFireHoseWriter(firehoseStreamName, fhClient, comChan, shutdownChan)
		sync = zapcore.AddSync(writer)
	default:
		slog.Error(fmt.Sprintf("Log output '%s' is not supported", currentConfig.LogOutput))
		return
	}

	// elastic encoder
	encoderConfig := ecszap.NewDefaultEncoderConfig()
	core := ecszap.NewCore(encoderConfig, sync, zap.DebugLevel)

	logger := zap.New(core, zap.AddCaller())
	runner := newRunner(logger)
	runner.start(processors, time.Duration(delay)*time.Second)

	select {
	case <-sigs:
		runner.end()
		close(shutdownChan)
		<-time.After(1 * time.Second)
	}

	slog.Info("Exiting")
}

type fireHoseWriter struct {
	client     *firehose.Client
	streamName string
	listener   chan []byte
	stopChan   chan interface{}
}

func newFireHoseWriter(streamName string, client *firehose.Client, listener chan []byte, stopChan chan interface{}) *fireHoseWriter {
	f := &fireHoseWriter{streamName: streamName, client: client, listener: listener, stopChan: stopChan}
	f.startFireHoseEmitter()

	return f
}

func (w fireHoseWriter) Write(p []byte) (n int, err error) {
	w.listener <- p

	return len(p), nil
}

func (w fireHoseWriter) startFireHoseEmitter() {
	// run in background
	go func() {
		for {
			select {
			case <-w.stopChan:
				// exit
				return
			case line := <-w.listener:
				record := types.Record{
					Data: line,
				}

				putRecord := firehose.PutRecordInput{
					DeliveryStreamName: &w.streamName,
					Record:             &record,
				}

				_, err := w.client.PutRecord(context.Background(), &putRecord)
				if err != nil {
					slog.Error("error from firehose put record: ", err)
					return
				}
			}
		}
	}()
}
