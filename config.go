package sp2p

import (
	"crypto/ecdsa"
	"time"

	"github.com/kooksee/common"
	"net"
	"github.com/dgraph-io/badger"
	"github.com/kooksee/log"
)

var (
	cfg *KConfig
)

type KConfig struct {
	// 接收数据的最大缓存区
	MaxBufLen int

	// ntp服务器检测超时次数
	NtpFailureThreshold int
	//在重复NTP警告之前需要经过的最短时间
	NtpWarningCooldown time.Duration
	// ntpPool is the NTP server to query for the current time
	NtpPool string
	// Number of measurements to do against the NTP server
	NtpChecks int

	// Allowed clock drift before warning user
	DriftThreshold time.Duration

	PingTick     *time.Ticker
	FindNodeTick *time.Ticker
	NtpTick      *time.Ticker

	// Kademlia concurrency factor
	Alpha int
	// 节点响应的数量
	NodeResponseNumber int
	// 节点广播的数量
	NodeBroadcastNumber int
	// 节点分区存储的数量
	NodePartitionNumber int

	PingNodeNum int
	FindNodeNUm int

	// 节点ID长度
	HashBits int

	ConnReadTimeout  time.Duration
	ConnWriteTimeout time.Duration

	Host string
	Port int

	NodesBackupKey string

	DELIMITER byte

	PriV *ecdsa.PrivateKey

	BucketSize int

	MaxNodeSize int
	MinNodeSize int
	Version     string

	AdvertiseAddr *net.UDPAddr

	StoreAckNum int

	uuidC chan string
	db    *badger.DB
	l     log.Logger
}

func (t *KConfig) InitLog(l log.Logger) {
	t.l = l.New("package", "sp2p")
}
func (t *KConfig) InitDb(db *badger.DB) {
	t.db = db
}
func GetLog() log.Logger {
	if GetCfg().l == nil {
		panic("please init log")
	}
	return GetCfg().l
}

func GetDb() *badger.DB {
	if GetCfg().db == nil {
		panic("please init db")
	}
	return GetCfg().db
}

func GetCfg() *KConfig {
	if cfg == nil {
		panic("please init kconfig")
	}
	return cfg
}

func InitKConfig() *KConfig {
	cfg = &KConfig{
		MaxBufLen:           1024 * 16,
		NtpFailureThreshold: 32,
		NtpWarningCooldown:  10 * time.Minute,
		NtpPool:             "pool.ntp.org",
		NtpChecks:           3,
		DriftThreshold:      10 * time.Second,
		Alpha:               3,
		NodeResponseNumber:  8,
		NodeBroadcastNumber: 16,
		NodePartitionNumber: 8,
		HashBits:            len(common.Hash{}) * 8,
		PingNodeNum:         8,
		FindNodeNUm:         20,
		ConnReadTimeout:     5 * time.Second,
		ConnWriteTimeout:    5 * time.Second,

		Host:           "0.0.0.0",
		Port:           8080,
		NodesBackupKey: "nbk:",
		DELIMITER:      '\n',

		PingTick:     time.NewTicker(10 * time.Minute),
		FindNodeTick: time.NewTicker(1 * time.Hour),
		NtpTick:      time.NewTicker(10 * time.Minute),

		MaxNodeSize: 2000,
		MinNodeSize: 100,
		Version:     "1.0.0",

		AdvertiseAddr: nil,
		BucketSize:    16,
		StoreAckNum:   2,

		uuidC: make(chan string, 1000),
	}
	return cfg
}
