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
