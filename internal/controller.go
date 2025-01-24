package internal

import (
	"context"
	"fmt"
	"net/http"
)

type Controller struct {
	config *Config
}

type Hook[T ~string] interface {
	Parse(r *http.Request, events ...T) (interface{}, error)
}

func NewController(config *Config) *Controller {
	return &Controller{config}
}

func (c *Controller) Start() error {
	fmt.Println("Update frequency:", c.config.UpdateFrequency)
	fmt.Println("Private key path:", c.config.PrivateKeyPath)

	fmt.Println("\nProjects:\n=========================")
	for _, project := range c.config.Projects {
		fmt.Println("Name:", project.Name)
		fmt.Println("Type:", project.Type)
		fmt.Println("Repository:", project.Repository)
		fmt.Println("Branch:", project.Branch)
		fmt.Println("UpdateFrequency:", project.UpdateFrequency)
		fmt.Println("Webhook:", project.WebhookPath())
		fmt.Println("Options:", project.Options)
		fmt.Println("=========================")
	}

	fmt.Println()

	ctx := context.Background()
	for _, project := range c.config.Projects {
		if err := project.RegisterWebhook(); err != nil {
			return err
		}
		go project.KeepUpdated(ctx)
	}

	return http.ListenAndServe(":3000", nil)
}
