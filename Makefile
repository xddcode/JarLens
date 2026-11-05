.PHONY: check fmt tidy vet staticcheck test race integration security build dev

check: fmt tidy vet staticcheck test

fmt:
	@files=$$(git ls-files --cached --others --exclude-standard '*.go'); \
		unformatted=$$(gofmt -l $$files); \
		if [ -n "$$unformatted" ]; then echo "$$unformatted"; exit 1; fi

tidy:
	go mod tidy -diff

vet:
	go vet ./...

staticcheck:
	go tool staticcheck ./...

test:
	go test -cover ./... -count=1

race:
	go test -race -p 1 ./... -count=1

integration:
	go test -tags integration ./... -count=1

security:
	go tool govulncheck ./...

build:
	go tool viva build

dev:
	go tool viva dev .
