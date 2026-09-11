.PHONY: test test-race vet verify run compose-up compose-down

test:
	go test ./...

test-race:
	go test -race ./...

vet:
	go vet ./...

verify: test-race vet

run:
	go run ./cmd/devops-may-cry

compose-up:
	docker compose up --build -d

compose-down:
	docker compose down -v
