# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get

# Name of the binary
BINARY_NAME = generic-gitops

# Service configuration
USER_NAME = root
GROUP_NAME = root
SERVICE_NAME = $(BINARY_NAME).service
SERVICE_PATH = /etc/systemd/system/$(SERVICE_NAME)

all: build

build:
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/main.go

docker-build:
	docker build --build-arg BINARY_NAME=$(BINARY_NAME) . -t $(BINARY_NAME)
	docker run $(BINARY_NAME) cat $(BINARY_NAME) > $(BINARY_NAME)
	chmod +x $(BINARY_NAME)

test:
	$(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

install_plugins:
	mkdir -p /var/lib/$(BINARY_NAME)
	cp -r ./plugins /var/lib/$(BINARY_NAME)

uninstall_plugins:
	rm -rf /var/lib/$(SERVICE_PATH)/plugins

install: generic-gitops install_plugins
	mkdir -p /etc/$(BINARY_NAME)
	if [ ! -f /etc/$(BINARY_NAME)/config.yaml ]; then install -m 0644 ./systemd/config.yaml /etc/$(BINARY_NAME)/config.yaml; fi
	install -m 0755 $(BINARY_NAME) /usr/bin/$(BINARY_NAME)
	USER_NAME=$(USER_NAME) GROUP_NAME=$(GROUP_NAME) BINARY_NAME=$(BINARY_NAME) envsubst < ./systemd/$(SERVICE_NAME) | install -m 0644 /dev/stdin $(SERVICE_PATH)
	systemctl daemon-reload
	systemctl enable $(SERVICE_NAME)
	systemctl start $(SERVICE_NAME)

uninstall:
	systemctl stop $(SERVICE_NAME) || true
	systemctl disable $(SERVICE_NAME) || true
	rm -f /usr/bin/$(BINARY_NAME)
	rm -f $(SERVICE_PATH)
	systemctl daemon-reload

.PHONY: all build docker-build test clean run install_plugins uninstall_plugins install uninstall
