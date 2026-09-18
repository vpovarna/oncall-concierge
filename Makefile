.PHONY: fmt vet run

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

run: vet
	go run cmd/main.go web webui api
