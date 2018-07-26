package sp2p

import (
	"errors"
)

type KMsg struct {
	Version string   `json:"version,omitempty"`
	ID      string   `json:"id"`
	TID     string   `json:"tid"`
	TAddr   string   `json:"taddr,omitempty"`
	FAddr   string   `json:"faddr,omitempty"`
	FID     string   `json:"fid,omitempty"`
	Data    IMessage `json:"data,omitempty"`
}

func (t *KMsg) Decode(msg []byte) error {
	dt := msg[0]
	if !hm.contain(dt) {
		return errors.New(f("msg type %s is nonexistent", dt))
	}

	t.Data = hm.getHandler(dt)
	return json.Unmarshal(msg[1:], t)
}

func (t *KMsg) Dumps() []byte {
	d, _ := json.Marshal(t)
	return append([]byte{t.Data.T()}, append(d, "\n"...)...)
}
