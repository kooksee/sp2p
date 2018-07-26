package sp2p

type findNodeReq struct {
	N int `json:"n,omitempty"`
}

func (t *findNodeReq) T() byte        { return FindNodeReqT }
func (t *findNodeReq) String() string { return FindNodeReqS }
func (t *findNodeReq) OnHandle(p *SP2p, msg *KMsg) {

	node, err := nodeFromKMsg(msg)
	if err != nil {
		getLog().Error("NodeFromKMsg error", "err", err)
		return
	}
	go p.tab.UpdateNode(node)

	ns := make([]string, 0)

	// 最多不能超过16
	if t.N > 16 {
		t.N = 16
	}

	for _, n := range p.tab.FindMinDisNodes(node.ID, t.N) {
		ns = append(ns, n.String())
	}
	p.Write(&KMsg{TAddr: msg.FAddr, Data: &findNodeResp{Nodes: ns}})
}

type findNodeResp struct {
	Nodes []string `json:"nodes,omitempty"`
}

func (t *findNodeResp) T() byte        { return FindNodeRespT }
func (t *findNodeResp) String() string { return FindNodeRespS }
func (t *findNodeResp) OnHandle(p *SP2p, msg *KMsg) {
	for _, n := range t.Nodes {
		node, err := NodeParse(n)
		if err != nil {
			getLog().Error("parse node error", "err", err)
			continue
		}
		p.tab.UpdateNode(node)
	}
}
