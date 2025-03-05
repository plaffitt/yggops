# 🌳 YggOps

Inspired by [Yggdrasil](https://en.wikipedia.org/wiki/Yggdrasil), the [world tree](https://en.wikipedia.org/wiki/World_tree) in Norse mythology, YggOps is a tool for managing application deployments in a [GitOps](https://opengitops.dev) fashion. Just as the tree connects the nine worlds, YggOps connects the various technologies in a unified way of deploying them.

The idea of YggOps came to me as I was realizing that while tools such as ArgoCD and FluxCD deploy applications on Kubernetes the GitOps way, no such tool exists when you solely want to deploy something as simple as a Docker Compose stack on a single bare-metal server. I then realized that I could easily write some code to regularly pull a git repository and apply its contents in a generic manner, bringing the GitOps philosophy to any tool you can ever imagine.

## Configuration

| entry | default | description |
|-|-|-|
| `updateFrequency` | | Frequency of repository updates |
| `privateKeyPath` | | Path to private key to use to pull (optional) |
| `projects` | | List of projects to handle |
| `projects.type` | | Name of the plugin to use |
| `projects.name` | [repository name] | Name of the project |
| `projects.repository` | | Url of the repository to clone |
| `projects.branch` | main | Name of the branch to be synchronized with |
| `projects.options` | | Options to pass to the plugin as command line flags |

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
      script: ls && cat README.md | head -n 1
```
