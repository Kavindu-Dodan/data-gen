package exporters

import (
	"context"
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
	shChan   chan struct{}
}

// eventHubCfg specifies the Event Hub namespace and name.
type eventHubCfg struct {
	Namespace        string `yaml:"namespace"`
	EventHubName     string `yaml:"event_hub_name"`
	ConnectionString string `yaml:"connection_string"`
}

func newEventHubExporter(ctx context.Context, c *conf.Config) (*EventHubExporter, error) {
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
		shChan:   make(chan struct{}),
	}, nil
}

func (e *EventHubExporter) send(data *[]byte) error {
	batch, err := e.producer.NewEventDataBatch(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("failed to create event batch: %w", err)
	}

	event := &azeventhubs.EventData{
		Body: *data,
	}

	err = batch.AddEventData(event, nil)
	if err != nil {
		return fmt.Errorf("failed to add event to batch: %w", err)
	}

	err = e.producer.SendEventDataBatch(context.Background(), batch, nil)
	if err != nil {
		return fmt.Errorf("failed to send event to event hub %s: %w", e.cfg.EventHubName, err)
	}

	return nil
}

func (e *EventHubExporter) stop() {
	close(e.shChan)
	if e.producer != nil {
		e.producer.Close(context.Background())
	}
}
