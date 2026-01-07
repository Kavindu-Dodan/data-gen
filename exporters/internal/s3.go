package internal

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"data-gen/conf"

	"github.com/aws/aws-sdk-go-v2/config"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	defaultBucketPrefix = "logFile-"
	arnPrefix           = "arn:aws:s3:::"
)

// S3BucketExporter uploads generated data to AWS S3 with optional gzip compression.
type S3BucketExporter struct {
	cfg    s3Config
	client *awss3.Client
}

// s3Config defines S3 bucket, path, and compression settings.
type s3Config struct {
	Bucket      string `yaml:"s3_bucket"`
	PathPrefix  string `yaml:"path_prefix"`
	Compression string `yaml:"compression"`
}

func newDefaultS3Config() s3Config {
	return s3Config{
		PathPrefix: defaultBucketPrefix,
	}
}

func NewS3BucketExporter(ctx context.Context, configuration *conf.Config) (*S3BucketExporter, error) {
	cfg := newDefaultS3Config()
	err := configuration.Output.Conf.Decode(&cfg)
	if err != nil {
		return nil, err
	}

	// load env variable overrides if any
	if v := os.Getenv(conf.EnvOutS3Bucket); v != "" {
		cfg.Bucket = v
	}

	if v := os.Getenv(conf.EnvOutPathPrefix); v != "" {
		cfg.PathPrefix = v
	}

	if v := os.Getenv(conf.EnvOutCompression); v != "" {
		cfg.Compression = v
	}

	if cfg.Bucket == "" {
		return nil, fmt.Errorf("s3 Bucket Name is empty, please configure and try again")
	}

	cfg.Bucket = toBucketName(cfg.Bucket)

	loadedAwsConfig, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(configuration.Profile), config.WithRegion(configuration.Region))
	if err != nil {
		return nil, fmt.Errorf("failed to load default aws config: %w", err)
	}

	return &S3BucketExporter{
		cfg:    cfg,
		client: awss3.NewFromConfig(loadedAwsConfig),
	}, nil
}

func (s *S3BucketExporter) Send(data *[]byte) error {
	var content io.Reader
	var encoding string

	key := fmt.Sprintf("%s%s", s.cfg.PathPrefix, time.Now().Format("2006-01-02_15-04-05"))

	// check and compress
	if s.cfg.Compression == "gzip" {
		key = key + ".gz"
		compressString, err := gzipCompress(*data)
		if err != nil {
			return fmt.Errorf("failed to compress s3 object: %w", err)
		}

		content = bytes.NewReader(compressString)
		encoding = "gzip"
	} else {
		content = bytes.NewReader(*data)
	}

	input := awss3.PutObjectInput{
		Bucket:          &s.cfg.Bucket,
		Key:             &key,
		Body:            content,
		ContentEncoding: &encoding,
	}

	_, err := s.client.PutObject(context.Background(), &input)
	if err != nil {
		return fmt.Errorf("unable to upload to S3 bucket  %s: %w", s.cfg.Bucket, err)
	}

	return nil
}

func gzipCompress(input []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	_, err := gz.Write(input)
	if err != nil {
		return nil, err
	}

	err = gz.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func toBucketName(bucket string) string {
	if strings.HasPrefix(bucket, arnPrefix) {
		return strings.TrimPrefix(bucket, arnPrefix)
	}

	return bucket
}
