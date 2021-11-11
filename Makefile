# rabtap makefile

SOURCE=$(shell find . -name "*go" -a -not -path "./vendor/*" -not -path "./cmd/testgen/*" )
VERSION=$(shell git describe --tags)
TOXICMD:=docker-compose exec toxiproxy /go/bin/toxiproxy-cli

.PHONY: test-app test-lib build build tags short-test test run-broker clean dist-clean toxiproxy-setup toxiproxy-cmd

build:
	cd cmd/rabtap && GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags \
				"-s -w -X main.version=$(VERSION)" -o ../../bin/rabtap

tags: $(SOURCE)
	@gotags -f tags $(SOURCE)

lint:
	@./pre-commit

short-test: 
	go test -v -race  github.com/jandelgado/rabtap/cmd/rabtap
	go test -v -race  github.com/jandelgado/rabtap/pkg

test-app:
	go test -race -v -tags "integration" -cover -coverprofile=coverage_app.out github.com/jandelgado/rabtap/cmd/rabtap

test-lib:
	go test -race -v -tags "integration" -cover -coverprofile=coverage.out github.com/jandelgado/rabtap/pkg

test: test-app test-lib
	grep -v "^mode:" coverage_app.out >> coverage.out
	go tool cover -func=coverage.out

# docker-compose up must be first called. Then create a proxy with
# this target and connect to to localhost:55672 (amqp).
toxiproxy-setup:
	$(TOXICMD) c amqp --listen :55672 --upstream rabbitmq:5672 || true

# call with e.g. 
# make toxiproxy-cmd         -- show help
# make TOXIARGS="toggle amqp"  -- toggle amqp proxy
toxiproxy-cmd:
	$(TOXICMD) $(TOXIARGS)

# run rabbitmq server for integration test using docker container.
run-broker:
	cd inttest/pki && ./mkcerts.sh
	cd inttest/rabbitmq && docker-compose up

dist-clean: clean
	rm -rf *.out bin/ dist/

clean:
	cd cmd/rabtap && go clean -r
	cd cmd/testgen && go clean -r

