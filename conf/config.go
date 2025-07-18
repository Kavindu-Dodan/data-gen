package conf

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"strings"
)

const (
	defaultProfile = "default"
	defaultRegion  = "us-east-1"
	defaultDelay   = "5s"
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
	Type  string    `yaml:"type"`
	Conf  yaml.Node `yaml:"config"`
	Delay string    `yaml:"delay"`
}

func newDefaultInputConfig() *InputConfig {
	return &InputConfig{
		Delay: defaultDelay,
	}
}

func (cfg *InputConfig) Print() string {
	sb := strings.Builder{}

	sb.WriteString(fmt.Sprintf("Type: %s, ", cfg.Type))
	sb.WriteString(fmt.Sprintf("Delay: %s", cfg.Delay))

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

	return cfg, nil
}
