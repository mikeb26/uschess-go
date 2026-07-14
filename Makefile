export GO111MODULE=on
export GOFLAGS=-mod=vendor

# Each direct child of examples/ containing main.go is built as an executable
# named after its directory (for example, examples/member/member).
EXAMPLE_DIRS := $(patsubst %/,%,$(dir $(wildcard examples/*/main.go)))
EXAMPLE_BINS := $(foreach dir,$(EXAMPLE_DIRS),$(dir)/$(notdir $(dir)))

.PHONY: build
build: client.gen.go vendor examples
	go build ./...

.PHONY: examples
examples: $(EXAMPLE_BINS)

examples/%: FORCE
	go build -o $@ ./$(@D)

vendor: go.mod
	go mod download
	go mod vendor

test: build FORCE
	go test ./...

unit-tests.xml: FORCE
	gotestsum --junitfile unit-tests.xml ./...

swagger.json:
	wget https://ratings-api.uschess.org/swagger/v1/swagger.json

client.gen.go: vendor client.go swagger.json oapi-codegen.yaml oapi-codegen.overlay.yaml
	go generate ./

.PHONY: deps
deps:
	rm -rf go.mod go.sum vendor swagger.json
	go mod init github.com/mikeb26/uschess-go
	GOPROXY=direct go mod tidy
	go get -tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
	wget https://ratings-api.uschess.org/swagger/v1/swagger.json
	go mod vendor
	rm -f client.gen.go

.PHONY: clean
clean:
	rm -f unit-tests.xml $(EXAMPLE_BINS)

FORCE:
