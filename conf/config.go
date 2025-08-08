package conf

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"strconv"
	"strings"
)

const (
	defaultProfile  = "default"
	defaultRegion   = "us-east-1"
	defaultDelay    = "5s"
	defaultBatching = "0s"

	EnvInputType          = "ENV_INPUT_TYPE"
	EnvInputDelay         = "ENV_INPUT_DELAY"
	EnvInputBatching      = "ENV_INPUT_BATCHING"
	EnvInputMaxBatchSize  = "ENV_INPUT_MAX_BATCH_SIZE"
	EnvInputMaxDataPoints = "ENV_INPUT_MAX_DATA_POINTS"

	EnvOutType        = "ENV_OUT_TYPE"
	EnvOutLocation    = "ENV_OUT_LOCATION"
	EnvOutCompression = "ENV_OUT_COMPRESSION"
	EnvOutS3Bucket    = "ENV_OUT_S3_BUCKET"
	EnvOutPathPrefix  = "ENV_OUT_PATH_PREFIX"
	EnvOutStreamName  = "ENV_OUT_STREAM_NAME"
	EnvOutLogGroup    = "ENV_OUT_LOG_GROUP"
	EnvOutLogStream   = "ENV_OUT_LOG_STREAM"

	EnvAWSRegion  = "AWS_REGION"
	EnvAWSProfile = "AWS_PROFILE"
)

type Config struct {
	Input  InputConfig  `yaml:"input"`
	Output OutputConfig `yaml:"output"`
	AWSCfg `yaml:"aws,omitempty"`
}

func newDefaultConfig() *Config {
	return &Config{
		AWSCfg: *newDefaultAWSCfg(),
		Input:  *newDefaultInputConfig(),
	}
}

func (cfg *Config) Print() string {
	sb := strings.Builder{}

	sb.WriteString(fmt.Sprintf("[ Input - %s \t", cfg.Input.Print()))
	sb.WriteString(fmt.Sprintf("Output - %s \t", cfg.Output.Print()))
	sb.WriteString(fmt.Sprintf("AWS - %s ]", cfg.AWSCfg.Print()))

	return strings.TrimSpace(sb.String())
}

type InputConfig struct {
	Type          string    `yaml:"type"`
	Conf          yaml.Node `yaml:"config"`
	Delay         string    `yaml:"delay"`
	Batching      string    `yaml:"batching"`
	MaxSize       int       `yaml:"max_batch_size"`
	MaxDataPoints int64     `yaml:"max_data_points"`
}

func newDefaultInputConfig() *InputConfig {
	return &InputConfig{
		Delay:    defaultDelay,
		Batching: defaultBatching,
	}
}

func (cfg *InputConfig) Print() string {
	sb := strings.Builder{}

	sb.WriteString(fmt.Sprintf("Type: %s, ", cfg.Type))
	sb.WriteString(fmt.Sprintf("Delay: %s ", cfg.Delay))
	sb.WriteString(fmt.Sprintf("Batching: %s ", cfg.Batching))
	sb.WriteString(fmt.Sprintf("MaxBatchBytes: %d ", cfg.MaxSize))
	sb.WriteString(fmt.Sprintf("MaxData points: %d ", cfg.MaxDataPoints))

	return sb.String()
}

type OutputConfig struct {
	Type string    `yaml:"type"`
	Conf yaml.Node `yaml:"config"`
}

func (cfg *OutputConfig) Print() string {
	return fmt.Sprintf("Type: %s", cfg.Type)
}

type AWSCfg struct {
	Profile string `yaml:"profile"`
	Region  string `yaml:"region"`
}

func (c *AWSCfg) Print() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("Profile: %s, ", c.Profile))
	sb.WriteString(fmt.Sprintf("Region: %s", c.Region))

	return sb.String()
}

func newDefaultAWSCfg() *AWSCfg {
	return &AWSCfg{
		Profile: defaultProfile,
		Region:  defaultRegion,
	}
}

func NewConfig(input []byte) (*Config, error) {
	cfg := newDefaultConfig()
	err := yaml.Unmarshal(input, &cfg)
	if err != nil {
		return nil, err
	}

	// override with env variables if defined
	if v := os.Getenv(EnvInputType); v != "" {
		cfg.Input.Type = v
	}
	if v := os.Getenv(EnvInputDelay); v != "" {
		cfg.Input.Delay = v
	}
	if v := os.Getenv(EnvInputBatching); v != "" {
		cfg.Input.Batching = v
	}
	if v := os.Getenv(EnvInputMaxBatchSize); v != "" {
		size, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid value for %s: %w", EnvInputMaxBatchSize, err)
		}
		cfg.Input.MaxSize = size
	}

	if v := os.Getenv(EnvInputMaxDataPoints); v != "" {
		size, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid value for %s: %w", EnvInputMaxBatchSize, err)
		}
		cfg.Input.MaxDataPoints = size
	}

	if v := os.Getenv(EnvOutType); v != "" {
		cfg.Output.Type = v
	}

	if v := os.Getenv(EnvAWSRegion); v != "" {
		cfg.AWSCfg.Region = v
	}

	if v := os.Getenv(EnvAWSProfile); v != "" {
		cfg.AWSCfg.Profile = v
	}

	return cfg, nil
}
