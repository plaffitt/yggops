package internal

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"gopkg.in/yaml.v2"
)

type Config struct {
	UpdateFrequency    time.Duration `yaml:"updateFrequency"`
	PrivateKeyPath     string        `yaml:"privateKeyPath"`
	Projects           []*Project    `yaml:"projects"`
	Listen             string        `yaml:"listen"`
	WebhookSecretsPath string
	RepositoriesPath   string
	PluginsPath        string
}

func LoadConfig(configPath string) (*Config, error) {
	yamlFile, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading YAML file: %v", err)
	}

	var config Config

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling YAML: %v", err)
	}

	listenFlag := flag.Lookup("listen")
	if listenFlag.Value.String() != listenFlag.DefValue || config.Listen == "" {
		config.Listen = listenFlag.Value.String()
	}
	config.WebhookSecretsPath = flag.Lookup("webhook-secrets").Value.String()
	config.PluginsPath = flag.Lookup("plugins").Value.String()
	config.RepositoriesPath = flag.Lookup("repositories").Value.String()

	if err = config.loadProjectsConfig(); err != nil {
		return nil, err
	}

	return &config, err
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
		if project.UpdateFrequency == 0 {
			project.UpdateFrequency = c.UpdateFrequency
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

		// TODO check that plugin project.Type exists
	}

	return nil
}
