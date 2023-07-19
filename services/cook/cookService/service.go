package cookservice

import ()

type CookService interface {
	InitBackground()

	ListenAndServeCookQueue()
}
