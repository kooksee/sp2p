package tests

import (
	"testing"
	"github.com/kooksee/sp2p"
	"net"
	"net/url"
	"fmt"
)

func TestName(t *testing.T) {
	sp2p.GenNodeID()
	cfg:=sp2p.DefaultConfig()
	cfg.InitP2P()
	//p2p:=cfg.GetP2P()
}

func TestName1(t *testing.T) {
	//u, err := url.Parse("sp2p://2345532132@127.0.0.1:8989")
	//u, err := url.Parse("sp2p://2345532132@uuu.com:8989")
	u, err := url.Parse("sp2p://uuu.com:8989")
	if err != nil {
		panic(err.Error())
	}
	if u.Scheme != "sp2p" {
		panic("oo")
	}
	// Parse the node ID from the user portion.
	if u.User == nil {
		panic("pp")
	}

	fmt.Println(u.User.String())

	// Parse the IP address.
	host, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println(host)
	fmt.Println(port)
}
