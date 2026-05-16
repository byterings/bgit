FROM golang:1.24-bookworm

RUN apt-get update \
    && apt-get install -y --no-install-recommends git openssh-client ca-certificates bash make \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /src
COPY . .

ENV HOME=/tmp/bgit-home
ENV GOCACHE=/tmp/bgit-go-cache
ENV PATH=/usr/local/go/bin:/usr/local/bin:/usr/bin:/bin

RUN mkdir -p "$HOME" \
    && go build -o /usr/local/bin/bgit .

ENTRYPOINT ["bash", "tests/e2e/run.sh"]
