.DEFAULT_GOAL := build

.PHONY:fmt vet staticcheck revive lint vulcheck build
fmt:
	go fmt ./...

vet: fmt
	go vet ./...

staticcheck: vet
	staticcheck ./...

revive: staticcheck
	revive ./...

lint: revive
	golangci-lint run

vulcheck: lint
	govulncheck ./...

build: vulcheck
	go build -o greenlight ./cmd/api

clean:
	rm -f greenlight
