.PHONY: run
run:
	docker compose -f ./.ci/docker-compose/integration-test.yml up

.PHONY: lint
lint:
	golangci-lint run --config=.golangci.yml

.PHONY: test
test:
	go test ./... -coverprofile=coverage.txt -covermode count
	go get github.com/boumenot/gocover-cobertura
	go run github.com/boumenot/gocover-cobertura < coverage.txt > coverage.xml
	go test -race ./...
	go tool cover -func coverage.txt

