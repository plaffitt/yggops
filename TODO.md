# TODO

## roadmap

- [x] webhook tokens from env (GG_${NAME}_WEBHOOK_TOKEN) or from file (${name}.webhook.token) (config file may be commited to git)
- [x] webhook handle gitlab
- [x] webhook gitlab bi-directionnal rename events (push => Push Hook)
  - [x] test that it works (I didn't yet, only implemented it)
- [x] webhook handle multiple events
- [x] webhook events => nothing should default to push
- [x] webhook bind ip + port flags (or config ?)

---

- [ ] check required config are set + readme required config
- [ ] check that plugin project.Type exists in load config
- [ ] improve readme and documentation
- [ ] open source

---

- [ ] post on Enix' slack #veille-techno
- [ ] demo to Enix monkeys + ask for name ?
- [ ] version command
- [ ] semantic release (ZeroVer 0.x.y) + remove make docker-build and dockerfile as they aren't needed to build without golang installed anymore
- [ ] structured logging (log/slog)

## ideas (to discuss with Antoine and challenge against real world examples)

- [ ] guess provider: webhook.provider should be optionnal (handle self hosted gitlabs) but guessed from repository url
- [ ] support both GET and POST in generic provider
- [ ] genericHMAC provider
- [ ] list of plugins (for instance: sops + apt-install-deps + script)
  - [ ] implement it as a plugin (list-plugin) ?
- [ ] implement sops to support encrypted secrets <https://github.com/getsops/sops>
  - [ ] as a plugin first ?
  - [ ] then built-in when the api is stable
- [ ] github / gitlab deployment integrations
- [ ] deduplicate repository pull (separate sources and projects)
- [ ] use cobra and <https://github.com/knadh/koanf> to handle cli and config
- [ ] simple UI to see the status of each projects (deployed, deploying, error...) and informations (last trigger, last success, current revision)
- [ ] outgoing webhooks on status change
- [ ] repository cleanup (remove old branches, tags, ...)

### sources and projects

```yaml
sources:
  - name: foo
    repository: git@gitlab.com/plaffitt/foo.git
    branch: main
    updateFrequency: 1h
    webhook:
      provider: gitlab
      event: push
      secret: changeme


# projects are updated everytime its source is updated
projects:
  - name: foo-bar
    source: foo
    type: shell
    workdir: ./bar # workdir is new
    options:
      script: ./install.sh
  - name: foo-docker
    source: foo
    type: docker_compose
    workdir: ./docker
    options:
      env-file: /etc/generic-gitops/secrets/foo-docker
```
