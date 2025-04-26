package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	gtime "github.com/plaffitt/yggops/internal/time"
)

type Project struct {
	Name             string         `yaml:"name"`
	Type             string         `yaml:"type"`
	Repository       string         `yaml:"repository"`
	Branch           string         `yaml:"branch"`
	UpdateInterval   time.Duration  `yaml:"updateInterval"`
	Webhook          *Webhook       `yaml:"webhook,omitempty"`
	Options          map[string]any `yaml:"options"`
	RepositoriesPath *string
	PluginsPath      *string
	Auth             transport.AuthMethod

	repository       *git.Repository
	worktree         *git.Worktree
	lastAppliedPatch plumbing.Hash
	updateMutex      sync.Mutex
	ticker           *gtime.TriggerableTicker
}

func (p *Project) Load() error {
	if err := p.openWorktree(); err == git.ErrRepositoryNotExists {
		if err = p.clone(); err != nil {
			return err
		}
		if err = p.openWorktree(); err != nil {
			return err
		}
	}

	if err := p.loadLastAppliedPatch(); err != nil {
		return err
	}

	return nil
}

func (p *Project) UpdateSources() error {
	fmt.Printf("Updating %s (%s)...\n", p.Name, p.Repository)

	currentHead, err := p.repository.Head()
	if err != nil {
		return err
	}

	// Fetch updates from remote
	err = p.repository.Fetch(&git.FetchOptions{
		RemoteName: "origin",
		Force:      true,
		RefSpecs: []config.RefSpec{
			config.RefSpec(fmt.Sprintf("+refs/heads/%s:refs/remotes/origin/%s", p.Branch, p.Branch)),
		},
		Auth: p.Auth,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}

	remoteRef, err := p.repository.Reference(plumbing.NewRemoteReferenceName("origin", p.Branch), true)
	if err != nil {
		return err
	}

	branchRefName := plumbing.NewBranchReferenceName(p.Branch)
	branchRef := plumbing.NewHashReference(branchRefName, remoteRef.Hash())

	err = p.repository.Storer.SetReference(branchRef)
	if err != nil {
		return err
	}

	// Checkout branch
	err = p.worktree.Checkout(&git.CheckoutOptions{
		Branch: branchRefName,
		Force:  true,
	})
	if err != nil {
		return err
	}

	// Update submodules
	sbs, err := p.worktree.Submodules()
	if err != nil {
		return err
	}

	err = sbs.Update(&git.SubmoduleUpdateOptions{
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	})
	if err != nil {
		return err
	}

	if currentHead.Hash() != remoteRef.Hash() {
		fmt.Printf("Successfully updated %s sources!\n", p.Name)
	}

	return nil
}

func (p *Project) ApplyPatch() error {
	headHash, err := p.getHeadHash()
	if err != nil {
		fmt.Println(err)
	}

	if headHash == p.lastAppliedPatch {
		fmt.Printf("Project %s is up to date (%s)\n", p.Name, headHash)
		return nil
	}

	fmt.Printf("Applying %s patch %s...\n", p.Name, headHash)
	jsonOptions, err := json.Marshal(p.Options)
	if err != nil {
		return err
	}

	pluginPath, err := p.PluginPath()
	if err != nil {
		return err
	}

	cmd := exec.Command(pluginPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = bytes.NewReader(jsonOptions)
	cmd.Dir = p.RepositoryPath()

	if err = cmd.Run(); err != nil {
		return err
	}

	if err = p.updateLastAppliedPatch(); err != nil {
		return err
	}

	fmt.Printf("Successfully applied %s patch %s!\n", p.Name, headHash)

	return nil
}

func (p *Project) Update() {
	p.updateMutex.Lock()
	defer p.updateMutex.Unlock()

	if err := p.Load(); err != nil {
		fmt.Printf("Could not load %s: %s\n", p.Name, err)
		return
	}

	if err := p.UpdateSources(); err != nil {
		fmt.Printf("Could not update %s sources: %s\n", p.Name, err)
		return
	}

	if err := p.ApplyPatch(); err != nil {
		fmt.Printf("Could not apply patch to %s: %s\n", p.Name, err)
		return
	}
}

func (p *Project) KeepUpdated(ctx context.Context) {
	p.ticker = gtime.NewTriggerableTicker(p.UpdateInterval, ctx)
	p.TriggerUpdate()

	for {
		select {
		case <-p.ticker.C:
			p.Update()
		case <-ctx.Done():
			return
		}
	}
}

func (p *Project) TriggerUpdate() {
	p.ticker.TriggerUpdate()
}

func (p *Project) updateLastAppliedPatch() (err error) {
	p.lastAppliedPatch, err = p.getHeadHash()
	if err != nil {
		return
	}

	err = os.WriteFile(p.RepositoryLastAppliedPatchPath(), []byte(p.lastAppliedPatch.String()), 0o644)
	if err != nil {
		return
	}

	return
}

func (p *Project) loadLastAppliedPatch() error {
	lastAppliedPatch, err := os.ReadFile(p.RepositoryLastAppliedPatchPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	p.lastAppliedPatch = plumbing.NewHash(string(lastAppliedPatch))

	return nil
}

func (p *Project) clone() error {
	fmt.Printf("Cloning %s into %s...\n", p.Repository, p.RepositoryPath())
	_, err := git.PlainClone(p.RepositoryPath(), false, &git.CloneOptions{
		URL:               p.Repository,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		SingleBranch:      true,
		ReferenceName:     plumbing.NewBranchReferenceName(p.Branch),
		Auth:              p.Auth,
	})

	if err != nil {
		return fmt.Errorf("could not clone %s: %s", p.Repository, err)
	} else {
		fmt.Printf("Successfully cloned %s!\n", p.Repository)
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

func (p *Project) RepositoryLastAppliedPatchPath() string {
	return *p.RepositoriesPath + "/" + p.Name + ".last_applied_patch"
}

func (p *Project) WebhookPath() string {
	if p.Webhook == nil {
		return ""
	}

	return p.Webhook.Path()
}

func (p *Project) RegisterWebhook() {
	if p.Webhook == nil {
		return
	}

	p.Webhook.Register()
}

func (p *Project) PluginPath() (string, error) {
	return filepath.Abs(*p.PluginsPath + "/" + p.Type)
}
