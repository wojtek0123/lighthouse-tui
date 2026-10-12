run:
	go run ./cmd/lighthouse-tui/main.go

dev:
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "'air' is not installed. Installation..."; \
		go install github.com/cosmtrek/air@latest; \
		air; \
	fi
