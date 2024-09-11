package godhcpdconfig

import (
	"fmt"
	"godhcpdconfig/isccfg"
	"io"
	"net"
	"sort"
	"strings"
)

type ToSubnetI interface {
	ToSubnet() Subnet
}

type Subnet struct {
	Network net.IPNet
	Options Options
}

func NewSubnet() *Subnet {
	return &Subnet{
		Options: Options{
			Custom: CustomOptions{},
		},
	}
}

type Subnets map[string]Subnet

func (subnet Subnet) ConfigWrite(out io.Writer, root Root) (err error) {
	_, err = fmt.Fprintf(out, "\nsubnet %v netmask %v {\n", subnet.Network.IP, net.IP(subnet.Network.Mask).String())
	if err != nil {
		return err
	}

	// A hacky workaround for a bug of https://github.com/xaionaro-go/fwsmConfig/blob/master/dhcp.go
	if len(subnet.Options.Routers) == 0 {
		router := subnet.Network.IP
		routerWords := strings.Split(router.String(), ".")
		routerWords[3] = "1"
		subnet.Options.Routers = append(subnet.Options.Routers, strings.Join(routerWords, "."))
	}
	err = subnet.Options.configWrite(out, root, "\t")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "}\n")
	if err != nil {
		return err
	}
	return nil
}

func (subnets Subnets) ConfigWrite(out io.Writer, root Root) error {
	var keys []string
	for k := range subnets {
		keys = append(keys, k)
	}
	sort.StringSlice(keys).Sort()

	for _, k := range keys {
		subnet := subnets[k]
		err := subnet.ConfigWrite(out, root)
		if err != nil {
			return err
		}
	}

	return nil
}

func (subnet Subnet) ToSubnet() Subnet {
	return subnet
}

func (subnets Subnets) ISet(subnetI ToSubnetI) error {
	subnet := subnetI.ToSubnet()
	if subnet.Network.IP.String() == (net.IP{}).String() {
		return fmt.Errorf("the subnet's IP is not set")
	}
	subnets[subnet.Network.IP.String()] = subnet
	return nil
}

func (subnet *Subnet) parse(root *Root, netStr string, cfgRaw *isccfg.Config) (err error) {
	var maskStr string
	cfgRaw, _ = cfgRaw.Unwrap()
	cfgRaw, maskStr = cfgRaw.Unwrap()

	subnet.Network.IP = net.ParseIP(netStr)
	subnet.Network.Mask = net.IPMask(net.ParseIP(maskStr))

	for k, v := range *cfgRaw {
		err = subnet.Options.parse(root, k, v)
		if err != nil {
			return
		}
	}

	return nil
}
