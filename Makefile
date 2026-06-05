.PHONY: test test-integration test-docker test-backup-portability test-real desktop-dev desktop-build

test:
	GOCACHE=/tmp/bgit-gocache go test ./...

test-integration:
	bash tests/integration.sh

test-docker:
	docker build -f Dockerfile.test -t bgit-test .
	docker run --rm bgit-test

test-backup-portability:
	bash tests/backup-portability.sh

test-real:
	bash tests/real-accounts.sh

desktop-dev:
	cd desktop && wails dev

desktop-build:
	cd desktop && wails build
