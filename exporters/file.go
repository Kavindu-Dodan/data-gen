package exporters

import (
	"fmt"
	"log/slog"
	"os"

	"data-gen/conf"
)

const defaultLocation = "./out"

type FileExporter struct {
	cfg    *fileCfg
	shChan chan struct{}
}

type fileCfg struct {
	Location string `yaml:"location"`
}

func newDefaultFileCfg() *fileCfg {
	return &fileCfg{
		Location: defaultLocation,
	}
}

func newFileExporter(config *conf.Config) (*FileExporter, error) {
	cfg := newDefaultFileCfg()
	err := config.Output.Conf.Decode(&cfg)
	if err != nil {
		return nil, err
	}

	return &FileExporter{
		cfg:    cfg,
		shChan: make(chan struct{}),
	}, nil
}

func (f FileExporter) Start(c <-chan []byte) <-chan error {
	errChan := make(chan error)

	go func() {
		file, err := os.OpenFile(f.cfg.Location, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			errChan <- fmt.Errorf("unable to open file %s: %w", f.cfg.Location, err)
			return
		}

		for {
			select {
			case d := <-c:
				_, err := file.Write(d)
				if err != nil {
					errChan <- fmt.Errorf("unable to write to file %s: %w", f.cfg.Location, err)
					return
				}
			case <-f.shChan:
				slog.Info("shutting down file exporter")
				return
			}
		}
	}()

	return errChan
}

func (f FileExporter) Stop() {
	close(f.shChan)
}
