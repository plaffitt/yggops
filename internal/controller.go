package internal

import (
	"context"
	"fmt"
	"net/http"
)

type Controller struct {
	config *Config
}

func NewController(config *Config) *Controller {
	return &Controller{config}
}

func (c *Controller) Start() error {
	fmt.Println("Update interval:", c.config.UpdateInterval)
	fmt.Println("Private key path:", c.config.PrivateKeyPath)

	fmt.Println("\nProjects:\n=========================")
	for _, project := range c.config.Projects {
		fmt.Println("Name:", project.Name)
		fmt.Println("Type:", project.Type)
		fmt.Println("Repository:", project.Repository)
		fmt.Println("Branch:", project.Branch)
		fmt.Println("UpdateInterval:", project.UpdateInterval)
		fmt.Println("Webhook:", project.WebhookPath())
		fmt.Println("Options:", project.Options)
		fmt.Println("=========================")
	}

	fmt.Println()

	ctx := context.Background()
	for _, project := range c.config.Projects {
		project.RegisterWebhook()
		go project.KeepUpdated(ctx)
	}

	fmt.Println("Starting webhook server, listening on", c.config.Listen)

	return http.ListenAndServe(c.config.Listen, nil)
}
