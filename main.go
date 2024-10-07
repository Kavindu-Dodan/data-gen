package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"data-gen/conf"
	"data-gen/exporters"
	"data-gen/generators"
)

func main() {
	ctx := context.Background()

	var cfgLocation = flag.String("configFile", "./config.yaml", "configuration file. Default to `./config.yaml`")
	flag.Parse()

	b, err := os.ReadFile(*cfgLocation)
	if err != nil {
		slog.Error(fmt.Sprintf("Config file reading error: %s", err.Error()))
		return
	}

	configurations, err := conf.NewCfgFrom(b)
	if err != nil {
		slog.Error(fmt.Sprintf("Config file parsomg error: %s", err.Error()))
		return
	}

	marshal, err := json.Marshal(configurations)
	if err != nil {
		slog.Error("Error marshalling config", err)
		return
	}
	slog.Info(fmt.Sprintf("Starting with configurations: %v", string(marshal)))

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	generator, err := makeGenerator(configurations)
	if err != nil {
		slog.Error(fmt.Sprintf("Error creating generator: %s", err.Error()))
		return
	}

	exporter, err := makeExporter(ctx, configurations)
	if err != nil {
		slog.Error(fmt.Sprintf("Error creating exporter: %s", err.Error()))
		return
	}

	errChan := make(chan error, 2)
	outChan := generator.Start(time.Duration(configurations.Delay)*time.Second, errChan)
	exporter.Start(outChan, errChan)

	select {
	case <-sigs:
		generator.Stop()
		exporter.Stop()
		<-time.After(time.Second)
	case err := <-errChan:
		slog.Error(fmt.Sprintf("Error occured: %s", err.Error()))
	}

	slog.Info("Shutting down")
}

func makeGenerator(cfg *conf.Cfg) (generators.IGen, error) {
	switch cfg.Type {
	case conf.TypeLogs:
		return generators.NewLogGenerator(), nil
	case conf.TypeMetrics:
		return generators.NewMetricGenerator(cfg), nil
	default:
		return nil, fmt.Errorf("unknown generator type: %s", cfg.Type)
	}
}

func makeExporter(ctx context.Context, cfg *conf.Cfg) (exporters.IExport, error) {
	switch cfg.Output {
	case conf.OutFile:
		return exporters.NewFileExporter(cfg.FileLocation), nil
	case conf.OutFirehose:
		return exporters.NewFirehoseExporter(ctx, cfg.AWSCfg)
	case conf.OutCloudwatchLogs:
		return exporters.NewCloudWatchLogExporter(ctx, cfg.AWSCfg)
	default:
		return nil, fmt.Errorf("unknown exporter output: %s", cfg.Output)
	}
}
