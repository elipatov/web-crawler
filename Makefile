.PHONY: run
run:
	docker-compose -f ./.ci/docker-compose/integration-test.yml up --build --scale app=3 -d

.PHONY: run0
run0:
	docker-compose -f ./.ci/docker-compose/integration-test.yml up --build --scale app=0 --scale nginx=0 -d

.PHONY: run1
run1:
	docker-compose -f ./.ci/docker-compose/integration-test.yml up --build --scale app=1 -d

.PHONY: run2
run2:
	docker-compose -f ./.ci/docker-compose/integration-test.yml up --build --scale app=2 -d

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

