package sp2p

import (
	"bytes"
	"io"
	"io/ioutil"
)

type KMsg struct {
	Version string      `json:"version,omitempty"`
	ID      string      `json:"id,omitempty"`
	TAddr   string      `json:"taddr,omitempty"`
	FAddr   string      `json:"faddr,omitempty"`
	FID     string      `json:"fid,omitempty"`
	Event   string      `json:"event,omitempty"`
	Data    interface{} `json:"data,omitempty"`
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
	return json.Unmarshal(msg, t)
}

func (t *KMsg) Dumps() []byte {
	d, _ := json.Marshal(t)
	return append(d, '\n')
}
