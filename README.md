# Data Generator

A simple data generator with export capability to various destinations

## Quick Start

Require Go version 1.24+

- Clone the repository
- Copy `config.sample.yaml` and edit as needed.
- Start the data generator with config file as the only parameter
  `go run cmd/main.go --configFile ./config.yaml`

## Configurations

Given below are the supported configuration options.
Check `config.sample.yaml` for reference.

### Input configurations

Given below are supported input types,

- LOGS : ECS (Elastic Common Schema) formated logs based on zap.
- METRICS: Generate metrics similar to a CloudWatch metrics entry
- ALB : Generate AWS ALB formatted log with some random content

Other input configuration,

- delay : Delay between a data point. Accepts value in acceptable format like 5s (5 Seconds), 10ms (10 milliseconds)

Example:

```yaml
input:
  type: LOGS # Input type LOGS
  delay: 2s  # 2 Seconds between each data point
```

### Output configurations

Given below are supported output types,

- FILE: Output to a file
- FIREHOSE: Output to a Firehose stream
- CLOUDWATCH_LOG: Output to a CloudWatch log group
- S3: Output to a S3 bucket

Sections below provide output specific configurations

#### FILE

- location : Output file location. Default to `./out`

Example:

```yaml
output:
  type: FILE
  config:
    location: "./data"
```

#### S3

- s3Bucket : S3 bucket name (required)
- bucketSeconds: Period between two bucket entries. Default to 120 Seconds
- pathPrefix: Optional prefix for the bucket entry. Default to `logFile-`

Example:

```yaml
output:
  type: S3
  config:
    s3Bucket: "testing-bucket"
    bucketSeconds: 10
    pathPrefix: "datagen"
```

#### FIREHOSE

- stream_name: Firehose stream name (required)

Example:

```yaml
output:
  type: FIREHOSE
  config:
    stream_name: "my-firehose-stream"
```

#### CLOUDWATCH_LOG

- logGroup : Cloudwatch log group name
- logStream : Log group stream name

Example:

```yaml
output:
  type: CLOUDWATCH_LOG
  config:
    logGroup: "MyGroup"
    logStream: "data"
```

### CSP configurations

Currently, this project only support AWS CSP. Given below are available configurations,

- region: Region to use by exporters. Default to us-east-1
- profile: Credential profile to use by exporters. Default to default