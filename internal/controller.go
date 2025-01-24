package internal

import (
	"fmt"
	"time"
)

type Controller struct {
	config *Config
}

func NewController(config *Config) *Controller {
	return &Controller{config}
}

func (c *Controller) Start(pluginsPath string, repositoriesPath string) {
	fmt.Println("Update frequency:", c.config.UpdateFrequency)
	fmt.Println("Private key path:", c.config.PrivateKeyPath)

	fmt.Println("\nProjects:\n=========================")
	for _, project := range c.config.Projects {
		fmt.Println("Name:", project.Name)
		fmt.Println("Type:", project.Type)
		fmt.Println("Repository:", project.Repository)
		fmt.Println("Branch:", project.Branch)
		fmt.Println("Options:", project.Options)
		fmt.Println("=========================")
	}

	fmt.Println()

	for {
		for _, project := range c.config.Projects {
			err := project.Load()
			if err != nil {
				fmt.Printf("could not load %s: %s\n", project.Name, err)
				continue
			}

			err = project.UpdateSources()
			if err != nil {
				fmt.Printf("could not update %s sources: %s\n", project.Name, err)
				continue
			}

			err = project.ApplyPatch(pluginsPath)
			if err != nil {
				fmt.Printf("could not apply patch to %s: %s\n", project.Name, err)
			}

			fmt.Println("=========================")
		}
		time.Sleep(c.config.UpdateFrequency)
	}
}
