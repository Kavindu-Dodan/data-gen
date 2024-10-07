package conf

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"log/slog"
)

const (
	OutFirehose       = "FIREHOSE"
	OutFile           = "FILE"
	OutCloudwatchLogs = "CLOUDWATCH_LOG"
	TypeLogs          = "LOGS"
	TypeMetrics       = "METRICS"

	defaultProfile      = "default"
	defaultRegion       = "us-east-1"
	defaultFileLocation = "./out"
	defaultDelay        = 5
)

type Cfg struct {
	Type         string `yaml:"type"`
	Output       string `yaml:"output"`
	Delay        int    `yaml:"delay"`
	FileLocation string `yaml:"file_location,omitempty"`
	AWSCfg       `yaml:"aws,omitempty"`
}

type AWSCfg struct {
	Profile                 string `yaml:"profile"`
	Region                  string `yaml:"region"`
	FirehoseStreamName      string `yaml:"firehoseStreamName"`
	CloudwatchLogGroup      string `yaml:"cloudwatchLogGroup"`
	CloudwatchLogStreamName string `yaml:"cloudwatchLogStreamName"`
}

func NewCfgFrom(from []byte) (*Cfg, error) {
	// parse
	var cfg Cfg
	err := yaml.Unmarshal(from, &cfg)
	if err != nil {
		return nil, fmt.Errorf("error parsing config: %v", err)
	}

	// validate
	if cfg.Output == OutFirehose {
		err := firehoseCfg(&cfg)
		if err != nil {
			return nil, err
		}
	}

	if cfg.Output == OutFile {
		err := fileCfg(&cfg)
		if err != nil {
			return nil, err
		}
	}

	if cfg.Delay == 0 {
		cfg.Delay = defaultDelay
	}

	return &cfg, nil
}

func fileCfg(cfg *Cfg) error {
	if cfg.FileLocation == "" {
		cfg.FileLocation = defaultFileLocation
	}

	return nil
}

func firehoseCfg(cfg *Cfg) error {
	if cfg.FirehoseStreamName == "" {
		return fmt.Errorf("firehose stream name must be non-empty for FIREHOSE output type")
	}

	if cfg.Region == "" {
		slog.Info(fmt.Sprintf("empty AWS region provided, setting to default region %s ", defaultRegion))
		cfg.Region = defaultRegion
	}

	if cfg.Profile == "" {
		slog.Info(fmt.Sprintf("empty AWS Profile provided, setting to default region %s ", defaultProfile))
		cfg.Profile = defaultProfile
	}

	return nil
}
