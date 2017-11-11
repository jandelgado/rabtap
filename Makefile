# rabtap makefile

BINARY_WIN64=rabtap-win-amd64.exe
BINARY_DARWIN64=rabtap-darwin-amd64
BINARY_LINUX64=rabtap-linux-amd64

build:	lint $(BINARY_LINUX64)

build-all:	build $(BINARY_WIN64)  $(BINARY_DARWIN64)

$(BINARY_DARWIN64): *.go app/main/*.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o $(BINARY_DARWIN64) app/main/*.go

$(BINARY_LINUX64): *.go app/main/*.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(BINARY_LINUX64) app/main/*.go

$(BINARY_WIN64): *.go app/main/*.go
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o $(BINARY_WIN64) app/main/*.go

lint:
	@./pre-commit

short-test: lint
	go test -v  -cover -coverprofile=coverage.out github.com/jandelgado/rabtap 
	go test -v  -cover -coverprofile=coverage_app.out github.com/jandelgado/rabtap/app/main
	grep -v "^mode:" coverage_app.out >> coverage.out
	go tool cover -func=coverage.out

test: 
	go test -race -v -tags "integration" -cover -coverprofile=coverage.out github.com/jandelgado/rabtap 
	go test -race -v -tags "integration" -cover -coverprofile=coverage_app.out github.com/jandelgado/rabtap/app/main
	grep -v "^mode:" coverage_app.out >> coverage.out
	go tool cover -func=coverage.out

# run rabbitmq server for integration test using docker container.
run-server:
	sudo docker run -ti --rm -p 5672:5672 \
		        -p 15672:15672 rabbitmq:3-management

dist-clean: clean
	rm -f *.out $(BINARY_WIN64) $(BINARY_LINUX64) $(BINARY_DARWIN64)

clean:
	go clean -r

