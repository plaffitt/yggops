# 🌳 YggOps

Inspired by [Yggdrasil](https://en.wikipedia.org/wiki/Yggdrasil), the [world tree](https://en.wikipedia.org/wiki/World_tree) in Norse mythology, YggOps is a tool for managing application deployments in a [GitOps](https://opengitops.dev) fashion. Just as the tree connects the nine worlds, YggOps connects the various technologies in a unified way of deploying them.

The idea of YggOps came to me as I was realizing that while tools such as ArgoCD and FluxCD deploy applications on Kubernetes the GitOps way, no such tool exists when you solely want to deploy something as simple as a Docker Compose stack on a single bare-metal server. I then realized that I could easily write some code to regularly pull a git repository and apply its contents in a generic manner, bringing the GitOps philosophy to any tool you can ever imagine.

## How it works

YggOps can be run as a systemd service or as a standalone binary. It will regularly pull git repositories and apply their contents using plugins. It can also be triggered by a webhook to reconcile a repository immediately.

Plugins are what allow you to deploy your code in various ways. They are idempotent and are responsible for deploying specific things such as docker compose stacks, shell scripts, etc. You can easily imagine writing a plugin to deploy a Terraform configuration, a systemd service or anything else you can think of. Actually you could even deploy a Kubernetes manifest with a plugin if you wanted to, but the main interest is to deploy code that should run on the host where YggOps is running, otherwise tools like GitLab CI/CD are largely enough.

## Installation

For now, YggOps hasn't been released yet, so you will have to build it yourself. You will need to have [Go](https://go.dev/doc/install) installed on your machine as well as `make` (`sudo apt install make` should do it). To install YggOps on a remote host, clone this repository and run the following commands within it:

```sh
make build # or make docker-build if you don't have Go installed
sudo make install
```

It will install YggOps systemd service, copy its default configuration to `/etc/yggops/config.yaml`, and start it.

### SSH configuration


In order to let YggOps clone git repositories over ssh, you will need to provide a private key using the `privateKeyPath` configuration option. Don't forget to set the correct permissions on the private key file and add the corresponding public key to your git repositories. You may want to use a dedicated key instead of your personnal one.

Read the following documentation to know how to add an ssh key to your git server:

- [GitHub](https://docs.github.com/en/authentication/connecting-to-github-with-ssh/adding-a-new-ssh-key-to-your-github-account)
- [GitLab](https://docs.gitlab.com/user/ssh/#add-an-ssh-key-to-your-gitlab-account)

After adding the public key to your git server, you can test the connection with the following command:

```sh
ssh -T git@your-git-server.example.com -i /path/to/private-key
```

Remember that you have to trust every single git server you want to clone from, so you may have to run this command for each of them.

### Configuration

| entry | default | description |
|-|-|-|
| `updateInterval` | | Interval of repository updates |
| `privateKeyPath` | | Path to private key to use to pull (optional) |
| `listen` | :3000 | Address to listen to |
| `projects` | | List of projects to handle |
| `projects.type` | | Name of the plugin to use |
| `projects.name` | [repository name] | Name of the project |
| `projects.repository` | | Url of the repository to clone |
| `projects.branch` | main | Name of the branch to be synchronized with |
| `projects.updateInterval` | | Overrides the global value |
| `projects.options` | | Options to pass to the plugin |
| `projects.webhook` | | Configuration of a webhook to trigger reconciliation |
| `projects.webhook.provider` | | Can be `github`, `gitlab` or `generic`|
| `projects.webhook.secret` | | The webhook secret |
| `projects.webhook.getSecretCommand` | | A command that outputs the webhook secret on stdout |
| `projects.webhook.events` | `push` | A list of events to react to |

Example:

```yaml
updateInterval: 5m
privateKeyPath: /home/user/.ssh/id_ed25519
projects:
  - type: docker-compose
    repository: git@github.com:username/docker-compose-project.git
    updateInterval: 1h
    webhook:
      provider: github
      secret: my_secret
      events: [push]
  - type: shell
    name: custom_name
    repository: git@github.com:username/project.git
    branch: deploy
    options:
      script: install awesome_script.sh /usr/local/bin
```

`projects.webhook.secret` and `projects.webhook.getSecretCommand` cannot be set at the same time, but if both are empty, YggOps will read the secret from `/etc/yggops/webhook-secrets/<project_name>`.

## Plugins

YggOps comes with a few plugins by default. For now it includes `shell` and `docker-compose` plugins, but it may includes additional plugins in the future. Plugins are located in `/var/lib/yggops/plugins/`.

### Shell

The shell plugin is a bit special because it is almost not a plugin since it does nothing but running a command that you give to it. So providing the source code of another plugin written in bash as the `script` option of this plugin would be the same as using the given plugin directly.

| option | required | default | description |
|-|-|-|-|
| `script` | yes | | Script to run |

### Docker Compose

| option | required | default | description |
|-|-|-|-|
| `env-file` | no | | Flag `--env-file` of the docker compose CLI |
| `build` | no | `true` | Flag `--build` of the docker compose CLI |
| `remove-orphans` | no | `true` | Flag `--remove-orphans` of the docker compose CLI |

### Write your own

If default plugins don't suit your needs, you can easily write your own plugin in any language you want. There are only a few rules:

- They have to be idempotent
- They have to assume that the working directory will be reset before reconciliation (there is no persistence, so state should be kept somewhere else)

Options will be passed to plugins as flags. The filename of the plugin will be used to reference it in the `type` entry of project definition.
