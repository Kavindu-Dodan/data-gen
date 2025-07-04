# Data Generator

A simple data generator for easy validations of data ingesting.

## Quick Start

Require Go version 1.24+

- Clone the repository
- Copy `config.sample.yaml` and edit as needed.
- Start the data generator with config file as the only parameter
  `go run main.go --configFile ./config.yaml`

## Configurations

Check `config.sample.yaml` as a reference configuration.

### Sources

#### Logs

ECS (Elastic Common Schema) formated logs based on zap.
Check [ecs-logging-go-zap](https://github.com/elastic/ecs-logging-go-zap) for adoption

#### Metrics

Generate metrics similar to a CloudWatch metrics entry. Limited support and tobe improved

### Exporters

> [!IMPORTANT]  
> AWS specific exporters require aws.profile to configure with a valid credential profile name.
> And make sure you have sufficient permissions to use the exporter with these credentials.

#### File

Generates a file with new line as data point delimiter.

Available options,

- `file_location` : Optional file location to write the output to. Default to `./out`


#### S3

Write generated data to a S3 bucket.

Available options,

- `s3Bucket` : Bucket name (required)
- `bucketSeconds`: Period between two bucket entries. Optional and default to 120seconds
- `bucketPrefix`: Optional prefix for the entry. Default to `logFile-`

#### FIREHOSE

Push data to an AWS firehose data stream. 

Available options,

- `firehoseStreamName` : firehose stream name (required)

#### CLOUDWATCH_LOG

Push data to AWS CloudWatch log group or stream name.

Available options,

- `cloudwatchLogGroup` : Cloudwatch log group name
- `cloudwatchLogStreamName` : Define a specific log stream name

> [!IMPORTANT]  
> One of these options must be set to use the exporter

