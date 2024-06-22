build:
	rm -f lax
	go build -o lax ./cmd/lax

tests/integration: build
	go test ./tests/integration/simple_test.go -v