package sp2p

import (
	"crypto/ecdsa"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/kooksee/common"
	"net"
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

	// 节点ID长度
	HashBits int

	ConnReadTimeout  time.Duration
	ConnWriteTimeout time.Duration

	Host string
	Port int

	Db *badger.DB

	NodesBackupKey string

	DELIMITER byte

	PriV *ecdsa.PrivateKey

	BucketSize int

	MaxNodeSize int
	MinNodeSize int
	Version     string

	ExportAddr *net.UDPAddr
	LogLevel   string

	StoreAckNum int

	uuidC chan string
}

func DefaultKConfig() *KConfig {
	return &KConfig{
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

		ConnReadTimeout:  5 * time.Second,
		ConnWriteTimeout: 5 * time.Second,

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

		ExportAddr:  nil,
		LogLevel:    "info",
		BucketSize:  16,
		StoreAckNum: 2,

		uuidC: make(chan string, 1000),
	}
}
