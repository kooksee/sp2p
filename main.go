package sp2p

import (
	"net"
	"time"
	"strings"
	"io"
	"github.com/satori/go.uuid"
)

func NewSP2p() *SP2p {
	logger := GetLog()

	taddr, err := net.ResolveTCPAddr("tcp", cfg.Adds[0])
	if err != nil {
		panic(err.Error())
	}

	p2p := &SP2p{
		txRC:      make(chan *KMsg, 10000),
		txWC:      make(chan *KMsg, 10000),
		localAddr: taddr,
		laddr:     cfg.Adds[0],
	}

	uad, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		panic(err.Error())
	}
	cnn, err := net.ListenUDP("udp", uad)
	if err != nil {
		panic(err.Error())
	}
	p2p.rconn = cnn

	if conn, err := net.DialTCP("udp", nil, p2p.localAddr); err != nil {
		panic(Errs(Fmt("udp %s listen error", p2p.localAddr), err.Error()))
	} else {
		p2p.conn = conn
	}

	nodeId := MustHexID(If(cfg.NodeId == "", GenNodeID(), cfg.NodeId).(string))

	logger.Debug("node id", "id", nodeId)
	logger.Debug("create table", "table")

	p2p.tab = newTable(nodeId, p2p.localAddr)

	go p2p.accept()
	go p2p.loop()
	go p2p.genUUID()

	return p2p
}

type SP2p struct {
	tab       *Table
	txRC      chan *KMsg
	txWC      chan *KMsg
	conn      *net.TCPConn
	rconn     *net.UDPConn
	localAddr *net.TCPAddr
	laddr     string
}

// 生成uuid的队列
func (s *SP2p) genUUID() {
	for {
		uid, err := uuid.NewV4()
		if err == nil {
			cfg.uuidC <- uid.String()
		}
	}
}

func (s *SP2p) GetAddr() string {
	return s.laddr
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
	if msg.Addr == "" {
		msg.Addr = s.GetAddr()
	}
	if msg.ID == "" {
		msg.ID = <-cfg.uuidC
	}
	if msg.Version == "" {
		msg.Version = cfg.Version
	}
	if msg.TAddr == "" {
		GetLog().Error("target udp addr is nonexistent")
		return
	}

	addr, err := net.ResolveUDPAddr("udp", msg.TAddr)
	if err != nil {
		GetLog().Error("ResolveUDPAddr error", "err", err)
		return
	}

	if _, err := s.rconn.WriteToUDP(msg.Dumps(), addr); err != nil {
		GetLog().Error("WriteToUDP error", "err", err)
		return
	}
}

func (s *SP2p) pingNode(taddr string) {
	s.Write(&KMsg{TAddr: taddr, ID: s.tab.selfNode.ID.ToHex(), Data: &PingReq{}})
}

func (s *SP2p) pingN() {
	for _, n := range s.tab.FindRandomNodes(cfg.PingNodeNum) {
		s.pingNode(n.AddrString())
	}
}

func (s *SP2p) findNode(taddr string, n int) {
	s.Write(&KMsg{TAddr: taddr, Data: &FindNodeReq{N: n}, ID: s.tab.selfNode.ID.ToHex()})
}

func (s *SP2p) findN() {
	for _, b := range s.tab.buckets {
		if b == nil || b.size() == 0 {
			continue
		}
		s.findNode(b.Random().AddrString(), cfg.FindNodeNUm)
	}
}

func (s *SP2p) kvSetReq(req *KVSetReq) {
	s.writeTx(&KMsg{Data: req, TAddr: s.GetAddr()})
}

func (s *SP2p) kvGetReq(req *KVGetReq) {
	s.writeTx(&KMsg{Data: req, TAddr: s.GetAddr()})
}

func (s *SP2p) gkvSetReq(req *GKVSetReq) {
	s.writeTx(&KMsg{Data: req, TAddr: s.GetAddr()})
}

func (s *SP2p) gkvGetReq(req *GKVGetReq) {
	s.writeTx(&KMsg{Data: req, TAddr: s.GetAddr()})
}

// 获得本地存储的value
func (s *SP2p) getValue(k []byte) ([]byte, error) {
	return GetDb().KHash(kvPrefix).Get(k)
}

func (s *SP2p) accept() {
	kb := NewKBuffer()
	logger := GetLog()
	for {
		buf := make([]byte, cfg.MaxBufLen)
		n, err := s.conn.Read(buf)
		if err == nil {
			logger.Debug(string(buf))
			messages := kb.Next(buf[:n])
			if messages == nil {
				continue
			}

			for _, m := range messages {
				if m == nil || len(m) == 0 {
					continue
				}

				msg := &KMsg{}
				if err := msg.Decode(m); err != nil {
					logger.Error("kmsg decode error", "err", err.Error(), "method", "sp2p.accept")
					continue
				}

				s.txRC <- msg
			}
			continue
		}

		if strings.Contains(err.Error(), "timeout") {
			GetLog().Error("timeout", "err", err)
		} else if err == io.EOF {
			GetLog().Error("udp read eof ", "err", err)
		} else if err != nil {
			GetLog().Error("udp read error ", "err", err)
		}

		time.Sleep(time.Second * 2)
	}
}
