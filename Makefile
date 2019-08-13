# rabtap makefile

BINARY_WIN64=bin/rabtap-win-amd64.exe
BINARY_DARWIN64=bin/rabtap-darwin-amd64
BINARY_LINUX64=bin/rabtap-linux-amd64
SOURCE=$(shell find . -name "*go" -a -not -path "./vendor/*" -not -path "./cmd/testgen/*" )
VERSION=$(shell git describe --tags)
TOXICMD:=docker-compose exec toxiproxy /go/bin/toxiproxy-cli

.PHONY: test-app test-lib build build-all tags short-test test run-broker clean  toxiproxy-setup toxiproxy-cmd

build: build-linux

build-all:	build build-mac build-win

build-mac:
	cd cmd/rabtap && GO111MODULE=on CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags \
				"-X main.RabtapAppVersion=$(VERSION)" -o ../../$(BINARY_DARWIN64) 

build-linux:
	cd cmd/rabtap && GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags \
				"-X main.RabtapAppVersion=$(VERSION)" -o ../../$(BINARY_LINUX64) 

build-win:
	cd cmd/rabtap && GO111MODULE=on CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags \
				"-X main.RabtapAppVersion=$(VERSION)" -o ../../$(BINARY_WIN64) 

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
	sudo docker run -ti --rm -p 5672:5672 \
		        -p 15672:15672 rabbitmq:3-management

dist-clean: clean
	rm -f *.out $(BINARY_WIN64) $(BINARY_LINUX64) $(BINARY_DARWIN64)

clean:
	cd cmd/rabtap && go clean -r
	cd cmd/testgen && go clean -r

