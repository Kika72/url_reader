test:
	go test -v ./...

build:
	go build -o .build/url-reader cmd/main.go
