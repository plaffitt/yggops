package internal

import (
	"context"
	"fmt"
	"log"
	"net/http"
)

type Controller struct {
	config *Config
	ctx    context.Context
	cancel context.CancelFunc
	server *http.Server
}

func NewController(config *Config) *Controller {
	return &Controller{config: config}
}

func (c *Controller) Start() error {
	c.config.Watch(func() {
		fmt.Println("Configuration file changed, restarting the controller...")
		if err := c.Stop(); err != nil {
			log.Fatalf("could not restart the controller: %s", err.Error())
		}
	})

	var err error

	for err == nil || err == http.ErrServerClosed {
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

		handler := http.NewServeMux()
		c.server = &http.Server{Addr: c.config.Listen, Handler: handler}
		c.ctx, c.cancel = context.WithCancel(context.Background())
		for _, project := range c.config.Projects {
			project.RegisterWebhook(handler)
			go project.KeepUpdated(c.ctx)
		}

		fmt.Println("Starting webhook server, listening on", c.config.Listen)

		err = c.server.ListenAndServe()
	}

	return err
}

func (c *Controller) Stop() error {
	c.cancel()
	if err := c.server.Shutdown(context.Background()); err != nil {
		return fmt.Errorf("could not shut down the server: %w", err)
	}
	return nil
}

func (c *Controller) Restart() error {
	if err := c.Stop(); err != nil {
		return fmt.Errorf("could not restart the controller: %w", err)
	}

	return c.Start()
}
