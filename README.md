# libp2p-proxy
> http&socks5 proxy service on libp2p peers.

libp2p-proxy creates a http and socks5 proxy service using libp2p peers.

libp2p-proxy can be used to access http service running on libp2p peers.

Client & Server Mode:
```
                                                                               XXX XXX XX
                                                                           XXXX         XXX
                 +--------+               +--------+                  XXXXX                XX
                 |        |               |        | HTTP or HTTPS   X                    XX
      TCP stream | Proxy  | libp2p stream | Proxy  <---------------> X  <THE INTERNET>   XX
App  <-----------> Client <---------------> Server |                XX   HTTP SERVER      XX
  http tunnel -> |        |               |        <---------------> X                     XX
socks5 tunnel -> | Local  |               | Remote | libp2p stream    XXX XXXX XXX        XX
                 +--------+               +--------+                               XXXX XXX
```

Standalone Mode:
```
                                                       XXX XXX XX
                                                   XXXX         XXX
                 +----------+                  XXXXX                XX
                 |  Proxy   | HTTP or HTTPS   X                    XX
      TCP stream |  Client  <---------------> X  <THE INTERNET>   XX
App  <----------->  Server  |                XX   HTTP SERVER      XX
  http tunnel -> |          <---------------> X                     XX
socks5 tunnel -> |  Local   | libp2p stream    XXX XXXX XXX        XX
                 +----------+                              XXXX XXX
```

## Install

```
go install github.com/p2pdao/libp2p-proxy/cmd/libp2p-proxy@latest
```

## Usage

### Configure proxy
The default listen addr for proxy is `127.0.0.1:1082`, so you can configure:
```
http://127.0.0.1:1082
```
or:
```
socks5://127.0.0.1:1082
```

On terminal, you can configure:
```sh
export http_proxy=http://127.0.0.1:1082 https_proxy=http://127.0.0.1:1082
```
or:
```sh
export http_proxy=socks5://127.0.0.1:1082 https_proxy=socks5://127.0.0.1:1082
```

We recommend using https://github.com/FelisCatus/SwitchyOmega on a browser.

Access a normal website with proxy:
```
https://www.google.com/
```

Access a p2p website with proxy:
```
http://p2p.to/p2p/12D3KooWE8HTd1GrfGLtEg3GTfea61EPBA5UPM77tevBsj9QAxYz/http/
http://p2p.to/p2p/12D3KooWE8HTd1GrfGLtEg3GTfea61EPBA5UPM77tevBsj9QAxYz/http/metadata
```

### Full config file

https://github.com/p2pdao/libp2p-proxy/blob/main/config/config_sample_full.yaml
```yaml
# `peer_key` is client & server side config, it is the peer's private key for running,
# you can generate key-pair by runing `libp2p-proxy -key`
# if omit, it will generate one randomly.
peer_key: "CAESQLcvtmSITUktckPrPSOQuTSPjTBBO7/FW3m5N1qnTfBv9ilHJ7GknXc/AKLaiekjqlm/STh97MDPTV8nkl4aRfM="
# `p2p_host` is client side config.
proxy:
  # `addr` is listen addr for proxy, it support http and socks5:
  addr: "127.0.0.1:1082"
  # `server_peer` is proxy server that client connect to.
  # default to empty, that means the libp2p-proxy will run in standalone mode!
  server_peer: "/ip4/127.0.0.1/tcp/11211/p2p/12D3KooWSPGy9bCrTRF5Nwsb3B6CQsZ9VGvEGPJ6ZT2ZWWCTXR3p"
# `p2p_host` is server side config, used to distinguish between normal websites and p2p websites.
# defaut to "p2p.to", for example:
# access a normal website: https://www.google.com/
# access a p2p website: http://p2p.to/p2p/12D3KooWE8HTd1GrfGLtEg3GTfea61EPBA5UPM77tevBsj9QAxYz/http/
p2p_host: "p2p.to"
# `serve_path` is server side config, used to run a http satic service on libp2p streams.
# it is a static files directory, defaut to "", not running a http service.
serve_path: "./my-local-static-website-directory"
# `network` is server side config.
network:
  # `enable_nat` will enable nat service, default to false.
  enable_nat: false
  # default to ["/ip4/127.0.0.1/tcp/4001", "/ip6/::1/tcp/4001"]
  # invalid addr will be ignored, such as port conflicting, IPv6 not supporting...
  listen_addrs:
    - "/ip4/0.0.0.0/udp/11211/quic" # QUIC & HTTP3 transport on IPv4
    - "/ip6/::/udp/11211/quic" # QUIC & HTTP3 transport on IPv6
    - "/ip4/0.0.0.0/tcp/11211"
    - "/ip6/::/tcp/11211"
    - "/ip4/0.0.0.0/tcp/11212/ws" # websocket transport on IPv4
    - "/ip6/::/tcp/11212/ws"
  # configures external peer addrs for public accessing.
  external_addrs:
    - "/ip4/1.2.3.4/udp/11211/quic"
    - "/ip4/1.2.3.4/tcp/11211"
    - "/ip4/1.2.3.4/tcp/11212/ws"
  # configures known relays for autorelay; when this option is enabled
  # then the system will use the configured relays instead of querying the DHT to discover relays.
  # default to empty, that means not use relays.
  relays:
    - "/ip4/147.75.70.221/tcp/4001/p2p/Qme8g49gm3q4Acp7xWBKg3nAa9fxZ1YmyDJdyGgoG6LsXh"
# `acl` is server side config.
acl:
  # `allow_peers` is a white list of allowed client side peers to access
  # default to empty, that means allow all.
  allow_peers: ["12D3KooWAMspLEqdE79kAuvMAmPNHeJdJGTpKb7rEmksrQodhU62"]
  # `allow_subnets` is a white list of allowed subnets that client side peers to access
  # default to ["127.0.0.1/32", "::1/128"], that means only allowing local peers.
  allow_subnets: []
# `dht` is server side config, run DHT client to find peers.
dht:
  # `datastore_path` configures a directory for storing data.
  # default to empty, that means using memory instead.
  datastore_path: "./datastore"
  # `bootstrap_peers` configures additional peers to connect to.
  # default to empty, that means using https://github.com/libp2p/go-libp2p-kad-dht/blob/master/dht_bootstrap.go#L25.
  bootstrap_peers:
    - "/ip4/104.131.131.82/tcp/4001/p2p/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ"
```

