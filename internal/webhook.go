package internal

import (
	"fmt"
	"net/http"

	"github.com/plaffitt/generic-gitops/internal/webhooks"
)

type WebhookProvider string

const (
	GenericProvider WebhookProvider = "generic"
	GithubProvider  WebhookProvider = "github"
	GitlabProvider  WebhookProvider = "gitlab"
)

type Webhook struct {
	Provider         WebhookProvider `yaml:"provider"`
	Secret           string          `yaml:"secret"`
	GetSecretCommand string          `yaml:"getSecretCommand"`
	Events           []string        `yaml:"events"`

	project *Project
	handler webhooks.Handler
}

func (w *Webhook) Init(project *Project) error {
	w.project = project

	var err error

	switch w.Provider {
	case GenericProvider:
		w.handler, err = webhooks.NewGeneric(w.Secret)
	case GithubProvider:
		w.handler, err = webhooks.NewGithub(w.Secret, w.Events, w.project.Branch)
	case GitlabProvider:
		w.handler, err = webhooks.NewGitlab(w.Secret, w.Events, w.project.Branch)
	default:
		return fmt.Errorf("invalid webhook provider: %s", w.Provider)
	}

	return err
}

func (w *Webhook) Register() {
	http.HandleFunc(w.Path(), func(writer http.ResponseWriter, request *http.Request) {
		if status, err := w.handler.Validate(request); err != nil {
			if status == http.StatusInternalServerError {
				err = fmt.Errorf("unexpected error %w", err)
			}

			http.Error(writer, err.Error(), status)
			fmt.Printf("Webhook discarded for %s: %s\n", w.project.Name, err.Error())
			return
		}

		fmt.Fprint(writer, "OK")
		fmt.Printf("Webhook triggered for %s\n", w.project.Name)
		w.project.TriggerUpdate()
	})
}

func (w *Webhook) Path() string {
	return "/webhooks/" + string(w.Provider) + "/" + w.project.Name
}
