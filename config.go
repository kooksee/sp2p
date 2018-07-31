package sp2p

import (
	"time"
	"github.com/inconshreveable/log15"
	"net"
	"github.com/kooksee/kdb"
	"path/filepath"
	"github.com/patrickmn/go-cache"
	"sync"
)

var (
	cfg  *kConfig
	once sync.Once
)

type kConfig struct {
	// 接收数据的最大缓存区
	MaxBufLen int

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

	conn      net.Conn
	localNode *node

	// －－－－－－－－－－－

	localConn *net.UDPConn
	uuidC     chan string
	db        kdb.IKDB
	l         log15.Logger
	cache     *cache.Cache
	p2p       ISP2P
}

func (t *kConfig) InitConn(conn net.Conn) *kConfig {
	t.conn = conn
	return t
}

func getConn() net.Conn {
	if cfg.conn == nil {
		panic("please init conn")
	}
	return cfg.conn
}

// 生成本地发送地址
func getLocalConn() *net.UDPConn {
	if cfg.localConn == nil {
		uad, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
		if err != nil {
			panic(err.Error())
		}
		cnn, err := net.ListenUDP("udp", uad)
		if err != nil {
			panic(err.Error())
		}
		cfg.localConn = cnn
	}
	return cfg.localConn
}

func (t *kConfig) InitLocalNode(nodeUrl string) *kConfig {
	t.localNode = MustNodeParse(nodeUrl)
	return t
}

func getLocalNode() *node {
	if getCfg().localNode == nil {
		panic("please local node")
	}
	return getCfg().localNode
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

func CreateNode(id Hash, ip net.IP, udpPort uint16) string {
	return newNode(id, ip, udpPort).string()
}

func getLog() log15.Logger {
	if getCfg().l == nil {
		panic("please init sp2p log")
	}
	return getCfg().l
}

func getDb() kdb.IKDB {
	if getCfg().db == nil {
		panic("please init sp2p db")
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
	once.Do(func() {
		cfg = &kConfig{
			MaxBufLen:           1024 * 16,
			Alpha:               3,
			NodeResponseNumber:  8,
			NodeBroadcastNumber: 16,
			HashBits:            len(Hash{}) * 8,
			PingNodeNum:         8,
			FindNodeNUm:         20,
			ConnReadTimeout:     5 * time.Second,
			ConnWriteTimeout:    5 * time.Second,

			NodesBackupKey: "nbk:",

			PingTick:     time.NewTicker(10 * time.Minute),
			FindNodeTick: time.NewTicker(1 * time.Hour),
			NtpTick:      time.NewTicker(10 * time.Minute),

			MaxNodeSize: 2000,
			MinNodeSize: 100,


			BucketSize: 16,

			uuidC: make(chan string, 500),
			cache: cache.New(10*time.Minute, 30*time.Minute),
		}
	})
	return cfg
}
