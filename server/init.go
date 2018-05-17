package server

import (
	"time"

	"github.com/json-iterator/go"
	"github.com/kooksee/log"
	"github.com/kooksee/srelay/config"
)

var (
	cfg    *config.Config
	json   = jsoniter.ConfigCompatibleWithStandardLibrary
	logger = log.New("package", "server")
)

const (
	connReadTimeout = 10 * time.Second
)

func SetCfg(cfg1 *config.Config) {
	cfg = cfg1
}
