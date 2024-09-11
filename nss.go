package godhcpdconfig

import "net"

type NS net.NS
type NSs []NS

func (nss NSs) ToIPs() (result []net.IP) {
	for _, ns := range nss {
		result = append(result, net.ParseIP(ns.Host))
	}
	return
}
func (nss NSs) ToStrings() (result []string) {
	for _, ns := range nss {
		result = append(result, ns.Host)
	}
	return
}
func (nss NSs) ToNetNSs() (result []net.NS) {
	for _, ns := range nss {
		result = append(result, net.NS(ns))
	}
	return
}
func (nss *NSs) Set(newNssRaw []string) {
	*nss = NSs{}
	for _, newNsRaw := range newNssRaw {
		*nss = append(*nss, NS{Host: newNsRaw})
	}
}
