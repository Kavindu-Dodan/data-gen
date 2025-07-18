package exporters

import (
	"context"
	"data-gen/conf"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
)

const defaultBucketPrefix = "logFile-"

type S3BucketExporter struct {
	cfg    s3Config
	client *awss3.Client
	shChan chan struct{}
}

type s3Config struct {
	Bucket        string `yaml:"s3Bucket"`
	BucketSeconds int64  `yaml:"bucketSeconds"`
	PathPrefix    string `yaml:"pathPrefix"`
}

func newDefaultS3Config() s3Config {
	return s3Config{
		BucketSeconds: 120,
		PathPrefix:    defaultBucketPrefix,
	}
}

func newS3BucketExporter(ctx context.Context, conf *conf.Config) (*S3BucketExporter, error) {
	cfg := newDefaultS3Config()
	err := conf.Output.Conf.Decode(&cfg)
	if err != nil {
		return nil, err
	}

	if cfg.Bucket == "" {
		return nil, fmt.Errorf("s3 Bucket Name is empty, please configure and try again")
	}

	loadedAwsConfig, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(conf.AWSCfg.Profile), config.WithRegion(conf.AWSCfg.Region))
	if err != nil {
		return nil, fmt.Errorf("failed to load default aws config: %w", err)
	}

	return &S3BucketExporter{
		cfg:    cfg,
		client: awss3.NewFromConfig(loadedAwsConfig),
		shChan: make(chan struct{}),
	}, nil
}

func (s *S3BucketExporter) Start(c <-chan []byte, errChan chan error) {
	go func() {
		lastStart := time.Now()
		sb := strings.Builder{}
		bucketingPeriod := time.Duration(s.cfg.BucketSeconds) * time.Second
		for {
			select {
			case d := <-c:
				sb.Write(d)

				if time.Since(lastStart) > bucketingPeriod {
					// upload to S3
					key := fmt.Sprintf("%s%s", s.cfg.PathPrefix, time.Now().Format("2006-01-02_15-04-05"))

					input := awss3.PutObjectInput{
						Bucket: &s.cfg.Bucket,
						Key:    &key,
						Body:   strings.NewReader(sb.String()),
					}

					_, err := s.client.PutObject(context.Background(), &input)
					if err != nil {
						errChan <- fmt.Errorf("unable to upload to S3 bucket  %s: %w", s.cfg.Bucket, err)
						return
					}

					// reset all
					lastStart = time.Now()
					sb.Reset()
				}
			case <-s.shChan:
				slog.Info("shutting down S3 exporter")
				return
			}
		}
	}()
}

func (s *S3BucketExporter) Stop() {
	close(s.shChan)
}
