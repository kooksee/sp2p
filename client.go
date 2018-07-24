package sp2p

func DecodeClient(data []byte) (CKMsg, error) {
	c := CKMsg{}
	return c, json.Unmarshal(data, &c)
}

type CKMsg struct {
	ID   string `json:"id,omitempty"`
	Addr string `json:"addr,omitempty"`
	Data []byte `json:"data,omitempty"`
}

func (c CKMsg) Bytes() []byte {
	d, _ := json.Marshal(c)
	return append(d, "\n"...)
}

type ErrCode struct {
	Code int    `json:"code,omitempty"`
	Msg  string `json:"msg,omitempty"`
}

func (e ErrCode) Bytes() []byte {
	d, _ := json.Marshal(e)
	return append(d, "\n"...)
}
