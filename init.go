package sp2p

import (
	"github.com/kooksee/cmn"
)

var f = cmn.F
var cond = cmn.If
var mustNotErr = cmn.Err.MustNotErr
var errs = cmn.Err.Err
var errPipe = cmn.Err.ErrWithMsg
var checkClockDrift = cmn.CheckClockDrift
var newKBuffer = cmn.NewKBuffer
var randBytes = cmn.Rand.RandBytes
var rand32 = cmn.Rand.Rand32
var jsonUnmarshal = cmn.Json.Unmarshal
var jsonMarshal = cmn.Json.Marshal
