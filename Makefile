.PHONY: run
run:
	docker compose -f ./.ci/docker-compose/integration-test.yml up
