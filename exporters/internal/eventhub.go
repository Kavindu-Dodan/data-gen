package internal

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"

	"data-gen/conf"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs/v2"
)

// EventHubExporter sends generated data to Azure Event Hubs.
type EventHubExporter struct {
	cfg      eventHubCfg
	producer *azeventhubs.ProducerClient
}

// eventHubCfg specifies the Event Hub namespace and name.
type eventHubCfg struct {
	Namespace        string `yaml:"namespace"`
	EventHubName     string `yaml:"event_hub_name"`
	ConnectionString string `yaml:"connection_string"`
}

func NewEventHubExporter(ctx context.Context, c *conf.Config) (*EventHubExporter, error) {
	var cfg eventHubCfg
	err := c.Output.Conf.Decode(&cfg)
	if err != nil {
		return nil, err
	}

	// load env variable overrides if any
	if v := os.Getenv(conf.EnvOutEventHubNamespace); v != "" {
		cfg.Namespace = v
	}
	if v := os.Getenv(conf.EnvOutEventHubName); v != "" {
		cfg.EventHubName = v
	}
	if v := os.Getenv(conf.EnvOutEventHubConnectionString); v != "" {
		cfg.ConnectionString = v
	}

	var producer *azeventhubs.ProducerClient

	// Support both connection string and Azure AD authentication
	if cfg.ConnectionString != "" {
		producer, err = azeventhubs.NewProducerClientFromConnectionString(cfg.ConnectionString, cfg.EventHubName, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create event hub producer from connection string: %w", err)
		}
	} else {
		if cfg.Namespace == "" || cfg.EventHubName == "" {
			return nil, fmt.Errorf("event hub namespace and name must be specified for output type %s", c.Output.Type)
		}

		credential, err := azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create azure credential: %w", err)
		}

		fullyQualifiedNamespace := fmt.Sprintf("%s.servicebus.windows.net", cfg.Namespace)
		producer, err = azeventhubs.NewProducerClient(fullyQualifiedNamespace, cfg.EventHubName, credential, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create event hub producer: %w", err)
		}
	}

	return &EventHubExporter{
		cfg:      cfg,
		producer: producer,
	}, nil
}

// Send delivers data to Event Hubs. The generator accumulates multiple
// {"records": [...]} objects separated by newlines (NDJSON). Each line is
// sent as its own EventData so that no individual message exceeds the 1 MB
// Event Hubs limit. Lines are packed into batches; when a batch is full the
// SDK returns ErrEventDataTooLarge, at which point the batch is flushed and a
// new one is started for the remaining lines.
func (e *EventHubExporter) Send(data *[]byte) error {
	ctx := context.Background()
	lines := bytes.Split(bytes.TrimRight(*data, "\n"), []byte("\n"))

	batch, err := e.producer.NewEventDataBatch(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to create event batch: %w", err)
	}

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		batch, err = e.addEventToBatch(ctx, batch, line)
		if err != nil {
			return err
		}
	}

	return e.flushBatch(ctx, batch)
}

// addEventToBatch adds a line to the batch, flushing and creating a new batch if needed.
func (e *EventHubExporter) addEventToBatch(ctx context.Context, batch *azeventhubs.EventDataBatch, line []byte) (*azeventhubs.EventDataBatch, error) {
	err := batch.AddEventData(&azeventhubs.EventData{Body: line}, nil)
	if err == nil {
		return batch, nil
	}

	if !errors.Is(err, azeventhubs.ErrEventDataTooLarge) {
		return nil, fmt.Errorf("failed to add event to batch: %w", err)
	}

	// Single record exceeds limit — cannot proceed
	if batch.NumEvents() == 0 {
		return nil, fmt.Errorf("single record (%d bytes) exceeds Event Hubs message size limit", len(line))
	}

	// Flush current batch and create a new one
	if err = e.producer.SendEventDataBatch(ctx, batch, nil); err != nil {
		return nil, fmt.Errorf("failed to send event batch to %s: %w", e.cfg.EventHubName, err)
	}

	newBatch, err := e.producer.NewEventDataBatch(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create event batch: %w", err)
	}

	// Retry the line in the fresh batch
	if err = newBatch.AddEventData(&azeventhubs.EventData{Body: line}, nil); err != nil {
		return nil, fmt.Errorf("failed to add event to fresh batch: %w", err)
	}

	return newBatch, nil
}

// flushBatch sends any remaining events in the batch.
func (e *EventHubExporter) flushBatch(ctx context.Context, batch *azeventhubs.EventDataBatch) error {
	if batch.NumEvents() == 0 {
		return nil
	}

	if err := e.producer.SendEventDataBatch(ctx, batch, nil); err != nil {
		return fmt.Errorf("failed to send event batch to %s: %w", e.cfg.EventHubName, err)
	}

	return nil
}
