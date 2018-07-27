package sp2p

type KMsg struct {
	ID   string   `json:"id"`
	TN   string   `json:"tn"`
	FN   string   `json:"fn"`
	Data IMessage `json:"data,omitempty"`
}

func (t *KMsg) Decode(msg []byte) error {
	dt := msg[0]
	if !hm.contain(dt) {
		return errs("msg type %s is nonexistent", dt)
	}

	t.Data = hm.getHandler(dt)
	return json.Unmarshal(msg[1:], t)
}

func (t *KMsg) FNode(msg []byte) error {
	dt := msg[0]
	if !hm.contain(dt) {
		return errs("msg type %s is nonexistent", dt)
	}

	t.Data = hm.getHandler(dt)
	return json.Unmarshal(msg[1:], t)
}

func (t *KMsg) Dumps() []byte {
	d, _ := json.Marshal(t)
	return append([]byte{t.Data.T()}, append(d, "\n"...)...)
}
