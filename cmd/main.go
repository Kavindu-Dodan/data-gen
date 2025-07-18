package main

import (
	"context"
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

	configurations, err := conf.NewConfig(b)
	if err != nil {
		slog.Error(fmt.Sprintf("Config file parsomg error: %s", err.Error()))
		return
	}

	slog.Info(fmt.Sprintf("Starting with configurations: %v", configurations.Print()))

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	generator, err := generators.GeneratorFor(configurations)
	if err != nil {
		slog.Error(fmt.Sprintf("Error creating generator: %s", err.Error()))
		return
	}

	exporter, err := exporters.ExporterFor(ctx, configurations)
	if err != nil {
		slog.Error(fmt.Sprintf("Error creating exporter: %s", err.Error()))
		return
	}

	errChan := make(chan error, 2)

	duration, err := time.ParseDuration(configurations.Input.Delay)
	if err != nil {
		slog.Error(fmt.Sprintf("Error parsing delay: %s, please provide value in acceptable string format like `5s`", err.Error()))
		return
	}

	outChan := generator.Start(duration, errChan)
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
