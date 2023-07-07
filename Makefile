build:
	@CGO_ENABLED=0 go build -o out/fs

run:
	@go run --race main.go
