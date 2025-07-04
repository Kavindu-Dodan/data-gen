package conf

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"log/slog"
)

const (
	OutFirehose       = "FIREHOSE"
	OutFile           = "FILE"
	OutCloudwatchLogs = "CLOUDWATCH_LOG"
	OutS3Bucket       = "S3"
	TypeLogs          = "LOGS"
	TypeMetrics       = "METRICS"

	defaultBucketPrefix = "logFile-"
	defaultProfile      = "default"
	defaultRegion       = "us-east-1"
	defaultFileLocation = "./out"
	defaultDelay        = "5s"
)

type Cfg struct {
	Type         string `yaml:"type"`
	Output       string `yaml:"output"`
	Delay        string `yaml:"delay"`
	FileLocation string `yaml:"file_location,omitempty"`
	AWSCfg       `yaml:"aws,omitempty"`
}

type AWSCfg struct {
	Profile                 string `yaml:"profile"`
	Region                  string `yaml:"region"`
	S3Bucket                string `yaml:"s3Bucket"`
	BucketPrefix            string `yaml:"bucketPrefix"`
	BucketPeriodSeconds     int    `yaml:"bucketSeconds"`
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

	if cfg.Output == OutS3Bucket {
		err := s3Cfg(&cfg)
		if err != nil {
			return nil, err
		}
	}

	if cfg.Delay == "" {
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

func s3Cfg(cfg *Cfg) error {
	if cfg.S3Bucket == "" {
		return errors.New("s3Bucket is required")
	}

	if cfg.BucketPrefix == "" {
		slog.Info(fmt.Sprintf("using default bucket prefix '%s' ", defaultBucketPrefix))
		cfg.BucketPrefix = defaultBucketPrefix
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
