package sp2p

import (
	"github.com/json-iterator/go"
	"github.com/kooksee/log"
	"os"
)

var (
	cfg    *KConfig
	json   = jsoniter.ConfigCompatibleWithStandardLibrary
	logger = log.New("package", "sp2p")
)

func SetCfg(cfg1 *KConfig) {
	cfg = cfg1

	if cfg.LogLevel != "error" {
		ll, err := log.LvlFromString(cfg.LogLevel)
		if err != nil {
			panic(err.Error())
		}
		logger.SetHandler(log.LvlFilterHandler(ll, log.StreamHandler(os.Stdout, log.TerminalFormat(true))))
	} else {
		logger.SetHandler(log.LvlFilterHandler(log.LvlError, log.StreamHandler(os.Stderr, log.LogfmtFormat())))
	}
}

func init() {
	hm := GetHManager()

	a1 := byte(0x1)
	MustNotErr(hm.Registry(a1, &PingReq{t: a1, s: "ping req"}))

	a2 := byte(0x2)
	MustNotErr(hm.Registry(a2, &FindNodeReq{t: a2, s: "find node req"}))

	a3 := byte(0x3)
	MustNotErr(hm.Registry(a3, &FindNodeResp{t: a3, s: "find node resp"}))

	a4 := byte(0x4)
	MustNotErr(hm.Registry(a4, &KVSetReq{t: a4, s: "kv set req"}))

	a5 := byte(0x5)
	MustNotErr(hm.Registry(a5, &KVGetReq{t: a5, s: "kv get req"}))

	a6 := byte(0x5)
	MustNotErr(hm.Registry(a6, &KVGetResp{t: a6, s: "kv get resp"}))
}
