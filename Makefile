# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get

# Build parameters
BINARY_NAME = yggops
LDFLAGS = -X 'main.Version=$(VERSION)' \
          -X 'main.CommitHash=$(COMMIT_HASH)' \
          -X 'main.BuildTime=$(BUILD_TIMESTAMP)'
 
# Service configuration
USER_NAME = root
GROUP_NAME = root
SERVICE_NAME = $(BINARY_NAME).service
SERVICE_PATH = /etc/systemd/system/$(SERVICE_NAME)

all: build

build:
	$(GOBUILD) -o $(BINARY_NAME) -v -ldflags="$(LDFLAGS)" ./cmd

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

install: yggops install_plugins
	mkdir -p /etc/$(BINARY_NAME)
	if [ ! -f /etc/$(BINARY_NAME)/config.yaml ]; then install -m 0644 ./packaging/config/config.yaml /etc/$(BINARY_NAME)/config.yaml; fi
	install -m 0755 $(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	install -m 0644 ./packaging/systemd/$(SERVICE_NAME) $(SERVICE_PATH)
	systemctl daemon-reload
	systemctl unmask $(SERVICE_NAME)
	systemctl enable $(SERVICE_NAME)
	systemctl start $(SERVICE_NAME)

uninstall:
	systemctl stop $(SERVICE_NAME) || true
	systemctl disable $(SERVICE_NAME) || true
	rm -f /usr/bin/$(BINARY_NAME)
	rm -f $(SERVICE_PATH)
	systemctl daemon-reload

.PHONY: all build test clean run install_plugins uninstall_plugins install uninstall
