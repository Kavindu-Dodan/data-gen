package exporters

type IExport interface {
	Start(c <-chan []byte, errChan chan error)
	Stop()
}
