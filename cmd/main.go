package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"data-gen/conf"
	"data-gen/exporters"
	"data-gen/generators"
)

func main() {
	ctx, signalStop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	args := parseArgs()
	b, err := os.ReadFile(args.configPath)
	if err != nil {
		slog.Error("Config file reading error", "error", err, "file", args.configPath)
		return
	}

	configurations, err := conf.NewConfig(b)
	if err != nil {
		slog.Error("Config file parsing error", "error", err)
		return
	}

	logLevel := slog.LevelInfo
	if args.debug {
		logLevel = slog.LevelDebug
	}

	// Use a cleaner text handler with better formatting
	opts := &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: false,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				t := a.Value.Time()
				return slog.String(slog.TimeKey, t.Format("2006-01-02T15:04:05"))
			}
			return a
		},
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, opts)))

	slog.Info("Starting data generator")
	slog.Info("Input", "config", configurations.Input.Print())
	slog.Info("Output", "config", configurations.Output.Print())
	if configurations.UsesAWS() {
		slog.Info("AWS", "config", configurations.AWSCfg.Print())
	}

	err = run(ctx, signalStop, configurations)
	if err != nil {
		slog.Error(fmt.Sprintf("Runtime error: %s", err.Error()))
		return
	}
}

// run starts the data generator and exporter based on the provided configuration.
// This is a blocking call that runs until a termination signal is received or an error occurs.
func run(ctx context.Context, sigStop context.CancelFunc, cfg *conf.Config) error {
	generator, err := generators.GeneratorFor(cfg)
	if err != nil {
		return fmt.Errorf("error creating generator: %s", err.Error())
	}

	exporter, err := exporters.ExporterFor(ctx, cfg)
	if err != nil {
		return fmt.Errorf("error creating exporter: %s", err.Error())
	}

	dataInput, inputComplete, genError := generator.Start()
	expErr := exporter.Start(dataInput)

	select {
	case <-ctx.Done():
		slog.Info("Context cancelled, shutting down...")
	case <-inputComplete:
		slog.Info("Input completed, shutting down...")
	case er := <-genError:
		slog.Error("Error from generator", "error", er)
	case er := <-expErr:
		slog.Error("Error from exporter", "error", er)
	}

	// stop listening to signals & stop the generator
	sigStop()
	generator.Stop()
	exporter.Stop()

	return nil
}

type flags struct {
	configPath string
	debug      bool
}

func parseArgs() flags {
	cfgPath := flag.String("config", "./config.yaml", "configuration file. Default to `./config.yaml`")
	debug := flag.Bool("debug", false, "enable debug logging")
	flag.Parse()

	return flags{
		configPath: *cfgPath,
		debug:      *debug,
	}
}
