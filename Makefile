run:
	go run ./cmd/lighthouse-tui/main.go

watch:
	watchexec -e go -r -- "make run"

test:
	go test -v ./...

lint:
	golangci-lint run

test-race:
	go test -v -race ./...

help:
	@echo "1. make run"
	@echo "2. make test"
	@echo "3. make lint"
	@echo "4. make test-race"
