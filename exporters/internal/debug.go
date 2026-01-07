package internal

import (
	"fmt"
	"log/slog"
	"strings"

	"data-gen/conf"
)

// DebugExporter writes generated data to stdout for debugging purposes.
// Similar to the OpenTelemetry collector debug exporter, it's useful for
// testing and development to see what data is being generated.
type DebugExporter struct {
	cfg debugCfg
}

// debugCfg specifies the debug exporter configuration.
type debugCfg struct {
	// Verbosity controls the output format: "basic", "normal", "detailed"
	Verbosity string `yaml:"verbosity"`
	// ShowTimestamp adds timestamp to each log entry
	ShowTimestamp bool `yaml:"show_timestamp"`
	// Prefix to add before each output line
	Prefix string `yaml:"prefix"`
}

func newDefaultDebugCfg() *debugCfg {
	return &debugCfg{
		Verbosity:     "normal",
		ShowTimestamp: true,
		Prefix:        "[DEBUG]",
	}
}

func NewDebugExporter(config *conf.Config) (*DebugExporter, error) {
	cfg := newDefaultDebugCfg()

	var override *debugCfg
	err := config.Output.Conf.Decode(&override)
	if err != nil {
		return nil, err
	}

	if override != nil {
		cfg = override
	}
	return &DebugExporter{
		cfg: *cfg,
	}, nil
}

func (d *DebugExporter) Send(data *[]byte) error {
	switch d.cfg.Verbosity {
	case "basic":
		// Just show that data was received
		slog.Info(fmt.Sprintf("%s Data received (%d bytes)\n", d.cfg.Prefix, len(*data)))
	case "detailed":
		// Show full data with formatting
		d.writeDetailed(data)
	default: // "normal"
		// Show data without extra formatting
		d.writeNormal(data)
	}

	return nil
}

func (d *DebugExporter) writeNormal(data *[]byte) {
	dataStr := string(*data)
	lines := strings.Split(strings.TrimSpace(dataStr), "\n")

	for _, line := range lines {
		if line != "" {
			slog.Info(fmt.Sprintf("%s %s\n", d.cfg.Prefix, line))
		}
	}
}

func (d *DebugExporter) writeDetailed(data *[]byte) {
	dataStr := string(*data)
	lines := strings.Split(strings.TrimSpace(dataStr), "\n")

	slog.Info(fmt.Sprintf("%s ========================================\n", d.cfg.Prefix))
	slog.Info(fmt.Sprintf("%s Data Size: %d bytes\n", d.cfg.Prefix, len(*data)))
	slog.Info(fmt.Sprintf("%s Number of Lines: %d\n", d.cfg.Prefix, len(lines)))
	slog.Info(fmt.Sprintf("%s ----------------------------------------\n", d.cfg.Prefix))

	for i, line := range lines {
		if line != "" {
			slog.Info(fmt.Sprintf("%s [%d] %s\n", d.cfg.Prefix, i+1, line))

		}
	}

	slog.Info(fmt.Sprintf("%s ========================================\n", d.cfg.Prefix))
}
