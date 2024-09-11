package godhcpdconfig

import (
	"testing"
)

func Test_Lease(t *testing.T) {
	lease := NewLease("")
	info, err := lease.Parse()
	if err != nil {
		t.Errorf("Error parsing lease file:%s", err)
	}
	if len(info) == 0 {
		t.Errorf("Expecting at least one lease")
	}
	for _, leaseInfo := range info {
		t.Logf("Hostname:%s, IP:%s, MAC:%s", leaseInfo.Hostname, leaseInfo.IP, leaseInfo.MAC)
	}
}
