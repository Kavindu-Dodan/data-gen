package exporters

import (
	"context"
	"data-gen/conf"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"log/slog"
	"strings"
	"time"
)

const defaultBucketingPeriod = 2 * time.Minute

type S3BucketExporter struct {
	bucketName   string
	bucketPrefix string
	timePeriod   time.Duration
	client       *s3.Client
	shChan       chan struct{}
}

func NewS3BucketExporter(ctx context.Context, awsConfig conf.AWSCfg) (*S3BucketExporter, error) {
	if awsConfig.S3Bucket == "" {
		return nil, fmt.Errorf("s3 Bucket Name is empty, please configure and try again")
	}

	loadedAwsConfig, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(awsConfig.Profile), config.WithRegion(awsConfig.Region))
	if err != nil {
		return nil, fmt.Errorf("failed to load default aws config: %w", err)
	}

	client := s3.NewFromConfig(loadedAwsConfig)

	var bucketing = defaultBucketingPeriod
	if awsConfig.BucketPeriodSeconds != 0 {
		bucketing = time.Duration(awsConfig.BucketPeriodSeconds) * time.Second
	}

	return &S3BucketExporter{
		bucketName:   awsConfig.S3Bucket,
		bucketPrefix: awsConfig.BucketPrefix,
		timePeriod:   bucketing,
		client:       client,
		shChan:       make(chan struct{}),
	}, nil
}

func (s *S3BucketExporter) Start(c <-chan []byte, errChan chan error) {
	go func() {
		lastStart := time.Now()
		sb := strings.Builder{}
		for {
			select {
			case d := <-c:
				sb.Write(d)

				if time.Since(lastStart) > s.timePeriod {
					// upload to S3
					key := fmt.Sprintf("%s%s", s.bucketPrefix, time.Now().Format("2006-01-02_15-04-05"))

					input := s3.PutObjectInput{
						Bucket: &s.bucketName,
						Key:    &key,
						Body:   strings.NewReader(sb.String()),
					}

					_, err := s.client.PutObject(context.Background(), &input)
					if err != nil {
						errChan <- fmt.Errorf("unable to upload to S3 bucket  %s: %w", s.bucketName, err)
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
