package server

import (
	"bufio"
	"bytes"
	"sync"
	"time"

	"github.com/kooksee/srelay/types"
	knet "github.com/kooksee/srelay/utils/net"
	"github.com/patrickmn/go-cache"
)

var ksInstance *KcpServer
var ksOnce sync.Once

type KcpServer struct {
	clients   *cache.Cache
	l         *knet.KcpListener
	usManager *UdpServerManager
	writeChan chan *types.KMsg
}

func GetKcpServer() *KcpServer {
	ksOnce.Do(func() {
		ksInstance = &KcpServer{
			clients:   cfg.Cache,
			usManager: &UdpServerManager{},
			writeChan: make(chan *types.KMsg, 3000),
		}
		go ksInstance.handleWriteLoop()
	})
	return ksInstance
}

func (ks *KcpServer) handleWriteLoop() {
	for {
		tx := <-ks.writeChan
		// 检查缓存中是否存在,存在跳过
		id, _ := cfg.Cache.Get(tx.ID)
		if id != nil {
			continue
		}

		cObj, b := ks.clients.Get(tx.TAddr)
		if !b {
			logger.Error("sid不存在")
			continue
		}
		conn, ok := cObj.(knet.Conn)
		if !ok {
			logger.Error("sid不存在")
			continue
		}
		conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		if _, err := conn.Write(tx.Dumps()); err != nil {
			logger.Error(err.Error())
			continue
		}

		// 执行成功之后先放入缓存
		cfg.Cache.SetDefault(tx.ID, true)
	}
}

func (ks *KcpServer) Send(tx *types.KMsg) {
	ks.writeChan <- tx
}

func (ks *KcpServer) Listen() error {
	block, err := knet.GetCrypt(cfg.Crypt, cfg.Key, cfg.Salt)
	if err != nil {
		panic(err.Error())
	}
	ks.l, err = knet.ListenKcp(cfg.Host, cfg.KcpPort, block)
	go ks.accept()
	return err
}

func (ks *KcpServer) onPing(tx *types.KMsg, conn knet.Conn) {
	if v, _ := cfg.Cache.Get(tx.Data.(string)); v != nil {
		ks.clients.SetDefault(v.(string), conn)
	}
}

func (ks *KcpServer) onReply(tx *types.KMsg) {
	ks.clients.SetDefault(tx.ID, tx)
}

func (ks *KcpServer) CreateUdp(port int) error {
	return ks.usManager.CreateUdp(port)
}

func (ks *KcpServer) onHandle(conn knet.Conn) {
	read := bufio.NewReader(conn)
	for {
		message, err := read.ReadBytes(types.Delim)
		if err != nil {
			logger.Info("kcp message error", "err", err)
			break
		}
		message = bytes.TrimSpace(message)
		logger.Debug("message data", "data", string(message))

		tx := &types.KMsg{}
		if err := json.Unmarshal(message, tx); err != nil {
			logger.Error("Unmarshal error", "err", err)
			continue
		}

		switch tx.Event {
		case types.PINGREQ:
			ks.onPing(tx, conn)
		default:
			logger.Warn("message event error", "err", "找不到该event", "event", tx.Event)
		}
	}
}

func (ks *KcpServer) accept() {
	for {
		c, err := ks.l.Accept()
		if err != nil {
			logger.Error("kcp conn error ", "err", err)
			continue
		}
		c.SetReadDeadline(time.Now().Add(connReadTimeout))
		go ks.onHandle(c)
	}
}
