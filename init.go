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
