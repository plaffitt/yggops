package main

import (
	"log"
	"os"

	"github.com/plaffitt/yggops/internal"
	"github.com/spf13/cobra"
)

var (
	configPath      string
	pluginsDir      string
	repositoriesDir string
	listenAddr      string
	webhookSecrets  string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "yggops",
		Short: "YggOps controller",
		Run: func(cmd *cobra.Command, args []string) {
			config, err := internal.LoadConfig(configPath, cmd.LocalFlags())
			if err != nil {
				log.Fatal(err.Error())
			}

			controller := internal.NewController(config)
			if err := controller.Start(); err != nil {
				log.Fatal(err.Error())
			}
		},
	}

	rootCmd.PersistentFlags().StringVar(&configPath, "config", "/etc/yggops/config.yaml", "Configuration file path")
	rootCmd.PersistentFlags().StringVar(&pluginsDir, "plugins", "/var/lib/yggops/plugins", "Plugins directory path")
	rootCmd.PersistentFlags().StringVar(&repositoriesDir, "repositories", "/var/lib/yggops/repositories", "Repositories directory path")
	rootCmd.Flags().StringVar(&listenAddr, "listen", ":3000", "Webhook listen address (<ip:port>)")
	rootCmd.Flags().StringVar(&webhookSecrets, "webhook-secrets", "/etc/yggops/webhook-secrets", "Webhook secrets directory path")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
