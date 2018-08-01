package sp2p

type KMsg struct {
	ID   string   `json:"id"`
	TN   string   `json:"tn"`
	FN   string   `json:"fn"`
	Time string   `json:"time"`
	Data IMessage `json:"data,omitempty"`
}

func (t *KMsg) Decode(msg []byte) error {
	dt := msg[0]
	if !hm.contain(dt) {
		return errs("msg type %s is nonexistent", dt)
	}

	t.Data = hm.getHandler(dt)
	return jsonUnmarshal(msg[1:], t)
}

func (t *KMsg) FNode(msg []byte) error {
	dt := msg[0]
	if !hm.contain(dt) {
		return errs("msg type %s is nonexistent", dt)
	}

	t.Data = hm.getHandler(dt)
	return jsonUnmarshal(msg[1:], t)
}

func (t *KMsg) Dumps() []byte {
	d, _ := jsonMarshal(t)
	return append([]byte{t.Data.T()}, append(d, "\n"...)...)
}
