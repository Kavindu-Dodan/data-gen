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

	var cfgLocation = flag.String("config", "./config.yaml", "configuration file. Default to `./config.yaml`")
	var debug = flag.Bool("debug", false, "enable debug logging")
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

	logLevel := slog.LevelInfo
	if *debug {
		logLevel = slog.LevelDebug
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})))

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

	duration, err := time.ParseDuration(configurations.Input.Delay)
	if err != nil {
		slog.Error(fmt.Sprintf("Error parsing delay: %s, please provide value in acceptable string format like `5s`", err.Error()))
		return
	}

	dataInput, inputClose, genError := generator.Start(duration)
	expErr := exporter.Start(dataInput)

	select {
	case <-sigs:
	case <-inputClose:
	case er := <-genError:
		slog.Error(fmt.Sprintf("Error from generator: %s", er.Error()))
	case er := <-expErr:
		slog.Error(fmt.Sprintf("Error from exporter: %s", er.Error()))
	}

	slog.Info("Shutting down...")
	// provide a grace period to complete the exports
	<-time.After(time.Second * 2)
	generator.Stop()
	exporter.Stop()

}
