.PHONY: container/image/integration
container/image/integration:
	docker build -t lax:integration -f Dockerfile.integration .

.PHONY: run
run: 
	docker run -w /app -v go-mod-cache:/go/pkg/mod -v $(PWD):/app golang:1.22.3 ./RUN.sh $(ARGS)

.PHONY: build
build: container/image/integration
	docker run -w /app -v go-mod-cache:/go/pkg/mod -v $(PWD):/app -t lax:integration \
	bash -c 'rm -rf lax; go build -buildvcs=false -o lax ./cmd/lax'

.PHONY: fmt
fmt: container/image/integration
	docker run -w /app -v go-mod-cache:/go/pkg/mod -v $(PWD):/app -t lax:integration \
	bash -c 'go fmt ./...'

.PHONY: debug
debug: container/image/integration
	docker run -w /app -v go-mod-cache:/go/pkg/mod -v $(PWD):/app -it lax:integration \
	bash $(ARGS)

.PHONY: tests/unit
tests/unit:
	docker run -w /app -v go-mod-cache:/go/pkg/mod -v $(PWD):/app -i lax:integration \
	bash -c 'go test -v -tags "!integration" ./...'

.PHONY: tests/unit/coverage
tests/coverage:
	docker run -w /app -v go-mod-cache:/go/pkg/mod -v $(PWD):/app -i lax:integration \
	bash -c 'go test -v -coverprofile cover.out -tags "!integration" ./... && go tool cover -html cover.out -o cover.html'

.PHONY: tests/integration
tests/integration: build
	docker run -w /app -v go-mod-cache:/go/pkg/mod -v $(PWD):/app -it lax:integration \
	bash -c 'go test -v -buildvcs=false -tags "integration" ./tests/integration'

.PHONY: tests/integration
tests/integration/nobuild:
	docker run -w /app -v go-mod-cache:/go/pkg/mod -v $(PWD):/app -it lax:integration \
	bash -c 'go test -v -buildvcs=false -tags "integration" ./tests/integration'
