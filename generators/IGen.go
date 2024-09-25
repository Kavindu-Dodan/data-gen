package generators

import "time"

type IGen interface {
	Start(delay time.Duration, errChan chan<- error) <-chan []byte
	Stop()
}
