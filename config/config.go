package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	PeerKey   string        `json:"peer_key"`
	P2PHost   string        `json:"p2p_host"`
	ServePath string        `json:"serve_path"`
	Network   NetworkConfig `json:"network"`
	ACL       ACLConfig     `json:"acl"`
	Proxy     *ProxyConfig  `json:"proxy"`
}

type ProxyConfig struct {
	Addr       string `json:"addr"`
	ServerPeer string `json:"server_peer"`
}

type NetworkConfig struct {
	EnableNAT   bool     `json:"enable_nat"`
	ListenAddrs []string `json:"listen_addrs"`
	Relays      []string `json:"relays"`
}

type ACLConfig struct {
	AllowPeers   []string `json:"allow_peers"`
	AllowSubnets []string `json:"allow_subnets"`
}

func Default() Config {
	return Config{
		Network: NetworkConfig{
			ListenAddrs: []string{
				"/ip4/127.0.0.1/tcp/4001",
				"/ip6/::1/tcp/4001",
			},
		},
		ACL: ACLConfig{
			AllowPeers:   []string{},
			AllowSubnets: []string{"127.0.0.1/32", "::1/128"},
		},
	}
}

func LoadConfig(cfgPath string) (Config, error) {
	cfg := Default()

	if cfgPath != "" {
		cfgFile, err := os.Open(cfgPath)
		if err != nil {
			return Config{}, err
		}
		defer cfgFile.Close()

		decoder := json.NewDecoder(cfgFile)
		err = decoder.Decode(&cfg)
		if err != nil {
			return Config{}, err
		}
	}

	return cfg, nil
}
