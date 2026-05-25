.PHONY: test test-integration test-docker test-real

test:
	GOCACHE=/tmp/bgit-gocache go test ./...

test-integration:
	bash tests/integration.sh

test-docker:
	docker build -f Dockerfile.test -t bgit-test .
	docker run --rm bgit-test

test-real:
	bash tests/real-accounts.sh
