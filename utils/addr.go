package utils

import (
	"fmt"
	"net/url"
	"strings"
)

type NetAddr struct {
	Addr    string
	Network string
}

func (a *NetAddr) String() string {
	return fmt.Sprintf("%v://%v", a.Network, a.Addr)
}

func ParseAddr(a string) (*NetAddr, error) {
	u, err := url.Parse(a)
	if err != nil {
		return nil, fmt.Errorf("failed to parse '%v':%v", a, err)
	}
	switch u.Scheme {
	case "tcp":
		return &NetAddr{Addr: u.Host, Network: u.Scheme}, nil
	case "unix":
		return &NetAddr{Addr: u.Path, Network: u.Scheme}, nil
	default:
		return nil, fmt.Errorf("unsupported scheme '%v': '%v'", a, u.Scheme)
	}
}

func NewNetAddrVal(defaultVal NetAddr, val *NetAddr) *NetAddrVal {
	*val = defaultVal
	return (*NetAddrVal)(val)
}

// NetAddrVal can be used with flag package
type NetAddrVal NetAddr

func (a *NetAddrVal) Set(s string) error {
	v, err := ParseAddr(s)
	if err != nil {
		return err
	}
	a.Addr = v.Addr
	a.Network = v.Network
	return nil
}

func (a *NetAddrVal) String() string {
	return ((*NetAddr)(a)).String()
}

func (a *NetAddrVal) Get() interface{} {
	return NetAddr(*a)
}

func NewNetAddrList(addrs *[]NetAddr) *NetAddrList {
	return &NetAddrList{addrs: addrs}
}

type NetAddrList struct {
	addrs *[]NetAddr
}

func (nl *NetAddrList) Set(s string) error {
	v, err := ParseAddr(s)
	if err != nil {
		return err
	}
	*nl.addrs = append(*nl.addrs, *v)
	return nil
}

func (nl *NetAddrList) String() string {
	var ns []string
	for _, n := range *nl.addrs {
		ns = append(ns, n.String())
	}
	return strings.Join(ns, " ")
}
