package godhcpdconfig

import (
	"net"
	"regexp"
)

var IsAsciiString = regexp.MustCompile(`^[a-zA-Z0-9,\.?=\-_\\/<>;':"{}\[\]~!@#$%^&*()+*]+$`).MatchString

const (
	UNKNOWN     = ValueType(0)
	BYTEARRAY   = ValueType(1)
	ASCIISTRING = ValueType(2)
)

type Range struct {
	Start net.IP `json:",omitempty"`
	End   net.IP `json:",omitempty"`
}

type ValueType int

func (vt ValueType) ConfigString() string {
	switch vt {
	case BYTEARRAY:
		return "array of integer 8"
	case ASCIISTRING:
		return "text"
	}
	return ""
}
