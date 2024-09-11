package godhcpdconfig

import (
	"encoding/json"
	"net"
	"os"
	"testing"
)

func TestConfig(t *testing.T) {
	cfg := NewConfig()
	err := cfg.LoadFrom("/etc/dhcp/dhcpd.conf")
	if err != nil {
		panic(err)
	}
	cfg.Subnets["192.168.58.0"] = Subnet{
		Network: net.IPNet{
			IP:   net.ParseIP("192.168.58.0"),              // TODO
			Mask: net.IPMask(net.ParseIP("255.255.255.0")), // TODO
		},
		Options: Options{
			Range: Range{
				Start: net.ParseIP("192.168.58.100"), // TODO
				End:   net.ParseIP("192.168.58.200"), // TODO
			},
			DomainNameServers: NSs{
				NS{Host: "223.5.5.5"},       // TODO
				NS{Host: "114.114.114.114"}, // TODO
			},
			Routers: []string{
				"192.168.58.1",
			},
		},
	}
	cfg.Subnets["192.168.56.0"] = Subnet{
		Network: net.IPNet{
			IP:   net.ParseIP("192.168.56.0"),              // TODO
			Mask: net.IPMask(net.ParseIP("255.255.255.0")), // TODO
		},
		Options: Options{
			Range: Range{
				Start: net.ParseIP("192.168.56.100"), // TODO
				End:   net.ParseIP("192.168.56.200"), // TODO
			},
			DomainNameServers: NSs{
				NS{Host: "223.5.5.5"},       // TODO
				NS{Host: "114.114.114.114"}, // TODO
			},
			Routers: []string{
				"192.168.56.1",
			},
		},
	}
	delete(cfg.Subnets, "192.168.58.0")
	jsonEncoder := json.NewEncoder(os.Stdout)
	jsonEncoder.SetIndent("", "  ")
	jsonEncoder.Encode(cfg)
	cfg.ConfigWriteTo("/etc/dhcp/dhcpd.conf")
}
