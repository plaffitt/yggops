package main

import (
	"flag"
	"log"

	"github.com/plaffitt/generic-gitops/internal"
)

func main() {
	configPath := flag.String("config", "/etc/generic-gitops/config.yaml", "Configuration file path")
	pluginsPath := flag.String("plugins", "/var/lib/generic-gitops/plugins", "Plugins directory path")
	repositoriesPath := flag.String("repositories", "/var/lib/generic-gitops/repositories", "Repositories directory path")
	flag.Parse()

	config, err := internal.LoadConfig(*configPath)
	if err != nil {
		log.Fatal(err.Error())
	}

	controller := internal.NewController(config)
	controller.Start(*pluginsPath, *repositoriesPath)
}
