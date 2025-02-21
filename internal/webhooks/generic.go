package webhooks

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
)

var bearerRegexp = regexp.MustCompile("^Bearer .+")
var errInvalidToken = errors.New("invalid token")
var errMissingToken = errors.New("missing token")

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
	} else {
		return http.StatusUnauthorized, errMissingToken
	}

	return http.StatusOK, nil
}
