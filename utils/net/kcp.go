// Copyright 2017 fatedier, fatedier@gmail.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package net

import (
	"crypto/sha1"
	"fmt"
	"net"

	log "github.com/sirupsen/logrus"

	"github.com/xtaci/kcp-go"

	"golang.org/x/crypto/pbkdf2"
)

func GetCrypt(crypt string, key string, salt string) (kcp.BlockCrypt, error) {

	pass := pbkdf2.Key([]byte(key), []byte(salt), 4096, 32, sha1.New)
	switch crypt {
	case "sm4":
		return kcp.NewSM4BlockCrypt(pass[:16])
	case "tea":
		return kcp.NewTEABlockCrypt(pass[:16])
	case "xor":
		return kcp.NewSimpleXORBlockCrypt(pass)
	case "none":
		return kcp.NewNoneBlockCrypt(pass)
	case "aes-128":
		return kcp.NewAESBlockCrypt(pass[:16])
	case "aes-192":
		return kcp.NewAESBlockCrypt(pass[:24])
	case "blowfish":
		return kcp.NewBlowfishBlockCrypt(pass)
	case "twofish":
		return kcp.NewTwofishBlockCrypt(pass)
	case "cast5":
		return kcp.NewCast5BlockCrypt(pass[:16])
	case "3des":
		return kcp.NewTripleDESBlockCrypt(pass[:24])
	case "xtea":
		return kcp.NewXTEABlockCrypt(pass[:16])
	case "salsa20":
		return kcp.NewSalsa20BlockCrypt(pass)
	default:
		crypt = "aes"
		return kcp.NewAESBlockCrypt(pass)
	}
}

type KcpListener struct {
	net.Addr
	listener  net.Listener
	accept    chan Conn
	closeFlag bool
	*log.Logger
}

func ListenKcp(bindAddr string, bindPort int, block kcp.BlockCrypt) (l *KcpListener, err error) {
	listener, err := kcp.ListenWithOptions(fmt.Sprintf("%s:%d", bindAddr, bindPort), block, 10, 3)
	if err != nil {
		return l, err
	}
	listener.SetReadBuffer(4194304)
	listener.SetWriteBuffer(4194304)

	l = &KcpListener{
		Addr:      listener.Addr(),
		listener:  listener,
		accept:    make(chan Conn),
		closeFlag: false,
		Logger:    log.StandardLogger(),
	}

	go func() {
		for {
			conn, err := listener.AcceptKCP()
			if err != nil {
				if l.closeFlag {
					close(l.accept)
					return
				}
				continue
			}
			conn.SetStreamMode(true)
			conn.SetWriteDelay(true)
			conn.SetNoDelay(1, 20, 2, 1)
			conn.SetMtu(1350)
			conn.SetWindowSize(1024, 1024)
			conn.SetACKNoDelay(false)

			l.accept <- WrapConn(conn)
		}
	}()
	return l, err
}

func (l *KcpListener) Accept() (Conn, error) {
	conn, ok := <-l.accept
	if !ok {
		return conn, fmt.Errorf("channel for kcp listener closed")
	}
	return conn, nil
}

func (l *KcpListener) Close() error {
	if !l.closeFlag {
		l.closeFlag = true
		l.listener.Close()
	}
	return nil
}
