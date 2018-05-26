package sp2p

import (
	"net"
	"time"
	"strings"
	"io"
	"bytes"
	"github.com/dgraph-io/badger"
	"encoding/hex"
	"github.com/satori/go.uuid"
)

func NewSP2p(seeds []string) *SP2p {
	p2p := &SP2p{
		txRC:      make(chan *KMsg, 2000),
		txWC:      make(chan *KMsg, 2000),
		localAddr: &net.UDPAddr{Port: cfg.Port, IP: net.ParseIP(cfg.Host)},
	}

	if cfg.AdvertiseAddr == nil {
		cfg.AdvertiseAddr = &net.UDPAddr{Port: cfg.Port, IP: net.ParseIP("127.0.0.1")}
		GetLog().Error("请设置ExportAddr")
	}

	conn, err := net.ListenUDP("udp", p2p.localAddr)
	if err != nil {
		panic(err.Error())
	}
	p2p.conn = conn
	p2p.tab = newTable(PubkeyID(&cfg.PriV.PublicKey), cfg.AdvertiseAddr)

	go p2p.accept()
	go p2p.loop()
	go p2p.genUUID()

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

// 生成uuid的队列
func (s *SP2p) genUUID() {
	for {
		uid, err := uuid.NewV4()
		if err == nil {
			cfg.uuidC <- hex.EncodeToString(uid.Bytes())
		}
	}
}

func (s *SP2p) loadSeeds(seeds []string) error {
	txn := GetDb().NewTransaction(true)
	defer txn.Discard()

	k := []byte(cfg.NodesBackupKey)
	iter := txn.NewIterator(badger.DefaultIteratorOptions)
	for iter.Seek(k); ; iter.Next() {
		if !iter.ValidForPrefix(k) {
			break
		}

		val, err := iter.Item().Value()
		if err != nil {
			GetLog().Error("loadSeeds", "err", err)
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
		go s.pingNode(n.Addr().String())
	}

	// 节点启动的时候如果发现节点数量少,就去请求其他节点
	if s.tab.Size() < cfg.MinNodeSize {
		// 节点太少的情况下，就去所有的节点去请求数据
		for _, b := range s.tab.buckets {
			b.peers.Each(func(index int, value interface{}) {
				go s.findNode(value.(*Node).Addr().String(), 8)
			})
		}
	} else if s.tab.Size() < cfg.MaxNodeSize {
		// 每一个域选取一个节点
		for _, b := range s.tab.buckets {
			go s.findNode(b.Random().Addr().String(), 8)
		}
	}

	return txn.Commit(nil)
}

func (s *SP2p) loop() {
	for {
		select {
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

func (s *SP2p) writeTx(msg *KMsg) {
	s.txWC <- msg
}

func (s *SP2p) write(msg *KMsg) {
	if msg.FAddr == "" {
		msg.FAddr = s.localAddr.String()
	}
	if msg.ID == "" {
		msg.ID = <-cfg.uuidC
	}
	if msg.Version == "" {
		msg.Version = cfg.Version
	}
	if msg.TAddr == "" {
		GetLog().Error("target udp addr does not exist")
		return
	}

	addr, err := net.ResolveUDPAddr("udp", msg.TAddr)
	if err != nil {
		GetLog().Error("ResolveUDPAddr error", "err", err)
		return
	}
	if _, err := s.conn.WriteToUDP(msg.Dumps(), addr); err != nil {
		GetLog().Error("WriteToUDP error", "err", err)
		return
	}
}

func (s *SP2p) pingNode(taddr string) {
	s.Write(&KMsg{TAddr: taddr, FID: s.tab.selfNode.ID.String(), Data: &PingReq{}})
}

func (s *SP2p) pingN() {
	for _, n := range s.tab.FindRandomNodes(cfg.PingNodeNum) {
		s.pingNode(n.Addr().String())
	}
}

func (s *SP2p) findNode(taddr string, n int) {
	s.Write(&KMsg{TAddr: taddr, Data: &FindNodeReq{N: n}, FID: s.tab.selfNode.ID.String()})
}

func (s *SP2p) findN() {
	for _, b := range s.tab.buckets {
		if b == nil || b.size() == 0 {
			continue
		}
		s.findNode(b.Random().Addr().String(), cfg.FindNodeNUm)
	}
}

func (s *SP2p) kvSetReq(req *KVSetReq) {
	s.writeTx(&KMsg{Data: req, TAddr: s.localAddr.String()})
}

func (s *SP2p) kvGetReq(req *KVGetReq) {
	s.writeTx(&KMsg{Data: req, TAddr: s.localAddr.String()})
}

func (s *SP2p) gkvSetReq(req *GKVSetReq) {
	s.writeTx(&KMsg{Data: req, TAddr: s.localAddr.String()})
}

func (s *SP2p) gkvGetReq(req *GKVGetReq) {
	s.writeTx(&KMsg{Data: req, TAddr: s.localAddr.String()})
}

// 获得本地存储的value
func (s *SP2p) getValue(k []byte) (value []byte, err error) {
	return value, GetDb().View(func(txn *badger.Txn) error {
		item, err := txn.Get(k)
		if err != nil {
			return err
		}
		v, err := item.Value()
		if err != nil {
			return err
		}
		value = v
		return nil
	})
}

func (s *SP2p) accept() {
	kb := NewKBuffer([]byte{'\n'})
	for {
		buf := make([]byte, cfg.MaxBufLen)
		n, addr, err := s.conn.ReadFromUDP(buf)
		if err == nil {
			GetLog().Debug("udp message", "addr", addr.String())
			GetLog().Debug(string(buf))
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
					GetLog().Error("tx msg decode error", "err", err, "method", "accept")
					continue
				}
				s.txRC <- msg
			}
			continue
		}
		if strings.Contains(err.Error(), "timeout") {
			GetLog().Error("timeout", "err", err)
			time.Sleep(time.Second * 2)
		} else if err == io.EOF {
			GetLog().Error("udp read eof ", "err", err)
			break
		} else if err != nil {
			GetLog().Error("udp read error ", "err", err)
			time.Sleep(time.Second * 2)
		}
	}
}
