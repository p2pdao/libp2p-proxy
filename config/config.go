package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	PeerKey   string        `json:"peer_key" yaml:"peer_key"`
	P2PHost   string        `json:"p2p_host" yaml:"p2p_host"`
	ServePath string        `json:"serve_path" yaml:"serve_path"`
	Network   NetworkConfig `json:"network" yaml:"network"`
	DHT       DHTConfig     `json:"dht" yaml:"dht"`
	ACL       ACLConfig     `json:"acl" yaml:"acl"`
	Proxy     *ProxyConfig  `json:"proxy" yaml:"proxy"`
}

type ProxyConfig struct {
	Addr       string `json:"addr" yaml:"addr"`
	ServerPeer string `json:"server_peer" yaml:"server_peer"`
}

type NetworkConfig struct {
	EnableNAT     bool     `json:"enable_nat" yaml:"enable_nat"`
	ListenAddrs   []string `json:"listen_addrs" yaml:"listen_addrs"`
	ExternalAddrs []string `json:"external_addrs" yaml:"external_addrs"`
	Relays        []string `json:"relays" yaml:"relays"`
}

type ACLConfig struct {
	AllowPeers   []string `json:"allow_peers" yaml:"allow_peers"`
	AllowSubnets []string `json:"allow_subnets" yaml:"allow_subnets"`
}

type DHTConfig struct {
	DatastorePath  string   `json:"datastore_path" yaml:"datastore_path"`
	BootstrapPeers []string `json:"bootstrap_peers" yaml:"bootstrap_peers"`
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
		ext := filepath.Ext(cfgPath)

		data, err := ioutil.ReadFile(cfgPath)
		if err != nil {
			panic(err)
		}

		err = parseConfig(data, ext, &cfg)
		if err != nil {
			panic(err)
		}
		if err != nil {
			return Config{}, err
		}
	}

	return cfg, nil
}

type unmarshaler func(data []byte, v interface{}) error

func parseConfig(data []byte, ext string, v interface{}) error {
	ext = strings.TrimLeft(ext, ".")

	var unmarshal unmarshaler

	switch ext {
	case "json":
		unmarshal = json.Unmarshal
	case "yaml", "yml":
		unmarshal = yaml.Unmarshal
	default:
		return fmt.Errorf("not supported config ext: %s", ext)
	}

	return unmarshal(data, v)
}
