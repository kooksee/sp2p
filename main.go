package sp2p

import (
	"net"
	"time"
	"strings"
	"io"
	"github.com/satori/go.uuid"
)

func NewSP2p() *ISP2P {
	logger := getLog()

	p2p := &SP2p{
		txRC:      make(chan *KMsg, 10000),
		txWC:      make(chan *KMsg, 10000),
		localAddr: &net.UDPAddr{Port: cfg.Port, IP: net.ParseIP(cfg.Host)},
	}

	if cfg.AdvertiseAddr == nil {
		logger.Error("没有设置AdvertiseAddr")
		cfg.AdvertiseAddr = &net.UDPAddr{Port: cfg.Port, IP: net.ParseIP("127.0.0.1")}
		logger.Warn("默认AdvertiseAddr", "addr", cfg.AdvertiseAddr.String())
	}

	logger.Debug("ListenUDP", "addr", p2p.localAddr.String())

	if conn, err := net.ListenUDP("udp", p2p.localAddr); err != nil {
		panic(errs(f("udp %s listen error", p2p.localAddr), err.Error()))
	} else {
		p2p.conn = conn
	}

	nodeId := MustHexID(If(cfg.NodeId == "", GenNodeID(), cfg.NodeId).(string))
	logger.Debug("node id", "id", nodeId)

	logger.Debug("create table", "table")
	p2p.tab = newTable(nodeId, cfg.AdvertiseAddr)

	go p2p.accept()
	go p2p.loop()
	go p2p.genUUID()

	return p2p
}

type SP2p struct {
	ISP2P

	tab       *table
	txRC      chan *KMsg
	txWC      chan *KMsg
	conn      *net.UDPConn
	localAddr *net.UDPAddr
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
	if s.laddr == "" {
		s.laddr = s.localAddr.String()
	}
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
	if msg.FAddr == "" {
		msg.FAddr = s.GetAddr()
	}
	if msg.ID == "" {
		msg.ID = <-cfg.uuidC
	}
	if msg.Version == "" {
		msg.Version = cfg.Version
	}
	if msg.TAddr == "" {
		getLog().Error("target udp addr is nonexistent")
		return
	}

	addr, err := net.ResolveUDPAddr("udp", msg.TAddr)
	if err != nil {
		getLog().Error("ResolveUDPAddr error", "err", err)
		return
	}

	if _, err := s.conn.WriteToUDP(msg.Dumps(), addr); err != nil {
		getLog().Error("WriteToUDP error", "err", err)
		return
	}
}

func (s *SP2p) pingN() {
	for _, n := range s.tab.findRandomNodes(cfg.PingNodeNum) {
		s.writeTx(&KMsg{TAddr: n.addrString(), FID: s.tab.selfNode.ID.Hex(), Data: &pingReq{}})
	}
}

func (s *SP2p) findN() {
	for _, b := range s.tab.buckets {
		if b == nil || b.size() == 0 {
			continue
		}
		s.writeTx(&KMsg{TAddr: b.random().addrString(), Data: &findNodeReq{N: cfg.FindNodeNUm}, FID: s.tab.selfNode.ID.Hex()})
	}
}

func (s *SP2p) accept() {
	kb := newKBuffer()
	logger := getLog()
	for {
		buf := make([]byte, cfg.MaxBufLen)
		n, addr, err := s.conn.ReadFromUDP(buf)
		if err != nil {
			if strings.Contains(err.Error(), "timeout") {
				logger.Error("timeout", "err", err)
			} else if err == io.EOF {
				logger.Error("udp read eof ", "err", err)
				break
			} else if err != nil {
				logger.Error("udp read error ", "err", err)
			}
			time.Sleep(time.Second * 2)
			continue
		}
		logger.Debug("udp message", "addr", addr.String())
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
	}
}
