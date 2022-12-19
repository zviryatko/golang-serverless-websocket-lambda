.PHONY: build clean deploy

build:
	go mod vendor
	env GOARCH=amd64 GOOS=linux go build -ldflags="-s -w" -o bin/app main.go

clean:
	rm -rf ./bin ./vendor

deploy: clean build
	sls deploy --verbose
