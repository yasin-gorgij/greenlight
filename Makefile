.DEFAULT_GOAL := build

BUILD_DIR=build

.PHONY:fmt
fmt:
	go fmt ./...

.PHONY:vet
vet: fmt
	go vet ./...

.PHONY:staticcheck
staticcheck: vet
	staticcheck ./...

.PHONY:revive
revive: staticcheck
	revive ./...

.PHONY:lint
lint: revive
	golangci-lint run

.PHONY:vulcheck
vulcheck: lint
	govulncheck ./...

.PHONY:test
test: vulcheck
	go test -race -vet=off ./...

.PHONY:build
build: test
	[ -d $(BUILD_DIR) ] || mkdir -p $(BUILD_DIR)
	go build -race -o $(BUILD_DIR)/greenlight ./cmd/api

.PHONY:clean
clean:
	rm -rf $(BUILD_DIR)

.PHONY:psql
psql:
	psql ${GREENLIGHT_DB_DSN}

.PHONY:migrations/up
migrations/up:
	@echo 'Running up migrations...'
	migrate -path ./migrations -database ${GREENLIGHT_DB_DSN} up

.PHONY: migrations/new
migrations/new:
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext=.sql -dir=./migrations ${name}
