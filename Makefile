run:
	go run ./cmd/lighthouse-tui/main.go

watch:
	watchexec -e go -r -- "make run"
