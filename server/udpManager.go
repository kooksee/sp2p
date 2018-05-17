package server

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"

	knet "github.com/kooksee/srelay/utils/net"

	kts "github.com/kooksee/srelay/types"
)

type UdpServerManager struct {
	MaxPort int
	MinPort int
	umap    map[int]*knet.UdpListener
}

func (u *UdpServerManager) CreateUdp(port int) error {
	if len(u.umap) > u.MaxPort-u.MinPort {
		return errors.New("端口数量操作设置的最大限制")
	}

	if port > u.MaxPort || port < u.MinPort {
		return errors.New("超出了端口范围")
	}

	if _, ok := u.umap[port]; ok {
		return errors.New("端口已经存在")
	}

	l, err := knet.ListenUDP(fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		return err
	}

	u.umap[port] = l
	go u.onHandleListen(u.umap[port])
	return nil
}

func (u *UdpServerManager) onHandleConn(conn knet.Conn) {
	r := bufio.NewReader(conn)
	for {
		message, err := r.ReadBytes(kts.Delim)
		if err != nil {
			logger.Error(err.Error())
			break
		}
		message = bytes.TrimSpace(message)

		// 解析请求数据
		msg := &kts.KMsg{}
		if err := json.Unmarshal(message, msg); err != nil {
			logger.Error(err.Error())
			continue
		}
		ksInstance.Send(msg)
	}
}
func (u *UdpServerManager) onHandleListen(l *knet.UdpListener) {
	for {
		conn, err := l.Accept()
		if err != nil {
			logger.Error(err.Error())
			break
		}
		go u.onHandleConn(conn)
	}
}
