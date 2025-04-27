package main

import (
	"log"
	"os"

	"github.com/plaffitt/yggops/internal"
	"github.com/plaffitt/yggops/internal/version"
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
		Use:     "yggops",
		Short:   "YggOps controller",
		Version: version.BuildVersion(),
		Run: func(cmd *cobra.Command, args []string) {
			config := internal.NewConfig(configPath, cmd.LocalFlags())
			if err := config.Load(); err != nil {
				log.Fatal("error while loading configuration: " + err.Error())
			}

			controller := internal.NewController(config)

			if err := controller.Start(); err != nil {
				log.Fatal(err.Error())
			}
		},
	}

	rootCmd.PersistentFlags().StringVar(&configPath, "config", "/etc/yggops/config.yaml", "configuration file path")
	rootCmd.PersistentFlags().StringVar(&pluginsDir, "plugins", "/var/lib/yggops/plugins", "plugins directory path")
	rootCmd.PersistentFlags().StringVar(&repositoriesDir, "repositories", "/var/lib/yggops/repositories", "repositories directory path")
	rootCmd.Flags().StringVar(&listenAddr, "listen", ":3000", "webhook listen address (<ip:port>)")
	rootCmd.Flags().StringVar(&webhookSecrets, "webhook-secrets", "/etc/yggops/webhook-secrets", "webhook secrets directory path")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
