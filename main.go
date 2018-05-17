package sp2p

import (
	"errors"
	"net"
	"time"
	"bufio"
	"bytes"

	"github.com/dgraph-io/badger"
)

func NewSP2p(seeds []string) *SP2p {
	p2p := &SP2p{
		txRC:      make(chan *KMsg, 2000),
		txWC:      make(chan *KMsg, 2000),
		localAddr: &net.UDPAddr{Port: cfg.Port, IP: net.ParseIP(cfg.Host)},
	}

	if cfg.ExportAddr == nil {
		panic("请设置ExportAddr")
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
	IP2p
	tab       *Table
	txRC      chan *KMsg
	txWC      chan *KMsg
	conn      *net.UDPConn
	localAddr *net.UDPAddr
}

func (s *SP2p) loadSeeds(seeds []string) error {
	return cfg.Db.Update(func(txn *badger.Txn) error {
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
			// 每一个域选取一个节点
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

		return nil
	})
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
		logger.Error(err.Error())
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
		msg.FID = s.tab.selfNode.ID.String()
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
	if _, err := s.conn.Write(msg.Dumps()); err != nil {
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
	s.conn.SetReadDeadline(time.Now().Add(cfg.ConnReadTimeout))
	read := bufio.NewReader(s.conn)
	for {
		message, err := read.ReadBytes(cfg.DELIMITER)
		if err != nil {
			logger.Info("udp read error ", "err", err.Error())
			break
		}
		message = bytes.TrimSpace(message)
		logger.Debug("udp message", "msg", string(message))

		msg := &KMsg{}
		if err := msg.Decode(message); err != nil {
			logger.Error("kmsg decode error", "err", err)
			continue
		}
		s.txRC <- msg
	}
}
