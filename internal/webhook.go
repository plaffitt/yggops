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
	ManualProvider WebhookProvider = "manual"
	GitHubProvider WebhookProvider = "github"
	GitLabProvider WebhookProvider = "gitlab"
)

var bearerRegexp *regexp.Regexp = regexp.MustCompile("^Bearer .+")

type Webhook struct {
	Provider         WebhookProvider `yaml:"provider"`
	Secret           string          `yaml:"secret"`
	GetSecretCommand string          `yaml:"getSecretCommand"`
	Event            string          `yaml:"event"`

	github *github.Webhook
	gitlab *gitlab.Webhook
}

func (w *Webhook) Init() error {
	switch w.Provider {
	case GitHubProvider:
		hook, err := github.New(github.Options.Secret(w.Secret))
		if err != nil {
			return err
		}
		w.github = hook
		return nil
	case GitLabProvider:
		hook, err := gitlab.New(gitlab.Options.Secret(w.Secret))
		if err != nil {
			return err
		}
		w.gitlab = hook
		w.Event = getGitlabEvent(w.Event)
		return nil
	case ManualProvider:
		return nil
	}

	return fmt.Errorf("invalid webhook provider: %s", w.Provider)
}

func (w *Webhook) Validate(writer http.ResponseWriter, request *http.Request) error {
	var err error

	switch w.Provider {
	case GitHubProvider:
		event := request.Header.Get("X-GitHub-Event")
		fmt.Printf("Received GitHub \"%s\" event\n", event)
		_, err = w.github.Parse(request, github.Event(w.Event))
		if slices.Contains([]error{github.ErrInvalidHTTPMethod, github.ErrEventNotFound}, err) {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return err
		} else if slices.Contains([]error{github.ErrMissingGithubEventHeader, github.ErrMissingHubSignatureHeader, github.ErrHMACVerificationFailed}, err) {
			http.Error(writer, err.Error(), http.StatusUnauthorized)
			return err
		}
	case GitLabProvider:
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
	case ManualProvider:
		fmt.Printf("Received manual webhook event\n")
		authorization := request.Header.Get("Authorization")
		if bearerRegexp.Match([]byte(authorization)) {
			token := authorization[len("Bearer "):]
			if token != w.Secret {
				err = errors.New("invalid token")
				http.Error(writer, err.Error(), http.StatusUnauthorized)
				return err
			}
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

func getGitlabEvent(name string) string {
	switch name {
	case "push":
		return string(gitlab.PushEvents)
	case "tag":
		return string(gitlab.TagEvents)
	case "merge_request":
		return string(gitlab.MergeRequestEvents)
	case "pipeline":
		return string(gitlab.PipelineEvents)
	case "build":
		return string(gitlab.BuildEvents)
	case "job":
		return string(gitlab.JobEvents)
	case "deployment":
		return string(gitlab.DeploymentEvents)
	case "release":
		return string(gitlab.ReleaseEvents)
	case "system":
		return string(gitlab.SystemHookEvents)
	default:
		return ""
	}
}
