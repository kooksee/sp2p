package sp2p

import (
	"errors"
	"net"
	"time"
	"github.com/dgraph-io/badger"
	"strings"
	"io"
	"bytes"
)

func NewSP2p(seeds []string) *SP2p {
	p2p := &SP2p{
		txRC:      make(chan *KMsg, 2000),
		txWC:      make(chan *KMsg, 2000),
		localAddr: &net.UDPAddr{Port: cfg.Port, IP: net.ParseIP("127.0.0.1")},
	}

	if cfg.ExportAddr == nil {
		cfg.ExportAddr = &net.UDPAddr{Port: cfg.Port, IP: net.ParseIP("127.0.0.1")}
		logger.Error("请设置ExportAddr")
	}

	conn, err := net.ListenUDP("udp", p2p.localAddr)
	if err != nil {
		panic(err.Error())
	}
	p2p.conn = conn

	p2p.tab = newTable(PubkeyID(&cfg.PriV.PublicKey), cfg.ExportAddr)

	go p2p.accept()
	go p2p.loop()

	if err := p2p.loadSeeds(seeds); err != nil {
		panic(err.Error())
	}
	return p2p
}

type SP2p struct {
	tab       *Table
	txRC      chan *KMsg
	txWC      chan *KMsg
	conn      *net.UDPConn
	localAddr *net.UDPAddr
}

func (s *SP2p) loadSeeds(seeds []string) error {
	txn := cfg.Db.NewTransaction(true)
	defer txn.Discard()

	k := []byte(cfg.NodesBackupKey)
	iter := txn.NewIterator(badger.DefaultIteratorOptions)
	for iter.Seek(k); ; iter.Next() {
		if !iter.ValidForPrefix(k) {
			break
		}

		val, err := iter.Item().Value()
		if err != nil {
			logger.Error(err.Error())
			continue
		}

		seeds = append(seeds, string(val))
	}

	for _, rn := range seeds {
		n := MustParseNode(rn)
		n.updateAt = time.Now()
		s.tab.AddNode(n)
		go s.pingNode(n.addr().String())
	}

	// 节点启动的时候如果发现节点数量少,就去请求其他节点
	if s.tab.Size() < cfg.MinNodeSize {
		// 节点太少的情况下，就去所有的节点去请求数据
		for _, b := range s.tab.buckets {
			b.peers.Each(func(index int, value interface{}) {
				logger.Error(value.(*Node).addr().String())
				go s.findNode(value.(*Node).addr().String(), 8)
			})
		}
	} else if s.tab.Size() < cfg.MaxNodeSize {
		// 每一个域选取一个节点
		for _, b := range s.tab.buckets {
			go s.findNode(b.Random().addr().String(), 8)
		}
	}

	return txn.Commit(nil)
}

func (s *SP2p) dumpSeeds() {
	if err := cfg.Db.Update(func(txn *badger.Txn) error {
		for _, n := range s.tab.GetAllNodes() {
			k := append([]byte(cfg.NodesBackupKey), n.ID.Bytes()...)
			if err := txn.Set(k, []byte(n.String())); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		logger.Error("dumpSeeds error", "err", err)
	}
}

func (s *SP2p) loop() {
	for {
		select {
		case <-cfg.NodeBackupTick.C:
			go s.dumpSeeds()
		case <-cfg.FindNodeTick.C:
			for _, b := range s.tab.buckets {
				if b == nil {
					continue
				}
				go s.findNode(b.Random().addr().String(), 8)
			}
		case <-cfg.PingTick.C:
			for _, n := range s.tab.FindRandomNodes(20) {
				go s.pingNode(n.addr().String())
			}
		case <-cfg.NtpTick.C:
			go checkClockDrift()
		case tx := <-s.txRC:
			if hm.Contain(tx.Event) {
				go hm.GetHandler(tx.Event)(s, tx)
			}
		case tx := <-s.txWC:
			if err := s.write(tx); err != nil {
				logger.Error("write tx error", "err", err)
			}
		}
	}
}

func (s *SP2p) Write(msg *KMsg) {
	s.txWC <- msg
}

func (s *SP2p) write(msg *KMsg) error {
	if msg.FAddr == "" {
		msg.FAddr = s.localAddr.String()
	}
	if msg.FID == "" {
		msg.FID = s.tab.selfNode.sha.String()
	}
	if msg.ID == "" {
		msg.ID = string(UUID())
	}
	if msg.Version == "" {
		msg.Version = cfg.Version
	}
	if msg.TAddr == "" {
		return errors.New("目标地址不存在")
	}

	addr, _ := net.ResolveUDPAddr("udp", msg.TAddr)
	if _, err := s.conn.WriteToUDP(msg.Dumps(), addr); err != nil {
		return err
	}
	return nil
}

func (s *SP2p) GetTable() *Table {
	return s.tab
}

func (s *SP2p) pingNode(taddr string) {
	s.Write(&KMsg{
		Event: "ping",
		TAddr: taddr,
	})
}

func (s *SP2p) findNode(taddr string, n int) {
	s.Write(&KMsg{
		Event: "findNode",
		TAddr: taddr,
		Data:  FindNodeReq{NID: s.tab.selfNode.ID.String(), N: n},
	})
}

func (s *SP2p) accept() {
	kb := NewKBuffer([]byte{'\n'})
	for {
		buf := make([]byte, 1024*16)
		n, addr, err := s.conn.ReadFromUDP(buf)
		if err == nil {
			logger.Debug("udp message", "msg", string(buf), "addr", addr.String())
			messages := kb.Next(buf[:n])
			if messages == nil {
				continue
			}

			for _, m := range messages {
				if m == nil || bytes.Equal(m, []byte{}) {
					continue
				}

				msg := &KMsg{}
				if err := msg.Decode(m); err != nil {
					logger.Error("tx msg decode error", "err", err, "method", "accept")
					continue
				}
				s.txRC <- msg
			}
			continue
		}
		if strings.Contains(err.Error(), "timeout") {
			logger.Debug(err.Error())

			for _, n := range s.tab.FindRandomNodes(20) {
				logger.Error(n.String())
				go s.pingNode(n.addr().String())
			}

			time.Sleep(time.Second * 2)
			continue
		} else if err == io.EOF {
			break
		} else if err != nil {
			logger.Error("udp read error ", "err", err)
		}
	}
}
