.PHONY: test test-desktop test-integration test-docker test-backup-portability test-real desktop-dev desktop-build desktop-harness

test:
	GOCACHE=/tmp/bgit-gocache go test ./...

# Desktop backend tests. The devharness tag builds the desktop package without
# the Wails window, so no WebKit/GTK libraries are needed.
test-desktop:
	GOCACHE=/tmp/bgit-gocache go test -tags devharness ./desktop/

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

# Serves the desktop frontend to a browser with the real backend on an
# isolated home directory: http://127.0.0.1:34116
desktop-harness:
	go run -tags devharness ./desktop -home /tmp/bgit-harness-home
