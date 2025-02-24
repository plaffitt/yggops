package webhooks

import (
	"fmt"
	"net/http"
	"slices"

	"github.com/go-playground/webhooks/v6/gitlab"
)

type Gitlab struct {
	git

	handler *gitlab.Webhook
	events  []gitlab.Event
}

func NewGitlab(secret string, events []string, branch string) (*Gitlab, error) {
	handler, err := gitlab.New(gitlab.Options.Secret(secret))
	if err != nil {
		return nil, err
	}

	gitlabEvents := []gitlab.Event{}
	for _, event := range events {
		gitlabEvents = append(gitlabEvents, gitlab.Event(event))
	}

	if len(gitlabEvents) == 0 {
		gitlabEvents = []gitlab.Event{gitlab.PushEvents}
	}

	return &Gitlab{
		git:     git{branch: branch},
		handler: handler,
		events:  gitlabEvents,
	}, nil
}

func (g *Gitlab) Validate(request *http.Request) (int, error) {
	event := request.Header.Get("X-Gitlab-Event")
	fmt.Printf("Received Gitlab \"%s\" event\n", event)

	payload, err := g.handler.Parse(request, g.events...)
	if slices.Contains([]error{gitlab.ErrInvalidHTTPMethod, gitlab.ErrEventNotFound}, err) {
		return http.StatusBadRequest, err
	} else if slices.Contains([]error{gitlab.ErrMissingGitLabEventHeader, gitlab.ErrGitLabTokenVerificationFailed}, err) {
		return http.StatusUnauthorized, err
	} else if err != nil {
		return http.StatusInternalServerError, err
	}

	if pushPayload, ok := payload.(gitlab.PushEventPayload); ok {
		if err := g.validateBranch(pushPayload.Ref); err != nil {
			return http.StatusBadRequest, err
		}
	}

	return http.StatusOK, nil
}
