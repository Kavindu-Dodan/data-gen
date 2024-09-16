package main

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type runner struct {
	shutdown chan struct{}
	log      *zap.Logger
}

func newRunner(log *zap.Logger) *runner {
	sh := make(chan struct{})

	return &runner{
		shutdown: sh,
		log:      log,
	}
}

func (r *runner) start(concurrent int, delay time.Duration) {
	for i := 0; i < concurrent; i++ {
		go func() {
			for {
				select {
				case <-time.After(delay):
					r.log.Info(fmt.Sprintf("log entry: %s", uuid.NewString()))
				case _ = <-r.shutdown:
					r.log.Info("Shutting down")
					return
				}
			}
		}()
	}
}

func (r *runner) end() {
	close(r.shutdown)
}
