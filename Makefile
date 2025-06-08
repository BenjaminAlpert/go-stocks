build:
	go build -o ./bin/go-stocks

build-and-run: build
	./bin/go-stocks