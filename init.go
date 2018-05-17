package sp2p

import (
	"github.com/json-iterator/go"
	"github.com/kooksee/log"
)

var (
	cfg    *KConfig
	json   = jsoniter.ConfigCompatibleWithStandardLibrary
	logger = log.New("package", "sp2p")
	hm     = GetHManager()
)

func SetCfg(cfg1 *KConfig) {
	cfg = cfg1
}
