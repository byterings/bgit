.PHONY: test e2e test-all

GO_CACHE ?= /tmp/bgit-go-cache
E2E_IMAGE ?= bgit-e2e:local
DOCKER ?= docker

test:
	GOCACHE=$(GO_CACHE) go test ./...

e2e:
	@$(DOCKER) info >/dev/null 2>&1 || { \
		echo "Docker daemon is not accessible."; \
		echo ""; \
		echo "Fix one of these, then rerun: make e2e"; \
		echo "  1. Start Docker Desktop / Docker Engine"; \
		echo "  2. Add your user to the docker group: sudo usermod -aG docker $$USER"; \
		echo "  3. Open a new shell or run: newgrp docker"; \
		echo "  4. Or run with sudo Docker access: make e2e DOCKER='sudo docker'"; \
		exit 1; \
	}
	$(DOCKER) build -f tests/e2e/Dockerfile -t $(E2E_IMAGE) .
	$(DOCKER) run --rm $(E2E_IMAGE)

test-all: test e2e
