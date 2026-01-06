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
		slog.Error("Config file reading error", "error", err, "file", *cfgLocation)
		return
	}

	configurations, err := conf.NewConfig(b)
	if err != nil {
		slog.Error("Config file parsing error", "error", err)
		return
	}

	logLevel := slog.LevelInfo
	if *debug {
		logLevel = slog.LevelDebug
	}

	// Use a cleaner text handler with better formatting
	opts := &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: false,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Format time as shorter HH:MM:SS instead of full RFC3339
			if a.Key == slog.TimeKey {
				t := a.Value.Time()
				return slog.String(slog.TimeKey, t.Format("15:04:05"))
			}
			return a
		},
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, opts)))

	// Print configuration in a readable format
	slog.Info("Starting data generator")
	fmt.Fprintf(os.Stdout, "%s\n", configurations.Print())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	generator, err := generators.GeneratorFor(configurations)
	if err != nil {
		slog.Error("Error creating generator", "error", err, "type", configurations.Input.Type)
		return
	}

	exporter, err := exporters.ExporterFor(ctx, configurations)
	if err != nil {
		slog.Error("Error creating exporter", "error", err, "type", configurations.Output.Type)
		return
	}

	duration, err := time.ParseDuration(configurations.Input.Delay)
	if err != nil {
		slog.Error("Error parsing delay", "error", err, "delay", configurations.Input.Delay, "hint", "use format like '5s' or '10ms'")
		return
	}

	dataInput, inputClose, genError := generator.Start(duration)
	expErr := exporter.Start(dataInput)

	select {
	case <-sigs:
		slog.Info("Received shutdown signal")
	case <-inputClose:
		slog.Info("Generator completed")
	case er := <-genError:
		slog.Error("Error from generator", "error", er)
	case er := <-expErr:
		slog.Error("Error from exporter", "error", er)
	}

	slog.Info("Shutting down...")
	// provide a grace period to complete the exports
	<-time.After(time.Second * 2)
	generator.Stop()
	exporter.Stop()

}
