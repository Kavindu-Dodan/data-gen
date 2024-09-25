package exporters

import (
	"fmt"
	"log/slog"
	"os"
)

type FileExporter struct {
	location string
	shChan   chan struct{}
}

func NewFileExporter(location string) *FileExporter {
	return &FileExporter{
		location: location,
		shChan:   make(chan struct{}),
	}
}

func (f FileExporter) Start(c <-chan []byte, errChan chan error) {
	file, err := os.OpenFile(f.location, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		errChan <- fmt.Errorf("unable to open file %s: %w", f.location, err)
		return
	}

	go func() {
		for {
			select {
			case d := <-c:
				_, err := file.Write(d)
				if err != nil {
					errChan <- fmt.Errorf("unable to write to file %s: %w", f.location, err)
					return
				}
			case <-f.shChan:
				slog.Info("shutting down file exporter")
				return
			}
		}
	}()
}

func (f FileExporter) Stop() {
	close(f.shChan)
}
