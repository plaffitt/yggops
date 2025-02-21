package webhooks

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
)

var bearerRegexp = regexp.MustCompile("^Bearer .+")
var errInvalidToken = errors.New("invalid token")

type Generic struct {
	secret string
}

func NewGeneric(secret string) (*Generic, error) {
	return &Generic{
		secret: secret,
	}, nil
}

func (m *Generic) Validate(request *http.Request) (int, error) {
	fmt.Printf("Received generic webhook event\n")

	authorization := request.Header.Get("Authorization")
	if bearerRegexp.Match([]byte(authorization)) {
		token := authorization[len("Bearer "):]
		if token != m.secret {
			return http.StatusUnauthorized, errInvalidToken
		}
	}

	return http.StatusOK, nil
}