### Generate key pair for server or client:

```sh
libp2p-proxy -key
```

### Run a standalone mode peer:
It will run client & server side together!

standalone.yaml:
```yaml
peer_key: "CAESQBa/lNg0/GHhzjf03oYvHfDYf9VnkQImE9lPB8Zrf4JICBKHPB5PbIzQoCkWwBrkha4xgpIerre4B5zZ5J7f/W8="
proxy:
  addr: "127.0.0.1:1082"
  server_peer: "" # keep it empty!
dht:
  datastore_path: "./datastore"
```

then:
```
libp2p-proxy -config standalone.yaml
```

### Run a server side peer:
server.yaml:
```yaml
peer_key: "CAESQLcvtmSITUktckPrPSOQuTSPjTBBO7/FW3m5N1qnTfBv9ilHJ7GknXc/AKLaiekjqlm/STh97MDPTV8nkl4aRfM="
network:
  listen_addrs:
    - "/ip4/0.0.0.0/udp/11211/quic"
    - "/ip4/0.0.0.0/tcp/11211"
    - "/ip4/0.0.0.0/tcp/11212/ws"
acl:
  allow_peers:
    - "12D3KooWAMspLEqdE79kAuvMAmPNHeJdJGTpKb7rEmksrQodhU62" # only allow this peer to access.
  allow_subnets: []
dht:
  datastore_path: "./datastore"
```

then:
```
libp2p-proxy -config server.yaml
```

### Run a client side peer:
client.yaml:
```yaml
peer_key: "CAESQBa/lNg0/GHhzjf03oYvHfDYf9VnkQImE9lPB8Zrf4JICBKHPB5PbIzQoCkWwBrkha4xgpIerre4B5zZ5J7f/W8="
proxy:
  addr: "127.0.0.1:1082"
  server_peer: "/ip4/{ACCESSIBLE_IP}/tcp/11211/p2p/12D3KooWSPGy9bCrTRF5Nwsb3B6CQsZ9VGvEGPJ6ZT2ZWWCTXR3p"
```

then:
```
libp2p-proxy -config client.yaml
```

### Run a server side peer with HTTP static service:
server_static.yaml:
```yaml
peer_key: "CAESQLcvtmSITUktckPrPSOQuTSPjTBBO7/FW3m5N1qnTfBv9ilHJ7GknXc/AKLaiekjqlm/STh97MDPTV8nkl4aRfM="
serve_path: "./static-website-directory"
network:
  listen_addrs:
    - "/ip4/0.0.0.0/udp/11211/quic"
    - "/ip4/0.0.0.0/tcp/11211"
    - "/ip4/0.0.0.0/tcp/11212/ws"
acl:
  allow_peers:
    - "12D3KooWAMspLEqdE79kAuvMAmPNHeJdJGTpKb7rEmksrQodhU62" # only allow this peer to access.
  allow_subnets: []
dht:
  datastore_path: "./datastore"
```

then:
```
libp2p-proxy -config server_static.yaml
```