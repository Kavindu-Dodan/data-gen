package exporters

import (
	"fmt"
	"os"
	"strings"

	"data-gen/conf"
)

// DebugExporter writes generated data to stdout for debugging purposes.
// Similar to the OpenTelemetry collector debug exporter, it's useful for
// testing and development to see what data is being generated.
type DebugExporter struct {
	cfg    debugCfg
	shChan chan struct{}
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

func newDebugExporter(config *conf.Config) (*DebugExporter, error) {
	cfg := newDefaultDebugCfg()
	err := config.Output.Conf.Decode(&cfg)
	if err != nil {
		// If no config is provided, use defaults
		cfg = newDefaultDebugCfg()
	}

	return &DebugExporter{
		cfg:    *cfg,
		shChan: make(chan struct{}),
	}, nil
}

func (d *DebugExporter) send(data *[]byte) error {
	switch d.cfg.Verbosity {
	case "basic":
		// Just show that data was received
		fmt.Fprintf(os.Stdout, "%s Data received (%d bytes)\n", d.cfg.Prefix, len(*data))
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
			fmt.Fprintf(os.Stdout, "%s %s\n", d.cfg.Prefix, line)
		}
	}
}

func (d *DebugExporter) writeDetailed(data *[]byte) {
	dataStr := string(*data)
	lines := strings.Split(strings.TrimSpace(dataStr), "\n")

	fmt.Fprintf(os.Stdout, "%s ========================================\n", d.cfg.Prefix)
	fmt.Fprintf(os.Stdout, "%s Data Size: %d bytes\n", d.cfg.Prefix, len(*data))
	fmt.Fprintf(os.Stdout, "%s Number of Lines: %d\n", d.cfg.Prefix, len(lines))
	fmt.Fprintf(os.Stdout, "%s ----------------------------------------\n", d.cfg.Prefix)

	for i, line := range lines {
		if line != "" {
			fmt.Fprintf(os.Stdout, "%s [%d] %s\n", d.cfg.Prefix, i+1, line)
		}
	}

	fmt.Fprintf(os.Stdout, "%s ========================================\n", d.cfg.Prefix)
}

func (d *DebugExporter) stop() {
	close(d.shChan)
}
