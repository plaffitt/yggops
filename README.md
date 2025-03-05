# 🌳 YggOps

Inspired by [Yggdrasil](https://en.wikipedia.org/wiki/Yggdrasil), the [world tree](https://en.wikipedia.org/wiki/World_tree) in Norse mythology, YggOps is a tool for managing application deployments in a [GitOps](https://opengitops.dev) fashion. Just as the tree connects the nine worlds, YggOps connects the various technologies in a unified way of deploying them.

The idea of YggOps came to me as I was realizing that while tools such as ArgoCD and FluxCD deploy applications on Kubernetes the GitOps way, no such tool exists when you solely want to deploy something as simple as a Docker Compose stack on a single bare-metal server. I then realized that I could easily write some code to regularly pull a git repository and apply its contents in a generic manner, bringing the GitOps philosophy to any tool you can ever imagine.

## How it works

YggOps can be run as a systemd service or as a standalone binary. It will regularly pull git repositories and apply their contents using plugins. It can also be triggered by a webhook to reconcile a repository immediately.

Plugins are what allow you to deploy your code in various ways. They are idempotent and are responsible for deploying specific things such as docker compose stacks, shell scripts, etc. You can easily imagine writing a plugin to deploy a Terraform configuration, a systemd service or anything else you can think of. Actually you could even deploy a Kubernetes manifest with a plugin if you wanted to, but the main interest is to deploy code that should run on the host where YggOps is running, otherwise tools like GitLab CI/CD are largely enough.

## Configuration

| entry | default | description |
|-|-|-|
| `updateFrequency` | | Frequency of repository updates |
| `privateKeyPath` | | Path to private key to use to pull (optional) |
| `listen` | :3000 | Address to listen to |
| `projects` | | List of projects to handle |
| `projects.type` | | Name of the plugin to use |
| `projects.name` | [repository name] | Name of the project |
| `projects.repository` | | Url of the repository to clone |
| `projects.branch` | main | Name of the branch to be synchronized with |
| `projects.updateFrequency` | | Overrides the global value |
| `projects.options` | | Options to pass to the plugin as command line flags |
| `projects.webhook` | | Configuration of a webhook to trigger reconciliation |

Example:

```yaml
updateFrequency: 5m
privateKeyPath: /home/user/.ssh/id_ed25519
projects:
  - type: docker_compose
    repository: git@github.com:username/docker-compose-project.git
  - type: shell
    name: custom_name
    repository: git@github.com:username/project.git
    branch: deploy
    options:
      script: install awesome_script.sh /usr/local/bin
```
