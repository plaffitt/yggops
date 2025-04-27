package internal

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/iancoleman/strcase"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/pflag"
)

type Config struct {
	UpdateInterval     time.Duration `koanf:"updateInterval"`
	PrivateKeyPath     string        `koanf:"privateKeyPath"`
	Projects           []*Project    `koanf:"projects"`
	Listen             string        `koanf:"listen"`
	WebhookSecretsPath string        `koanf:"webhookSecrets"`
	RepositoriesPath   string        `koanf:"repositories"`
	PluginsPath        string        `koanf:"plugins"`

	fileProvider *file.File
	flags        *pflag.FlagSet
}

func NewConfig(configPath string, flags *pflag.FlagSet) *Config {
	return &Config{
		fileProvider: file.Provider(configPath),
		flags:        flags,
	}
}

func (c *Config) Load() error {
	k := koanf.New(".")

	if err := k.Load(confmap.Provider(map[string]any{
		"updateInterval": "5m",
	}, "."), nil); err != nil {
		return err
	}

	if err := k.Load(c.fileProvider, yaml.Parser()); err != nil {
		return err
	}

	if err := k.Load(posflag.ProviderWithValue(c.flags, ".", k, func(key, value string) (string, any) {
		return strcase.ToCamel(key), value
	}), nil); err != nil {
		return err
	}

	if err := k.Unmarshal("", c); err != nil {
		return err
	}

	if err := c.loadProjectsConfig(); err != nil {
		return err
	}

	return nil
}

func (c *Config) Watch(callback func()) {
	c.fileProvider.Watch(func(event any, err error) {
		if err != nil {
			log.Fatalf("error while watching configuration: %s", err.Error())
		}

		if err := c.Load(); err != nil {
			log.Fatalf("error while reloading configuration: %s", err.Error())
		}

		callback()
	})
}

func (c *Config) getAuthMethod() (auth transport.AuthMethod, err error) {
	if c.PrivateKeyPath != "" {
		auth, err = ssh.NewPublicKeysFromFile("git", c.PrivateKeyPath, "")
		if err != nil {
			err = fmt.Errorf("failed to create auth: %v", err)
		}
	}

	return
}

func (c *Config) loadProjectsConfig() error {
	auth, err := c.getAuthMethod()
	if err != nil {
		return fmt.Errorf("failed to load projects: %v", err)
	}

	for _, project := range c.Projects {
		project.RepositoriesPath = &c.RepositoriesPath
		project.PluginsPath = &c.PluginsPath
		project.Auth = auth
		if project.Name == "" {
			repositorySlice := strings.Split(project.Repository, "/")
			project.Name = strings.Split(repositorySlice[len(repositorySlice)-1], ".")[0]
		}
		if project.Branch == "" {
			project.Branch = "main"
		}
		if project.UpdateInterval == 0 {
			project.UpdateInterval = c.UpdateInterval
		}

		if project.Webhook != nil {
			webhook := project.Webhook

			if webhook.Secret != "" && webhook.GetSecretCommand != "" {
				return fmt.Errorf("both secret and getSecretCommand are set for %s webhook", project.Name)
			} else if webhook.Secret == "" && webhook.GetSecretCommand == "" {
				secret, err := os.ReadFile(c.WebhookSecretsPath + "/" + project.Name)
				if err != nil {
					return fmt.Errorf("no secret was configured for %s webhook", project.Name)
				}
				webhook.Secret = strings.TrimRight(string(secret), "\n")
			} else if webhook.GetSecretCommand != "" {
				output, err := exec.Command("sh", "-c", webhook.GetSecretCommand).CombinedOutput()
				if err != nil {
					return fmt.Errorf("could not get secret for %s: %w", project.Name, err)
				}
				webhook.Secret = strings.TrimRight(string(output), "\n")
			}

			if err := webhook.Init(project); err != nil {
				return fmt.Errorf("could not init webhook for %s: %w", project.Name, err)
			}
		}

		pluginPath, err := project.PluginPath()
		if err != nil {
			return err
		}
		if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
			return fmt.Errorf("plugin \"%s\" not found for project \"%s\"", project.Type, project.Name)
		}
		if err := os.Chmod(pluginPath, 0700); err != nil {
			return fmt.Errorf("plugin \"%s\" is not executable", pluginPath)
		}
	}

	return nil
}
