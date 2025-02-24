package webhooks

import (
	"fmt"
	"net/http"
	"slices"

	"github.com/go-playground/webhooks/v6/github"
)

type Github struct {
	git

	handler *github.Webhook
	events  []github.Event
}

func NewGithub(secret string, events []string, branch string) (*Github, error) {
	handler, err := github.New(github.Options.Secret(secret))
	if err != nil {
		return nil, err
	}

	githubEvents := []github.Event{}
	for _, event := range events {
		githubEvents = append(githubEvents, github.Event(event))
	}

	if len(githubEvents) == 0 {
		githubEvents = []github.Event{github.PushEvent}
	}

	return &Github{
		git:     git{branch: branch},
		handler: handler,
		events:  githubEvents,
	}, nil
}

func (g *Github) Validate(request *http.Request) (int, error) {
	event := request.Header.Get("X-GitHub-Event")
	fmt.Printf("Received Github \"%s\" event\n", event)

	payload, err := g.handler.Parse(request, g.events...)
	if slices.Contains([]error{github.ErrInvalidHTTPMethod, github.ErrEventNotFound}, err) {
		return http.StatusBadRequest, err
	} else if slices.Contains([]error{github.ErrMissingGithubEventHeader, github.ErrMissingHubSignatureHeader, github.ErrHMACVerificationFailed}, err) {
		return http.StatusUnauthorized, err
	} else if err != nil {
		return http.StatusInternalServerError, err
	}

	if pushPayload, ok := payload.(github.PushPayload); ok {
		if err := g.validateBranch(pushPayload.Ref); err != nil {
			return http.StatusBadRequest, err
		}
	}

	return http.StatusOK, nil
}
