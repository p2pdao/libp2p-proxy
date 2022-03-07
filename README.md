# libp2p-proxy
> HTTP proxy service with libp2p

libp2p-proxy creates a http and socks5 proxy service using libp2p peers.

## Build

```
> cd libp2p-proxy
> make build
```
## Install

```
go get github.com/p2pdao/libp2p-proxy/cmd/libp2p-proxy
```

## Usage

1. Generate peer keys for server and client:
```
./libp2p-proxy -key
```

2. Update [server.json](https://github.com/p2pdao/libp2p-proxy/blob/main/config/sample_server.json) with server peer key and start remote peer first with:
```
./libp2p-proxy -config server.json
```

3. Then update [client.json](https://github.com/p2pdao/libp2p-proxy/blob/main/config/sample_client.json) with server peer multiaddress and start the local peer with:
```
./libp2p-proxy -config client.json
```

Then you can do something like:
```
export http_proxy=http://127.0.0.1:1082 https_proxy=http://127.0.0.1:1082
```

or:
```
export http_proxy=socks5://127.0.0.1:1082 https_proxy=socks5://127.0.0.1:1082
```

then:
```
curl "https://github.com"
```
