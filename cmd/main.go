package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"gopkg.in/yaml.v2"
)

type Project struct {
	Name             string            `yaml:"name"`
	Type             string            `yaml:"type"`
	Repository       string            `yaml:"repository"`
	Branch           string            `yaml:"branch"`
	Options          map[string]string `yaml:"options"`
	RepositoriesPath *string

	repository *git.Repository
	worktree   *git.Worktree
	headHash   plumbing.Hash
}

type Config struct {
	UpdateFrequency time.Duration `yaml:"updateFrequency"`
	Projects        []*Project    `yaml:"projects"`
}

func main() {
	configPath := flag.String("config", "/etc/generic-gitops/config.yaml", "Configuration file path")
	pluginsPath := flag.String("plugins", "/var/lib/generic-gitops/plugins", "Plugins directory path")
	repositoriesPath := flag.String("repositories", "/var/lib/generic-gitops/repositories", "Repositories directory path")
	flag.Parse()

	yamlFile, err := os.ReadFile(*configPath)
	if err != nil {
		log.Fatalf("error reading YAML file: %v", err)
	}

	var config Config

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalf("error unmarshalling YAML: %v", err)
	}

	fmt.Println("Update frequency:", config.UpdateFrequency)

	fmt.Println("\nProjects:\n=========================")
	for _, project := range config.Projects {
		project.RepositoriesPath = repositoriesPath
		if project.Name == "" {
			repositorySlice := strings.Split(project.Repository, "/")
			project.Name = strings.Split(repositorySlice[len(repositorySlice)-1], ".")[0]
		}
		if project.Branch == "" {
			project.Branch = "main"
		}

		// TODO check that plugin project.Type exists

		fmt.Println("Name:", project.Name)
		fmt.Println("Type:", project.Type)
		fmt.Println("Repository:", project.Repository)
		fmt.Println("Branch:", project.Branch)
		fmt.Println("Options:", project.Options)
		fmt.Println("=========================")
	}

	fmt.Println("")

	for true {
		for _, project := range config.Projects {
			updated, err := project.Update()
			if err != nil {
				fmt.Println(err)
			}

			// TODO check that plugin project.Type exists

			if updated {
				args := []string{}
				for name, option := range project.Options {
					args = append(args, "--"+name)
					args = append(args, option)
				}

				pluginPath, err := filepath.Abs(*pluginsPath + "/" + project.Type)
				if err != nil {
					log.Fatal(err)
				}

				cmd := exec.Command(pluginPath, args...)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				cmd.Dir = project.RepositoryPath()

				err = cmd.Run()
				if err != nil {
					log.Fatal(err)
				}
			}
			fmt.Println("=========================")
		}
		time.Sleep(config.UpdateFrequency)
	}
}

func (p *Project) Update() (bool, error) {
	err := p.openWorktree()
	if err == git.ErrRepositoryNotExists {
		if err = p.clone(); err != nil {
			return false, err
		}
		if err = p.openWorktree(); err != nil {
			return false, err
		}
		return p.updateHeadHash()
	}

	if err != nil {
		return false, err
	}

	fmt.Printf("Updating %s (%s)...\n", p.Name, p.Repository)
	err = p.worktree.Pull(&git.PullOptions{RemoteName: "origin"})
	if err != nil {
		return false, err
	}

	fmt.Printf("%s updated successfully!\n", p.Name)

	return p.updateHeadHash()
}

func (p *Project) updateHeadHash() (bool, error) {
	var err error
	previousHash := p.headHash

	p.headHash, err = p.getHeadHash()
	if err != nil {
		return false, err
	}

	return previousHash != p.headHash, nil
}

func (p *Project) clone() error {
	fmt.Printf("Cloning %s into %s...\n", p.Repository, p.RepositoryPath())
	_, err := git.PlainClone(p.RepositoryPath(), false, &git.CloneOptions{
		URL:               p.Repository,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		SingleBranch:      true,
		ReferenceName:     plumbing.ReferenceName(p.Branch),
	})

	if err != nil {
		return fmt.Errorf("could not clone %s: %s", p.Repository, err)
	} else {
		fmt.Printf("%s cloned successfully!\n", p.Repository)
	}

	return nil
}

func (p *Project) openWorktree() (err error) {
	p.repository, err = git.PlainOpen(p.RepositoryPath())
	if err != nil {
		return
	}

	p.worktree, err = p.repository.Worktree()
	if err != nil {
		return
	}

	return
}

func (p *Project) getHeadHash() (plumbing.Hash, error) {
	ref, err := p.repository.Head()
	if err != nil {
		return plumbing.Hash{}, err
	}

	return ref.Hash(), nil
}

func (p *Project) RepositoryPath() string {
	return *p.RepositoriesPath + "/" + p.Name
}
