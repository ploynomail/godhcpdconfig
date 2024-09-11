package godhcpdconfig

import (
	"os"

	leases "github.com/ploynomail/godhcpdconfig/lease"
)

type Lease struct {
	LeaseFile string
}

type LeaseInfo struct {
	Hostname string `json:"hostname"`
	IP       string `json:"ip"`
	MAC      string `json:"mac"`
}

func NewLease(leaseFile string) *Lease {
	if leaseFile == "" {
		leaseFile = "/var/lib/dhcp/dhcpd.leases"
	}
	return &Lease{LeaseFile: leaseFile}
}

func (l *Lease) Parse() (leaseInfo []*LeaseInfo, err error) {
	f, err := os.Open(l.LeaseFile)
	if err != nil {
		return nil, err
	}
	leases := leases.Parse(f)
	var leaseInfoList []*LeaseInfo
	for _, lease := range leases {
		if lease.BindingState == "active" {
			leaseInfo := &LeaseInfo{
				Hostname: lease.ClientHostname,
				IP:       lease.IP.String(),
				MAC:      lease.Hardware.MAC,
			}
			leaseInfoList = append(leaseInfoList, leaseInfo)
		}
	}
	return leaseInfoList, nil
}
