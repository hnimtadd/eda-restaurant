package tableservice

type TableService interface {
	Serve(msg []byte) error
	Clean(msg []byte) error

	InitBackground()

	ListenAndServeCookQueue()
}
