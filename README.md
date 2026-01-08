# Synthetic Data Generator

A flexible synthetic data generator in Go, designed to produce realistic logs and metrics for testing and development.

```mermaid
---
config:
  look: handDrawn
  theme: default
---
flowchart LR
 subgraph Config["Config Source"]
        YAML["config.yaml"]
        ENV["Env Vars"]
  end
 subgraph Tool["Synthetic Data Generator"]
    direction TB
        G["<b>Generator</b><br>Logs &amp; Metrics"]
        P["<b>Processor</b><br>Throttling &amp; Batching"]
        E["<b>Exporter Engine</b>"]
  end
 subgraph AWS["AWS Services"]
        S3["S3 Bucket"]
        CW["CloudWatch Logs"]
        FH["Kinesis Firehose"]
  end
 subgraph Azure["Azure"]
        EH["Event Hub"]
  end
 subgraph Local["Local Storage"]
        LF["Local File"]
  end
    G --> P
    P --> E
    Config --> Tool
    E --> AWS & Azure & Local

    style E fill:#d1e7ff,stroke:#004a99,stroke-width:2px
    style Tool fill:#f9f9f9,stroke:#D50000,stroke-width:2px,color:#000000
```

## Quick Start

- Clone the repository (alternatively download a release binary
  from [releases](https://github.com/Kavindu-Dodan/data-gen/releases))
- Copy following to `config.yaml`
  ```yaml
  input:
    type: LOGS
    delay: 500ms
    batching: 2s
    max_batch_size: 500_000
    max_runtime: 5s
  output:
    type: DEBUG
    config:
      verbosity: detailed
  ```
- Run data generator with one of the following commands,

  From source,

  `go run cmd/main.go --config ./config.yaml`

  Using a release binary,

  `./dataGenerator_darwin_arm64 --config ./config.yaml`

Observe the terminal for generated logs.

> [!TIP]
> Use `--debug` flag for debug level logs.

## Configurations

Given below are the supported configuration options.
Check `config.sample.yaml` for reference.

### Input configurations

Given below are supported input types and their related environment variable overrides,

| YAML Property     | Environment Variable        | Default                  | Description                                                                                                                                                                                                                                                                              |
|-------------------|-----------------------------|--------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `type`            | `ENV_INPUT_TYPE`            | - (required from user)   | Specifies the input data type (eg, `LOGS`, `METRICS`, `ALB`, `NLB`, `VPC`, `CLOUDTRAIL`, `WAF`).                                                                                                                                                                                         |
| `delay`           | `ENV_INPUT_DELAY`           | 1s                       | Delay between a data point. Accepts value in format like `5s` (5 seconds), `10ms` (10 milliseconds).                                                                                                                                                                                     |
| `batching`        | `ENV_INPUT_BATCHING`        | 0 (no batch duration)    | [Batching] Set time delay between data batches. Accepts a time value similar to delay. Note: Batching is most effective with bulk ingest endpoints like S3 and Firehose. For CloudWatch Logs, batching may not be suitable as it concatenates multiple log entries into single messages. |
| `max_batch_size`  | `ENV_INPUT_MAX_BATCH_SIZE`  | 0 (no max bytes)         | [Batching] Set maximum **byte** size for a batch.                                                                                                                                                                                                                                        |
| `max_batch_count` | `ENV_INPUT_MAX_BATCH_COUNT` | 0 (no max element count) | [Batching] Set maximum **element** count for a batch.                                                                                                                                                                                                                                    |                                                                                                                                                                                                                                                                               |
| `max_data_points` | `ENV_INPUT_MAX_DATA_POINTS` | - (no limit)             | [Runtime] Set maximum amount of data points to generate during runtime.                                                                                                                                                                                                                  |
| `max_runtime`     | `ENV_INPUT_MAX_RUNTIME`     | - (no max runtime)       | [Runtime] Set the duration for full load generation runtime.                                                                                                                                                                                                                             |

> [!NOTE]
> You must define one of the terminal conditions for batching ([Batching]) or runtime ([Runtime]).
> For example, define one of `max_data_points` or `max_runtime` when  `max_batch_size`, `max_batch_count` or `batching` duration is not set.
> On the other hand, for example when batching duration is set, you can run load generator indefinitely without defining runtime ([Runtime]) limits.

Given below are supported `type` values for input,

| Log Type              | Description                                                                                             |
|-----------------------|---------------------------------------------------------------------------------------------------------|
| `ALB`                 | Generate AWS ALB formatted logs with some random content                                                |
| `NLB`                 | Generate AWS NLB formatted logs with some random content                                                |
| `VPC`                 | Generate AWS VPC formatted logs with randomized content                                                 |
| `CLOUDTRAIL`          | Generate AWS CloudTrail formatted logs with randomized content. Data is generated for AWS S3 Data Event |
| `WAF`                 | Generate AWS WAF formatted logs with randomized content                                                 |
| `AZURE_RESOURCE_LOGS` | Generate Azure Resource logs with randomized content                                                    |
| `LOGS`                | ECS (Elastic Common Schema) formatted logs based on zap                                                 |
| `METRICS`             | Generate metrics similar to a CloudWatch metrics entry                                                  |

Example:

```yaml
input:
  type: LOGS             # Input type LOGS
  delay: 500ms           # 500 milliseconds between each data point
  batching: 10s          # Emit generated data batched within 10 seconds
  max_batch_size: 10000  # Limit maximum batch size to 10,000 bytes. The output is capped at 1000 bytes/second max
  max_data_points: 10000 # Exit input after generating 10,000 data points
```

> [!TIP]
> When max_batch_size is reached, elapsed time for batching will be considered before generating new data

### Output configurations

Given below are supported output configurations and their related environment variable overrides,

| YAML Property         | Environment Variable          | Description                                                                |
|-----------------------|-------------------------------|----------------------------------------------------------------------------|
| `type`                | `ENV_OUT_TYPE`                | Accepts the output type (see table below)                                  |
| `wait_for_completion` | `ENV_OUT_WAIT_FOR_COMPLETION` | Wait for output exports to complete when shutting down. Default is `true`. |

Given below are supported output types,

| Output Type      | Description                        |
|------------------|------------------------------------|
| `FIREHOSE`       | Export to AWS Firehose stream      |
| `CLOUDWATCH_LOG` | Export to AWS CloudWatch log group |
| `S3`             | Export to AWS S3 bucket            |
| `EVENTHUB`       | Export to Azure Event hub          |
| `FILE`           | Export to a file                   |

Sections below provide output specific configurations

#### S3

| YAML Property | Environment Variable  | Description                                                  |
|---------------|-----------------------|--------------------------------------------------------------|
| `s3_bucket`   | `ENV_OUT_S3_BUCKET`   | S3 bucket name (required).                                   |
| `compression` | `ENV_OUT_COMPRESSION` | To compress or not the output. Currently supports `gzip`.    |
| `path_prefix` | `ENV_OUT_PATH_PREFIX` | Optional prefix for the bucket entry. Default to `logFile-`. |

Example:

```yaml
output:
  type: S3
  config:
    s3_bucket: "testing-bucket"
    compression: gzip
    path_prefix: "datagen"
```

#### FIREHOSE

| YAML Property | Environment Variable  | Description                      |
|---------------|-----------------------|----------------------------------|
| `stream_name` | `ENV_OUT_STREAM_NAME` | Firehose stream name (required). |

Example:

```yaml
output:
  type: FIREHOSE
  config:
    stream_name: "my-firehose-stream"
```

#### CLOUDWATCH_LOG

| YAML Property | Environment Variable | Description                |
|---------------|----------------------|----------------------------|
| `log_group`   | `ENV_OUT_LOG_GROUP`  | CloudWatch log group name. |
| `log_stream`  | `ENV_OUT_LOG_STREAM` | Log group stream name.     |

Example:

```yaml
output:
  type: CLOUDWATCH_LOG
  config:
    logGroup: "MyGroup"
    logStream: "data"
```

> [!NOTE]
> CloudWatch Logs API (`PutLogEvents`) is optimized for single log messages per API call. When batching is enabled, multiple log entries are concatenated into a single message, which may not be ideal for log analysis and searching in CloudWatch. For CloudWatch destinations, consider setting `batching: 0s` (no batching) or using a small delay without batching. Batching is more suitable for bulk ingest endpoints like S3 and Firehose.


#### EVENTHUB

| YAML Property       | Environment Variable                 | Description                                     |
|---------------------|--------------------------------------|-------------------------------------------------|
| `connection_string` | `ENV_OUT_EVENTHUB_CONNECTION_STRING` | Connection string for the Event Hub namespace   |
| `event_hub_name`    | `ENV_OUT_EVENTHUB_NAME`              | Event hub entity name to export genereated data |
| `namespace`         | `ENV_OUT_EVENTHUB_NAMESPACE`         | Event hub namespace                             |

Example:

Using with connection string & event hub name:

```yaml
output:
  type: EVENTHUB
  config:
    connection_string: "Endpoint=sb:xxxxxx"
    event_hub_name: "<event_hub_name>"
```

Using with connection string & event hub name (requires IAM role assigned to the role running this application):

```yaml
output:
  type: EVENTHUB
  config:
    event_hub_name: "<event_hub_name>"
    namespace: "<namespace>"
```

#### FILE

| YAML Property | Environment Variable | Description                                                                                                                |
|---------------|----------------------|----------------------------------------------------------------------------------------------------------------------------|
| `location`    | `ENV_OUT_LOCATION`   | Output file location. Default to `./out`. When batching, file suffix will increment with numbers (e.g., `out_0`, `out_2`). |

Example:

```yaml
output:
  type: FILE
  config:
    location: "./data"
```

### Cloud provider configurations

#### AWS

| YAML Property | Environment Variable | Description                                                   |
|---------------|----------------------|---------------------------------------------------------------|
| `region`      | `AWS_REGION`         | Region to use by exporters. Default is `us-east-1`.           |
| `profile`     | `AWS_PROFILE`        | Credential profile to use by exporters. Default is `default`. |

Example:

```yaml
aws:
  region: "us-east-1"
  profile: "default"
```

## Examples

### 1. Continuous Log Generation to a S3 bucket

Generate ECS-formatted log every 2s, batch them in 10 seconds and forward to S3 bucket.
Default delay is 1s between data points.

```yaml
input:
  type: LOGS
  delay: 2s
  batching: 10s
output:
  type: s3
  config:
    s3_bucket: "testing-bucket"
```

### 2. Continuous Log Generation with max batch size based on the batch size based on bytes

Generate ALB logs. 
No delay between data points (continuous data generating).
Limit batching to 10 seconds and max batch size is set to 10MB. This translates to ~1 MB/second data load.
S3 files will be in `gzip` format.

```yaml
input:
  type: ALB
  delay: 0s
  batching: 10s
  max_batch_size: 10_000_000  # 10 MB max bytes per batch
output:
  type: s3
  config:
    s3_bucket: "testing-bucket"
    compression: "gzip"
```

### 3. Continuous Log Generation with max batch size based on the number of elements in a batch

Generate WAF logs.
Delay of 100ms between data points.
Limit batching to max 1000 elements per batch.
Send to S3 in `gzip` format.

```yaml
input:
  type: WAF
  delay: 100ms
  max_batch_count: 1000  # Max 1000 elements per batch
output:
  type: s3
  config:
    s3_bucket: "testing-bucket"
    compression: "gzip"
```


### 4. Limit total data points & exit

Generate VPC logs and limit to 200 data points. Then upload it to S3 in `gzip` format.

```yaml
input:
  type: VPC
  delay: 1s
  max_data_points: 200  # Limit to 200 data points total of the runtime
output:
  type: s3
  config:
    s3_bucket: "testing-bucket"
    compression: "gzip"
```

### 5. Limit runtime for 5 minutes & exit

Generate CLOUDTRAIL logs and limit to generator runtime of 5 minutes.

```yaml
input:
  type: CLOUDTRAIL
  delay: 10us       # 10 microseconds between data points
  batching: 10s
  max_runtime: 5m   # 5 minutes
output:
  type: s3
  config:
    s3_bucket: "testing-bucket"
    compression: "gzip"
```