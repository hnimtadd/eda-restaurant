package transport

type StorageTransport interface {
	Run(port string) error
}
