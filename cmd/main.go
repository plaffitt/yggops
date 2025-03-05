package main

import (
	"flag"
	"log"

	"github.com/plaffitt/yggops/internal"
)

func main() {
	configPath := flag.String("config", "/etc/yggops/config.yaml", "Configuration file path")
	flag.String("listen", ":3000", "Webhook listen address (<ip:port>)")
	flag.String("webhook-secrets", "/etc/yggops/webhook-secrets", "Webhook secrets directory path")
	flag.String("plugins", "/var/lib/yggops/plugins", "Plugins directory path")
	flag.String("repositories", "/var/lib/yggops/repositories", "Repositories directory path")
	flag.Parse()

	config, err := internal.LoadConfig(*configPath)
	if err != nil {
		log.Fatal(err.Error())
	}

	controller := internal.NewController(config)
	if err := controller.Start(); err != nil {
		log.Fatal(err.Error())
	}
}
