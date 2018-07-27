package sp2p

import (
	"time"
	"github.com/inconshreveable/log15"
	"net"
	"github.com/kooksee/kdb"
	"os"
	"path/filepath"
	"github.com/patrickmn/go-cache"
)

var (
	cfg *kConfig
)

type kConfig struct {
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

	NodesBackupKey string

	BucketSize int

	MaxNodeSize int
	MinNodeSize int
	Version     string

	StoreAckNum int

	Host          string
	Port          int
	AdvertiseAddr *net.UDPAddr
	NodeId        string

	Seeds []string

	uuidC chan string
	db    kdb.IKDB
	l     log15.Logger
	cache *cache.Cache
	p2p   ISP2P
}

func (t *kConfig) InitLog(l ... log15.Logger) *kConfig {
	if len(l) != 0 {
		t.l = l[0].New("package", "sp2p")
	} else {
		t.l = log15.New("package", "sp2p")
		t.l.SetHandler(log15.LvlFilterHandler(log15.LvlDebug, log15.StreamHandler(os.Stdout, log15.TerminalFormat())))
	}
	return t
}

func (t *kConfig) InitP2P() {
	t.p2p = newSP2p()
}

func (t *kConfig) GetP2P() ISP2P {
	if t.p2p == nil {
		panic("please init p2p")
	}
	return t.p2p
}

func (t *kConfig) InitDb(db ... kdb.IKDB) *kConfig {
	if len(db) != 0 {
		t.db = db[0]
	} else {
		kcfg := kdb.DefaultConfig()
		kcfg.InitKdb(filepath.Join("kdata", "db"))
		t.db = kcfg.GetDb()
	}
	return t
}

func getLog() log15.Logger {
	if getCfg().l == nil {
		panic("please init sp2p log")
	}
	return getCfg().l
}

func getDb() kdb.IKDB {
	if getCfg().db == nil {
		getLog().Error("please init sp2p db")
		panic("")
	}
	return getCfg().db
}

func getCfg() *kConfig {
	if cfg == nil {
		panic("please init sp2p config")
	}
	return cfg
}

func DefaultConfig() *kConfig {
	cfg = &kConfig{
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
		HashBits:            len(Hash{}) * 8,
		PingNodeNum:         8,
		FindNodeNUm:         20,
		ConnReadTimeout:     5 * time.Second,
		ConnWriteTimeout:    5 * time.Second,

		Host:           "0.0.0.0",
		Port:           8080,
		NodesBackupKey: "nbk:",

		PingTick:     time.NewTicker(10 * time.Minute),
		FindNodeTick: time.NewTicker(1 * time.Hour),
		NtpTick:      time.NewTicker(10 * time.Minute),

		MaxNodeSize: 2000,
		MinNodeSize: 100,
		Version:     "1.0.0",

		AdvertiseAddr: nil,
		BucketSize:    16,
		StoreAckNum:   2,

		uuidC: make(chan string, 500),
		cache: cache.New(10*time.Minute, 30*time.Minute),
	}

	return cfg
}
