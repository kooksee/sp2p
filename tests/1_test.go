package tests

import (
	"testing"
	"github.com/kooksee/sp2p"
)

func TestName(t *testing.T) {
	sp2p.GenNodeID()
	cfg:=sp2p.DefaultConfig()
	cfg.InitP2P()
	p2p:=cfg.GetP2P()
	p2p.Broadcast()
}
