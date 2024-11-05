.DEFAULT_GOAL := build

BUILD_DIR=build

.PHONY:fmt vet staticcheck revive lint vulcheck build up psql
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
	[ -d $(BUILD_DIR) ] || mkdir -p $(BUILD_DIR)
	go build -race -o $(BUILD_DIR)/greenlight ./cmd/api

clean:
	rm -rf $(BUILD_DIR)

psql:
	psql ${GREENLIGHT_DB_DSN}

up:
	@echo 'Running up migrations...'
	migrate -path ./migrations -database ${GREENLIGHT_DB_DSN} up
