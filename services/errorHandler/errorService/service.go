package errorservice

type ErrorService interface {
	InitBackground()
	ListenAndServeCookQueue()
}
