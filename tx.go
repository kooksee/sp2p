package sp2p

import (
	"bytes"
	"io"
	"io/ioutil"
	"github.com/kataras/iris/core/errors"
	"fmt"
)

type KMsg struct {
	Version string   `json:"version,omitempty"`
	ID      string   `json:"id,omitempty"`
	TAddr   string   `json:"taddr,omitempty"`
	FAddr   string   `json:"faddr,omitempty"`
	FID     string   `json:"fid,omitempty"`
	Data    IMessage `json:"data,omitempty"`
}

func (t *KMsg) DecodeFromConn(r io.Reader) error {
	message, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	message = bytes.TrimSpace(message)
	return json.Unmarshal(message, t)
}

func (t *KMsg) Decode(msg []byte) error {
	dt := msg[0]
	if !hm.Contain(dt) {
		return errors.New(fmt.Sprintf("not existed"))
	}

	t.Data = hm.GetHandler(dt)
	return json.Unmarshal(msg[1:], t)
}

func (t *KMsg) Dumps() []byte {
	d, _ := json.Marshal(t)
	return append([]byte{t.Data.T()}, append(d, '\n')...)
}
