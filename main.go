package sp2p

import (
	"net"
	"time"
	"strings"
	"io"
	"bytes"
	"github.com/dgraph-io/badger"
	"encoding/hex"
)

func NewSP2p(seeds []string) *SP2p {
	p2p := &SP2p{
		txRC:      make(chan *KMsg, 2000),
		txWC:      make(chan *KMsg, 2000),
		localAddr: &net.UDPAddr{Port: cfg.Port, IP: net.ParseIP(cfg.Host)},
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
			logger.Error("loadSeeds", "err", err)
			continue
		}

		seeds = append(seeds, string(val))
	}

	for _, rn := range seeds {
		if rn == s.tab.selfNode.String() {
			continue
		}
		n := MustParseNode(rn)
		s.tab.AddNode(n)
		go s.pingNode(n.addr().String())
	}

	// 节点启动的时候如果发现节点数量少,就去请求其他节点
	if s.tab.Size() < cfg.MinNodeSize {
		// 节点太少的情况下，就去所有的节点去请求数据
		for _, b := range s.tab.buckets {
			b.peers.Each(func(index int, value interface{}) {
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
	cfg.Db.Update(func(txn *badger.Txn) error {
		for _, n := range s.tab.GetAllNodes() {
			if err := txn.Set(NodesBackupKey(n.ID.Bytes()), []byte(n.String())); err != nil {
				logger.Error("dumpSeeds", "err", err)
				continue
			}
		}
		return nil
	})
}

func (s *SP2p) loop() {
	for {
		select {
		case <-cfg.NodeBackupTick.C:
			go s.dumpSeeds()
		case <-cfg.FindNodeTick.C:
			go s.findN()
		case <-cfg.PingTick.C:
			go s.pingN()
		case <-cfg.NtpTick.C:
			go checkClockDrift()
		case tx := <-s.txRC:
			go tx.Data.OnHandle(s, tx)
		case tx := <-s.txWC:
			go s.write(tx)
		}
	}
}

func (s *SP2p) Write(msg *KMsg) {
	s.txWC <- msg
}

func (s *SP2p) write(msg *KMsg) {
	if msg.FAddr == "" {
		msg.FAddr = s.localAddr.String()
	}
	if msg.ID == "" {
		msg.ID = hex.EncodeToString(UUID())
	}
	if msg.Version == "" {
		msg.Version = cfg.Version
	}
	if msg.TAddr == "" {
		logger.Error("目标地址不存在")
		return
	}

	addr, err := net.ResolveUDPAddr("udp", msg.TAddr)
	if err != nil {
		logger.Error("ResolveUDPAddr error", "err", err)
		return
	}
	if _, err := s.conn.WriteToUDP(msg.Dumps(), addr); err != nil {
		logger.Error("WriteToUDP error", "err", err)
		return
	}
}

func (s *SP2p) GetTable() *Table {
	return s.tab
}

func (s *SP2p) pingNode(taddr string) {
	s.Write(&KMsg{
		TAddr: taddr,
		FID:   s.tab.selfNode.ID.String(),
		Data:  &PingReq{},
	})
}
func (s *SP2p) pingN() {
	for _, n := range s.tab.FindRandomNodes(20) {
		s.pingNode(n.addr().String())
	}
}

func (s *SP2p) findNode(taddr string, n int) {
	s.Write(&KMsg{
		TAddr: taddr,
		Data:  &FindNodeReq{N: n},
		FID:   s.tab.selfNode.ID.String(),
	})
}

func (s *SP2p) findN() {
	for _, b := range s.tab.buckets {
		if b == nil || b.size() == 0 {
			continue
		}
		s.findNode(b.Random().addr().String(), 8)
	}
}

func (s *SP2p) accept() {
	kb := NewKBuffer([]byte{'\n'})
	for {
		buf := make([]byte, cfg.MaxBufLen)
		n, addr, err := s.conn.ReadFromUDP(buf)
		if err == nil {
			logger.Debug("udp message", "addr", addr.String())
			logger.Debug(string(buf))
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
			logger.Error("timeout", "err", err)
			time.Sleep(time.Second * 2)
		} else if err == io.EOF {
			logger.Error("udp read eof ", "err", err)
			break
		} else if err != nil {
			logger.Error("udp read error ", "err", err)
		}
	}
}
