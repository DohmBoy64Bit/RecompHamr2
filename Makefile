.PHONY: verify test docscheck archcheck covercheck

verify: docscheck archcheck covercheck

test:
	go test ./...

covercheck:
	go run ./internal/covercheck

archcheck:
	go run ./internal/archcheck

docscheck:
	go run ./internal/docscheck
