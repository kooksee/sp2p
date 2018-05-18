package sp2p

type FindNodeReq struct {
	s string
	t byte
	N int `json:"k,omitempty"`
}

func (t *FindNodeReq) T() byte        { return t.t }
func (t *FindNodeReq) String() string { return t.s }
func (t *FindNodeReq) OnHandle(p *SP2p, msg *KMsg) {
	node, err := NodeFromKMsg(msg)
	if err != nil {
		logger.Error("NodeFromKMsg error", "err", err)
		return
	}
	go p.tab.UpdateNode(node)

	ns := make([]string, 0)
	for _, n := range p.tab.FindMinDisNodes(node.sha, t.N) {
		ns = append(ns, n.String())
	}
	p.Write(&KMsg{
		TAddr: msg.FAddr,
		Data:  &FindNodeResp{Nodes: ns},
	})
}

type FindNodeResp struct {
	s     string
	t     byte
	Nodes []string `json:"k,omitempty"`
}

func (t *FindNodeResp) T() byte        { return t.t }
func (t *FindNodeResp) String() string { return t.s }
func (t *FindNodeResp) OnHandle(p *SP2p, msg *KMsg) {
	for _, n := range t.Nodes {
		node, err := ParseNode(n)
		if err != nil {
			logger.Error("parse node error", "err", err)
			continue
		}
		p.tab.AddNode(node)
	}
}
