package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ipfs/go-datastore"
	leveldb "github.com/ipfs/go-ds-leveldb"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-peerstore/pstoreds"
	"github.com/libp2p/go-libp2p/p2p/host/autorelay"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/p2pdao/libp2p-proxy/config"
	"github.com/p2pdao/libp2p-proxy/protocol"
)

const usage = `
libp2p-proxy creates a http and socks5 proxy service using two libp2p peers.
Generate two peer keys for server and client:
    ./libp2p-proxy -key
Update server.json with server peer key and start remote peer first with:
    ./libp2p-proxy -config server.json
Then update client.json with server peer multiaddress and start the local peer with:
    ./libp2p-proxy -config client.json

Then you can do something like:
    export http_proxy=http://127.0.0.1:1082 https_proxy=http://127.0.0.1:1082
or:
    export http_proxy=socks5://127.0.0.1:1082 https_proxy=socks5://127.0.0.1:1082
then:
    curl "https://github.com"
-------------------------------------------------------
Command flags:
`

func main() {
	// Parse some flags
	cfgPath := flag.String("config", "", "json configuration file; empty uses the default configuration")
	peerID := flag.String("peer", "", "proxy server peer address")
	proxyAddr := flag.String("addr", "", "proxy client address, default is 127.0.0.1:1082")
	help := flag.Bool("help", false, "show help info")
	genKey := flag.Bool("key", false, "generate a new peer private key")
	// version := flag.Bool("version", false, "show version info")
	flag.Parse()

	if *help {
		fmt.Println(usage)
		flag.PrintDefaults()
		os.Exit(0)
	}

	if *genKey {
		privKey, peerID, err := GeneratePeerKey()
		if err != nil {
			protocol.Log.Fatal(err)
		}
		fmt.Printf("Private Peer Key: %s\n", privKey)
		fmt.Printf("Public Peer ID: %s\n", peerID)
		os.Exit(0)
	}

	flag.Parse()
	cfg, err := config.LoadConfig(*cfgPath)
	if err != nil {
		protocol.Log.Fatal(err)
	}

	if peerID != nil && *peerID != "" {
		if cfg.Proxy == nil {
			cfg.Proxy = &config.ProxyConfig{}
		}
		cfg.Proxy.ServerPeer = *peerID
		if cfg.Proxy.Addr == "" {
			cfg.Proxy.Addr = "127.0.0.1:1082"
		}
		if proxyAddr != nil && *proxyAddr != "" {
			cfg.Proxy.Addr = *proxyAddr
		}
	}

	if cfg.PeerKey == "" {
		cfg.PeerKey, _, _ = GeneratePeerKey()
	}

	if cfg.P2PHost == "" {
		cfg.P2PHost = "p2p.to"
	}

	ctx := ContextWithSignal(context.Background())
	privk, err := ReadPeerKey(cfg.PeerKey)
	if err != nil {
		protocol.Log.Fatal(err)
	}

	var opts []libp2p.Option = []libp2p.Option{
		libp2p.Identity(privk),
		libp2p.UserAgent(protocol.ServiceName),
		libp2p.EnableRelay(),
		libp2p.EnableHolePunching(),
		libp2p.WithDialTimeout(time.Second * 60),
	}

	if len(cfg.Network.Relays) > 0 {
		cfg.Network.Relays = autorelay.DefaultRelays
		relays := make([]peer.AddrInfo, 0, len(cfg.Network.Relays))
		for _, addr := range cfg.Network.Relays {
			pi, err := peer.AddrInfoFromString(addr)
			if err != nil {
				protocol.Log.Fatal(fmt.Sprintf("failed to initialize default static relays: %s", err))
			}
			relays = append(relays, *pi)
		}
		opts = append(opts,
			libp2p.EnableAutoRelay(),
			libp2p.StaticRelays(relays),
		)
	}

	if cfg.ACL != nil {
		acl, err := protocol.NewACL(*cfg.ACL)
		if err != nil {
			protocol.Log.Fatal(err)
		}
		opts = append(opts, libp2p.ConnectionGater(acl))
	}

	if cfg.Proxy == nil || cfg.Proxy.ServerPeer == "" {
		// add DHT for server side
		var ds datastore.Batching
		if cfg.DHT.DatastorePath != "" {
			ds, err = leveldb.NewDatastore(cfg.DHT.DatastorePath, nil)
			if err != nil {
				protocol.Log.Fatal(err)
			}
		}
		if ds != nil {
			pds, err := pstoreds.NewPeerstore(ctx, ds, pstoreds.DefaultOpts())
			if err != nil {
				protocol.Log.Fatal(err)
			}
			opts = append(opts, libp2p.Peerstore(pds))
		}

		opts = append(opts,
			libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
				dhtpeers := dht.GetDefaultBootstrapPeerAddrInfos()
				if len(cfg.DHT.BootstrapPeers) > 0 {
					for _, s := range cfg.DHT.BootstrapPeers {
						if peer, err := peer.AddrInfoFromString(s); err == nil {
							dhtpeers = append(dhtpeers, *peer)
						}
					}
				}

				dhtopts := []dht.Option{
					dht.Mode(dht.ModeClient),
					dht.BootstrapPeers(dhtpeers...),
				}
				if ds != nil {
					dhtopts = append(dhtopts, dht.Datastore(ds))
				}

				idht, err := dht.New(ctx, h, dhtopts...)
				if err == nil {
					idht.Bootstrap(ctx)
				}
				return idht, err
			}),
		)
	}

	if cfg.Proxy == nil {
		opts = append(opts,
			libp2p.ListenAddrStrings(cfg.Network.ListenAddrs...),
		)

		if cfg.Network.EnableNAT {
			opts = append(opts,
				libp2p.NATPortMap(),
				libp2p.EnableNATService(),
			)
		}

		if len(cfg.Network.ExternalAddrs) > 0 {
			emas := make([]ma.Multiaddr, 0, len(cfg.Network.ExternalAddrs))
			for _, ad := range cfg.Network.ExternalAddrs {
				if addr, err := ma.NewMultiaddr(ad); err == nil {
					emas = append(emas, addr)
				}
			}

			if len(emas) > 0 {
				opts = append(opts,
					libp2p.ForceReachabilityPublic(),
					libp2p.AddrsFactory(func(_ []ma.Multiaddr) []ma.Multiaddr {
						return emas
					}),
				)
			}
		}

		host, err := libp2p.New(opts...)
		if err != nil {
			protocol.Log.Fatal(err)
		}

		fmt.Printf("Peer ID: %s\n", host.ID())
		fmt.Printf("Peer Addresses:\n")
		for _, addr := range host.Addrs() {
			fmt.Printf("\t%s/p2p/%s\n", addr, host.ID())
		}

		ping.NewPingService(host)
		proxy := protocol.NewProxyService(ctx, host, cfg.P2PHost)

		if cfg.ServePath != "" {
			if err := proxy.ServeHTTP(static(cfg.ServePath), nil); err != nil {
				protocol.Log.Fatal(err)
			}
		} else {
			if err := proxy.Wait(nil); err != nil {
				protocol.Log.Fatal(err)
			}
		}

	} else {
		opts = append(opts,
			libp2p.NoListenAddrs,
		)
		host, err := libp2p.New(opts...)
		if err != nil {
			protocol.Log.Fatal(err)
		}

		fmt.Printf("Peer ID: %s\n", host.ID())
		serverPeer := &peer.AddrInfo{ID: host.ID()}
		if cfg.Proxy.ServerPeer != "" {
			serverPeer, err = peer.AddrInfoFromString(cfg.Proxy.ServerPeer)
			if err != nil {
				protocol.Log.Fatal(err)
			}

			ctxt, cancel := context.WithTimeout(ctx, time.Second*10)
			if err = host.Connect(ctxt, *serverPeer); err != nil {
				protocol.Log.Fatal(err)
			}

			res := <-ping.Ping(ctx, host, serverPeer.ID)
			if res.Error != nil {
				protocol.Log.Fatal(res.Error)
			}
			cancel()
			fmt.Printf("Ping Server RTT: %s\n", res.RTT)
		}

		proxy := protocol.NewProxyService(ctx, host, cfg.P2PHost)
		fmt.Printf("Proxy Address: %s\n", cfg.Proxy.Addr)
		if err := proxy.Serve(cfg.Proxy.Addr, serverPeer.ID); err != nil {
			protocol.Log.Fatal(err)
		}
	}
}

func ContextWithSignal(ctx context.Context) context.Context {
	newCtx, cancel := context.WithCancel(ctx)
	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signals
		cancel()
	}()
	return newCtx
}

type static string

func (s static) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, string(s))
}
