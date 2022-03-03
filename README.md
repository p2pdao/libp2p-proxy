# libp2p-proxy
> HTTP proxy service with libp2p

In order to proxy an HTTP request, we create a local peer which listens on `localhost:1082`. HTTP requests performed to that address are tunneled via a libp2p stream to a remote peer, which then performs the HTTP requests and sends the response back to the local peer, which relays it to the user.

## Build

```
> cd libp2p-proxy
> go build
```

## Usage

First run the "remote" peer as follows. It will print a local peer address. If you would like to run this on a separate machine, please replace the IP accordingly:

```sh
> ./libp2p-proxy
Proxy server is ready
libp2p-peer addresses:
/ip4/127.0.0.1/tcp/12000/p2p/QmddTrQXhA9AkCpXPTkcY7e22NK73TwkUms3a44DhTKJTD
```

Then run the local peer, indicating that it will need to forward http requests to the remote peer as follows:

```
> ./libp2p-proxy -d /ip4/127.0.0.1/tcp/12000/p2p/QmddTrQXhA9AkCpXPTkcY7e22NK73TwkUms3a44DhTKJTD
Proxy server is ready
libp2p-peer addresses:
/ip4/127.0.0.1/tcp/12001/p2p/Qmaa2AYTha1UqcFVX97p9R1UP7vbzDLY7bqWsZw1135QvN
proxy listening on  127.0.0.1:1082
```

As you can see, the proxy prints the listening address `127.0.0.1:1082`. You can now use this address as a proxy, for example with `curl`:

```
> http_proxy=127.0.0.1:11082 curl "https://ipfs.io/p2p/QmfUX75pGRBRDnjeoMkQzuQczuCup2aYbeLxz5NzeSu9G6"
```
