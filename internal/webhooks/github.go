package webhooks

import (
	"fmt"
	"net/http"
	"slices"

	"github.com/go-playground/webhooks/v6/github"
)

type Github struct {
	github *github.Webhook
	event  string
}

func NewGithub(secret string, event string) (*Github, error) {
	github, err := github.New(github.Options.Secret(secret))
	if err != nil {
		return nil, err
	}

	return &Github{
		github: github,
		event:  event,
	}, nil
}

func (g *Github) Validate(request *http.Request) (int, error) {
	event := request.Header.Get("X-GitHub-Event")
	fmt.Printf("Received Github \"%s\" event\n", event)

	_, err := g.github.Parse(request, github.Event(g.event))
	if slices.Contains([]error{github.ErrInvalidHTTPMethod, github.ErrEventNotFound}, err) {
		return http.StatusBadRequest, err
	} else if slices.Contains([]error{github.ErrMissingGithubEventHeader, github.ErrMissingHubSignatureHeader, github.ErrHMACVerificationFailed}, err) {
		return http.StatusUnauthorized, err
	} else if err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}
