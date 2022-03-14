.PHONY: build build-all
build:
	@mkdir -p ./dist
	go build -o ./dist/libp2p-proxy ./cmd/libp2p-proxy
build-all:
	@mkdir -p ./dist
	GOOS=darwin GOARCH=amd64 go build -o ./dist/libp2p-proxy.darwin-amd64 ./cmd/libp2p-proxy
	GOOS=darwin GOARCH=arm64 go build -o ./dist/libp2p-proxy.darwin-arm64 ./cmd/libp2p-proxy
	GOOS=linux GOARCH=amd64 go build -o ./dist/libp2p-proxy.linux-amd64 ./cmd/libp2p-proxy
	GOOS=linux GOARCH=arm64 go build -o ./dist/libp2p-proxy.linux-arm64 ./cmd/libp2p-proxy
	GOOS=windows GOARCH=amd64 go build -o ./dist/libp2p-proxy.windows-amd64.exe ./cmd/libp2p-proxy
	GOOS=windows GOARCH=arm64 go build -o ./dist/libp2p-proxy.windows-arm64.exe ./cmd/libp2p-proxy
