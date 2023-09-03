.PHONY: dep
dep:
	@ go mod tidy && go mod verify

.PHONY: test
test:
	@ go test ./...

.PHONY: cover
cover:
	@ mkdir -p tmp
	@ go test -coverprofile=./tmp/cover.out ./...
	@ go tool cover -html=./tmp/cover.out

.PHONY: lint
lint:
	@ golangci-lint run --fix

.PHONY: build
build:
	@ go build -o ./connectbox-exporter .

.PHONY: run
run:
	@ ./connectbox-exporter

.PHONY: docker
docker:
	@ docker build -t ghcr.io/tetafro/connectbox-exporter .
