.PHONY: build build-darwin build-linux
build:
	@mkdir -p ./dist
	go build -o ./dist/libp2p-proxy ./cmd/libp2p-proxy
build-darwin:
	@mkdir -p ./dist/darwin
	GO111MODULE=on CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o ./dist/darwin/libp2p-proxy ./cmd/libp2p-proxy
build-linux:
	@mkdir -p ./dist/linux
	GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./dist/linux/libp2p-proxy ./cmd/libp2p-proxy