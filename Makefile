.PHONY: container/image/integration
container/image/integration:
	docker build -t lax:integration -f Dockerfile.integration .

.PHONY: build
build: container/image/integration
	docker run -w /app -v go-mod-cache:/go/pkg/mod -v $(PWD):/app -it lax:integration bash -c 'rm -rf lax; go build -buildvcs=false -o lax ./cmd/lax'

.PHONY: tests/unit
tests/unit:
	docker run -w /app -v go-mod-cache:/go/pkg/mod -v $(PWD):/app -it golang:1.22.3 bash -c 'go test ./... -tags "unit"'

.PHONY: tests/integration
tests/integration: build
	docker run -w /app -v go-mod-cache:/go/pkg/mod -v $(PWD):/app -it lax:integration bash -c 'go test -v -buildvcs=false ./tests/integration -tags "integration"'

.PHONY: tests/integration
tests/integration/nobuild:
	docker run -w /app -v go-mod-cache:/go/pkg/mod -v $(PWD):/app -it lax:integration bash -c 'go test -v -buildvcs=false ./tests/integration -tags "integration"'