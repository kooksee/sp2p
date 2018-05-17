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
	// Minimum amount of time to pass before repeating NTP warning
	//在重复NTP警告之前需要经过的最短时间
	NtpWarningCooldown time.Duration
	// ntpPool is the NTP server to query for the current time
	NtpPool string
	// Number of measurements to do against the NTP server
	NtpChecks int

	// Kademlia bucket size
	BucketSize int

	// Allowed clock drift before warning user
	DriftThreshold time.Duration

	// 节点列表备份时间
	NodeBackupTick *time.Ticker
	PingTick       *time.Ticker
	FindNodeTick   *time.Ticker
	PingKcpTick    *time.Ticker
	NtpTick        *time.Ticker

	// 节点长度
	NodeIDBits int

	// Kademlia concurrency factor
	Alpha int
	// 节点响应的数量
	NodeResponseNumber int
	// 节点广播的数量
	NodeBroadcastNumber int
	// 节点分区的数量
	NodePartitionNumber int

	// 节点ID长度
	HashBits int
	// K桶的数量
	NBuckets int

	ConnReadTimeout  time.Duration
	ConnWriteTimeout time.Duration

	Crypt string
	Key   string
	Salt  string
	Host  string
	Port  int

	Db *badger.DB

	NodesBackupKey string

	DELIMITER byte

	PriV *ecdsa.PrivateKey

	MaxNodeSize int
	MinNodeSize int
	Version     string

	ExportAddr *net.UDPAddr
}

func DefaultKConfig() *KConfig {
	return &KConfig{
		MaxBufLen:           1024 * 16,
		NtpFailureThreshold: 32,
		NtpWarningCooldown:  10 * time.Minute,
		NtpPool:             "pool.ntp.org",
		NtpChecks:           3,
		BucketSize:          16,
		DriftThreshold:      10 * time.Second,
		NodeIDBits:          512,
		Alpha:               3,
		NodeResponseNumber:  8,
		NodeBroadcastNumber: 16,
		NodePartitionNumber: 8,
		HashBits:            len(common.Hash{}) * 8,
		NBuckets:            len(common.Hash{})*8 + 1,

		ConnReadTimeout:  5 * time.Second,
		ConnWriteTimeout: 5 * time.Second,

		Crypt:          "aes-128",
		Key:            "hello",
		Salt:           "hello",
		Host:           "0.0.0.0",
		Port:           8080,
		NodesBackupKey: "nbk:",
		DELIMITER:      '\n',

		NodeBackupTick: time.NewTicker(10 * time.Minute),
		PingTick:       time.NewTicker(10 * time.Minute),
		FindNodeTick:   time.NewTicker(1 * time.Hour),
		PingKcpTick:    time.NewTicker(2 * time.Second),
		NtpTick:        time.NewTicker(10 * time.Minute),

		MaxNodeSize: 2000,
		MinNodeSize: 100,
		Version:     "1.0.0",
	}
}
