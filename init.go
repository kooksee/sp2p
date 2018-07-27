package sp2p

import (
	"github.com/json-iterator/go"
	"github.com/kooksee/cmn"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

var f = cmn.F
var cond = cmn.If
var mustNotErr = cmn.Err.MustNotErr
var errs = cmn.Err.Err
var checkClockDrift = cmn.CheckClockDrift
var newKBuffer = cmn.NewKBuffer
var randBytes = cmn.Rand.RandBytes
var rand32 = cmn.Rand.Rand32
