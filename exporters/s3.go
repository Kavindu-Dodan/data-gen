package exporters

import (
	"bytes"
	"compress/gzip"
	"context"
	"data-gen/conf"
	"fmt"
	"io"
	"log/slog"
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
	Bucket      string `yaml:"s3Bucket"`
	PathPrefix  string `yaml:"pathPrefix"`
	Compression string `yaml:"compression"`
}

func newDefaultS3Config() s3Config {
	return s3Config{
		PathPrefix: defaultBucketPrefix,
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

func (s *S3BucketExporter) Start(c <-chan *[]byte) <-chan error {
	errChan := make(chan error)

	go func() {
		var content io.Reader
		var encoding string

		for {
			select {
			case data := <-c:
				key := fmt.Sprintf("%s%s", s.cfg.PathPrefix, time.Now().Format("2006-01-02_15-04-05"))

				// check and compress
				if s.cfg.Compression == "gzip" {
					key = key + ".gz"
					compressString, err := gzipCompress(*data)
					if err != nil {
						errChan <- fmt.Errorf("failed to compress s3 object: %w", err)
						return
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
					errChan <- fmt.Errorf("unable to upload to S3 bucket  %s: %w", s.cfg.Bucket, err)
					return
				}
			case <-s.shChan:
				slog.Info("shutting down S3 exporter")
				return
			}
		}
	}()

	return errChan
}

func (s *S3BucketExporter) Stop() {
	close(s.shChan)
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
