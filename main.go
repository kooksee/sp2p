package sp2p

import (
	"net"
	"time"
	"strings"
	"io"
	"github.com/satori/go.uuid"
)

func newSP2p() ISP2P {
	logger := getLog()

	p2p := &sp2p{
		txRC:      make(chan *KMsg, 10000),
		txWC:      make(chan *KMsg, 10000),
		localAddr: cfg.localNode.udpAddr,
	}

	logger.Debug("ListenUDP", "addr", p2p.localAddr.String())

	logger.Debug("create table", "table")
	p2p.tab = newTable(cfg.localNode.ID, cfg.localNode.udpAddr)

	go p2p.accept()
	go p2p.loop()
	go p2p.genUUID()

	return p2p
}

type sp2p struct {
	ISP2P

	tab       *table
	txRC      chan *KMsg
	txWC      chan *KMsg
	localAddr *net.UDPAddr
	laddr     string
}

// 生成uuid的队列
func (s *sp2p) genUUID() {
	for {
		uid, err := uuid.NewV4()
		if err == nil {
			cfg.uuidC <- uid.String()
		}
	}
}

func (s *sp2p) getAddr() string {
	if s.laddr == "" {
		s.laddr = s.localAddr.String()
	}
	return s.laddr
}

func (s *sp2p) loop() {
	for {
		select {
		case <-cfg.FindNodeTick.C:
			go s.findRandom()
		case <-cfg.PingTick.C:
			go s.pingRandom()
		case <-cfg.NtpTick.C:
			go func() {
				if err := checkClockDrift(); err != nil {
					getLog().Error("checkClockDrift error", "err", err.Error())
				}
			}()
		case tx := <-s.txRC:
			go tx.Data.OnHandle(s, tx)
		case tx := <-s.txWC:
			go func() {
				if err := s.write(tx); err != nil {
					getLog().Error("sp2p loop write error", "err", err.Error())
				}
			}()
		}
	}
}

func (s *sp2p) writeTx(msg *KMsg) {
	s.txWC <- msg
}

func (s *sp2p) write(msg *KMsg) error {
	if msg.FN == "" {
		msg.FN = cfg.localNode.string()
	}
	if msg.ID == "" {
		msg.ID = <-cfg.uuidC
	}
	if msg.TN == "" {
		return errs("target node id is nonexistent")
	}

	n, err := NodeParse(msg.TN)
	if err != nil {
		return err
	}

	if _, err := getLocalConn().WriteToUDP(msg.Dumps(), n.udpAddr); err != nil {
		return errPipe("WriteToUDP error", err)
	}
	return nil
}

func (s *sp2p) pingRandom() {
	for _, n := range s.tab.findRandomNodes(cfg.PingNodeNum) {
		s.writeTx(&KMsg{TN: n.string(), Data: &pingReq{}})
	}
}

func (s *sp2p) findRandom() {
	for _, b := range s.tab.buckets {
		if b == nil || b.size() == 0 {
			continue
		}
		s.writeTx(&KMsg{TN: b.random().string(), Data: &findNodeReq{N: cfg.FindNodeNUm}})
	}
}

func (s *sp2p) accept() {
	kb := newKBuffer()
	logger := getLog()
	for {
		buf := make([]byte, cfg.MaxBufLen)
		n, err := getConn().Read(buf)
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

			// 检查该ID是否已经存在过,防止数据重复发送
			cache := getCfg().cache
			if _, b := cache.Get(msg.ID); b {
				continue
			} else {
				cache.SetDefault(msg.ID, true)
				s.txRC <- msg
			}
		}
	}
}
