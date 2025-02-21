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
		project.RegisterWebhook()
		go project.KeepUpdated(ctx)
	}

	return http.ListenAndServe(":3000", nil)
}
