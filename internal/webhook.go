package internal

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"slices"

	"github.com/go-playground/webhooks/v6/github"
	"github.com/go-playground/webhooks/v6/gitlab"
)

type WebhookProvider string

const (
	GenericProvider WebhookProvider = "generic"
	GithubProvider  WebhookProvider = "github"
	GitlabProvider  WebhookProvider = "gitlab"
)

var bearerRegexp *regexp.Regexp = regexp.MustCompile("^Bearer .+")

type Webhook struct {
	Provider WebhookProvider `yaml:"provider"`
	Secret   string          `yaml:"secret"`
	Event    string          `yaml:"event"`

	github *github.Webhook
	gitlab *gitlab.Webhook
}

func (w *Webhook) Init() error {
	switch w.Provider {
	case GithubProvider:
		hook, err := github.New(github.Options.Secret(w.Secret))
		if err != nil {
			return err
		}
		w.github = hook
		return nil
	case GitlabProvider:
		hook, err := gitlab.New(gitlab.Options.Secret(w.Secret))
		if err != nil {
			return err
		}
		w.gitlab = hook
		return nil
	case GenericProvider:
		return nil
	}

	return fmt.Errorf("invalid webhook provider: %s", w.Provider)
}

func (w *Webhook) Validate(writer http.ResponseWriter, request *http.Request) error {
	var err error

	switch w.Provider {
	case GithubProvider:
		event := request.Header.Get("X-GitHub-Event")
		fmt.Printf("Received Github \"%s\" event\n", event)
		_, err = w.github.Parse(request, github.Event(w.Event))
		if slices.Contains([]error{github.ErrInvalidHTTPMethod, github.ErrEventNotFound}, err) {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return err
		} else if slices.Contains([]error{github.ErrMissingGithubEventHeader, github.ErrMissingHubSignatureHeader, github.ErrHMACVerificationFailed}, err) {
			http.Error(writer, err.Error(), http.StatusUnauthorized)
			return err
		}
	case GitlabProvider:
		event := request.Header.Get("X-Gitlab-Event")
		fmt.Printf("Received Gitlab \"%s\" event\n", event)
		_, err = w.gitlab.Parse(request, gitlab.Event(w.Event))
		if slices.Contains([]error{gitlab.ErrInvalidHTTPMethod, gitlab.ErrEventNotFound}, err) {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return err
		} else if slices.Contains([]error{gitlab.ErrMissingGitLabEventHeader, gitlab.ErrGitLabTokenVerificationFailed}, err) {
			http.Error(writer, err.Error(), http.StatusUnauthorized)
			return err
		}
	case GenericProvider:
		fmt.Printf("Received generic webhook event\n")
		authorization := request.Header.Get("Authorization")
		if bearerRegexp.Match([]byte(authorization)) {
			token := authorization[len("Bearer "):]
			if token != w.Secret {
				err = errors.New("invalid token")
				http.Error(writer, err.Error(), http.StatusUnauthorized)
				return err
			}
		} else {
			err = errors.New("missing token")
			http.Error(writer, err.Error(), http.StatusUnauthorized)
			return err
		}
	}

	if err != nil {
		err = fmt.Errorf("unexpected error %w", err)
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return err
	}

	fmt.Fprint(writer, "OK")
	return nil
}
