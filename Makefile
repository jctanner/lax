.PHONY: container/image/integration
container/image/integration:
	docker build -t lax:integration -f Dockerfile.integration .

.PHONY: build
build: container/image/integration
	docker run -w /app -v go-mod-cache:/go/pkg/mod -v $(PWD):/app -it lax:integration bash -c 'rm -rf lax; go build -buildvcs=false -o lax ./cmd/lax'

.PHONY: tests/unit
tests/unit:
	docker run -w /app -v go-mod-cache:/go/pkg/mod -v $(PWD):/app -it lax:integration bash -c 'go test -v -coverprofile cover.out -tags "unit" ./...'

.PHONY: tests/coverage
tests/coverage:
	docker run -w /app -v go-mod-cache:/go/pkg/mod -v $(PWD):/app -it lax:integration bash -c 'go tool cover -html cover.out -o cover.html'

.PHONY: tests/integration
tests/integration: build
	docker run -w /app -v go-mod-cache:/go/pkg/mod -v $(PWD):/app -it lax:integration bash -c 'go test -v -buildvcs=false -tags "integration" ./tests/integration'

.PHONY: tests/integration
tests/integration/nobuild:
	docker run -w /app -v go-mod-cache:/go/pkg/mod -v $(PWD):/app -it lax:integration bash -c 'go test -v -buildvcs=false -tags "integration" ./tests/integration'