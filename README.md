# generic-gitops

A tool for integrating the gitops philosophy with any existing tools.

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
