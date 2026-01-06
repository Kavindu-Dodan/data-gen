package conf

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	defaultProfile     = "default"
	defaultRegion      = "us-east-1"
	defaultDelay       = "5s"
	defaultBatching    = "0s"
	defaultMaxDuration = "0s"

	EnvInputType          = "ENV_INPUT_TYPE"
	EnvInputDelay         = "ENV_INPUT_DELAY"
	EnvInputBatching      = "ENV_INPUT_BATCHING"
	EnvInputMaxBatchSize  = "ENV_INPUT_MAX_BATCH_SIZE"
	EnvInputMaxDataPoints = "ENV_INPUT_MAX_DATA_POINTS"
	EnvMaxRuntime         = "ENV_INPUT_MAX_RUNTIME"

	EnvOutType        = "ENV_OUT_TYPE"
	EnvOutLocation    = "ENV_OUT_LOCATION"
	EnvOutCompression = "ENV_OUT_COMPRESSION"
	EnvOutS3Bucket    = "ENV_OUT_S3_BUCKET"
	EnvOutPathPrefix  = "ENV_OUT_PATH_PREFIX"
	EnvOutStreamName  = "ENV_OUT_STREAM_NAME"
	EnvOutLogGroup    = "ENV_OUT_LOG_GROUP"
	EnvOutLogStream   = "ENV_OUT_LOG_STREAM"

	EnvOutEventHubNamespace        = "ENV_OUT_EVENTHUB_NAMESPACE"
	EnvOutEventHubName             = "ENV_OUT_EVENTHUB_NAME"
	EnvOutEventHubConnectionString = "ENV_OUT_EVENTHUB_CONNECTION_STRING"

	EnvAWSRegion  = "AWS_REGION"
	EnvAWSProfile = "AWS_PROFILE"
)

// Config holds the complete configuration for the data generator including input, output, and AWS settings.
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

	sb.WriteString("  Input:\n")
	sb.WriteString("    " + cfg.Input.Print() + "\n")
	sb.WriteString("  Output:\n")
	sb.WriteString("    " + cfg.Output.Print())

	// Only include AWS config if output type uses AWS
	if cfg.usesAWS() {
		sb.WriteString("\n  AWS:\n")
		sb.WriteString("    " + cfg.AWSCfg.Print())
	}

	return sb.String()
}

// usesAWS returns true if the output type or input type requires AWS configuration
func (cfg *Config) usesAWS() bool {
	// Check if output type requires AWS
	switch cfg.Output.Type {
	case "S3", "FIREHOSE", "CLOUDWATCH_LOG":
		return true
	}

	// Check if input type is AWS-specific (may need AWS config for region/profile context)
	switch cfg.Input.Type {
	case "ALB", "NLB", "VPC", "WAF", "CLOUDTRAIL":
		return true
	}

	return false
}

// InputConfig defines the data generation behavior including type, timing, and limits.
type InputConfig struct {
	Type          string    `yaml:"type"`
	Conf          yaml.Node `yaml:"config"`
	Delay         string    `yaml:"delay"`
	Batching      string    `yaml:"batching"`
	MaxSize       int       `yaml:"max_batch_size"`
	MaxDataPoints int64     `yaml:"max_data_points"`
	MaxRunTime    string    `yaml:"max_runtime"`
}

func newDefaultInputConfig() *InputConfig {
	return &InputConfig{
		Delay:      defaultDelay,
		Batching:   defaultBatching,
		MaxRunTime: defaultMaxDuration,
	}
}

func (cfg *InputConfig) Print() string {
	sb := strings.Builder{}

	sb.WriteString(fmt.Sprintf("Type: %s", cfg.Type))
	if cfg.Delay != "" && cfg.Delay != defaultDelay {
		sb.WriteString(fmt.Sprintf(", Delay: %s", cfg.Delay))
	}
	if cfg.Batching != "" && cfg.Batching != defaultBatching {
		sb.WriteString(fmt.Sprintf(", Batching: %s", cfg.Batching))
	}
	if cfg.MaxSize > 0 {
		sb.WriteString(fmt.Sprintf(", Max Batch Size: %d bytes", cfg.MaxSize))
	}
	if cfg.MaxDataPoints > 0 {
		sb.WriteString(fmt.Sprintf(", Max Data Points: %d", cfg.MaxDataPoints))
	}
	if cfg.MaxRunTime != "" && cfg.MaxRunTime != defaultMaxDuration {
		sb.WriteString(fmt.Sprintf(", Max Runtime: %s", cfg.MaxRunTime))
	}

	return sb.String()
}

// OutputConfig specifies where and how to export generated data.
type OutputConfig struct {
	Type string    `yaml:"type"`
	Conf yaml.Node `yaml:"config"`
}

func (cfg *OutputConfig) Print() string {
	return fmt.Sprintf("Type: %s", cfg.Type)
}

// AWSCfg contains AWS-specific configuration for credential profile and region.
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
	if v := os.Getenv(EnvMaxRuntime); v != "" {
		cfg.Input.MaxRunTime = v
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
