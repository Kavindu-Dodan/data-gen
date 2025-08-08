package exporters

import (
	"fmt"
	"os"

	"data-gen/conf"
)

const defaultLocation = "./out"

type FileExporter struct {
	cfg    *fileCfg
	entry  int
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

	// load env variable overrides if any
	if v := os.Getenv(conf.EnvOutLocation); v != "" {
		cfg.Location = v
	}

	return &FileExporter{
		cfg:    cfg,
		shChan: make(chan struct{}),
	}, nil
}

func (f *FileExporter) send(data *[]byte) error {
	file, err := os.OpenFile(fmt.Sprintf("%s_%d", f.cfg.Location, f.entry), os.O_APPEND|os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("unable to open file %s: %w", f.cfg.Location, err)
	}
	_, err = file.Write(*data)
	if err != nil {
		return fmt.Errorf("unable to write to file %s: %w", f.cfg.Location, err)
	}

	f.entry++
	return nil
}
