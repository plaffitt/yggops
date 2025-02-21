package webhooks

import "net/http"

type Handler interface {
	Validate(request *http.Request) (int, error)
}
