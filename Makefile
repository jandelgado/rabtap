# rabtap makefile
SHELL:=/bin/bash

SOURCE=$(shell find . -name "*go" -a -not -path "./vendor/*" -not -path "./cmd/testgen/*" )
INFO_VERSION=$(shell git describe --tags)
INFO_COMMIT=$(shell git rev-parse --short HEAD)
INFO_GO_VERSION:=$(shell go version | awk '{print $$3}')
INFO_BUILD_DATE:=$(shell date -R)
TOXICMD:=docker compose exec toxiproxy /go/bin/toxiproxy-cli
.PHONY: phony
LDFLAGS=-s -w -X main.BuildVersion=$(INFO_VERSION) \
        -X 'main.BuildDate=$(INFO_BUILD_DATE)' \
        -X 'main.BuildGoVersion=$(INFO_GO_VERSION)' \
        -X 'main.BuildCommit=$(INFO_COMMIT)'
build: phony
	GOTOOLCHAIN=auto CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o ./bin/rabtap ./cmd/rabtap
wasm-build: phony
	GOTOOLCHAIN=auto CGO_ENABLED=1 GOOS=wasip1 GOARCH=wasm go build -o ./bin/rabtap-wasm ./cmd/rabtap

tags: $(SOURCE)
	@gotags -f tags $(SOURCE)

lint: phony
	golangci-lint run

short-test:  phony
	go test -v $(TESTOPTS) -race  github.com/jandelgado/rabtap/cmd/rabtap
	go test -v $(TESTOPTS) -race  github.com/jandelgado/rabtap/pkg

test-app: phony
	go test -race -v -tags "integration" $(TESTOPTS) -cover -coverprofile=coverage_app.out github.com/jandelgado/rabtap/cmd/rabtap

test-lib: phony
	go test -race -v -tags "integration" $(TESTOPTS) -cover -coverprofile=coverage.out github.com/jandelgado/rabtap/pkg

test: test-app test-lib
	grep -v "^mode:" coverage_app.out >> coverage.out
	go tool cover -func=coverage.out

# docker-compose up must be first called. Then create a proxy with
# this target and connect to to localhost:55672 (amqp).
toxiproxy-setup: phony
	$(TOXICMD) c amqp --listen :55672 --upstream rabbitmq:5672 || true

# call with e.g. 
# make toxiproxy-cmd         -- show help
# make TOXIARGS="toggle amqp"  -- toggle amqp proxy
toxiproxy-cmd: phony
	$(TOXICMD) $(TOXIARGS)

# run rabbitmq server for integration test using docker container.
run-broker: phony
	cd inttest/pki && ./mkcerts.sh
	cd inttest/rabbitmq && docker compose up

dist-clean: clean
	rm -rf *.out bin/ dist/

clean: phony
	go clean -r ./cmd/rabtap
	go clean -r ./cmd/testgen

# create a test-release v99.9.9 to test the ci pipeline
gh-make-test-release: phony
	gh release delete v99.9.9 --cleanup-tag --yes || true
	gh release create v99.9.9 --notes "testing" --title "v99.9.9 testing" --prerelease --target update_dependencies

phony: