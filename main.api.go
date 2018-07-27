package sp2p

func (s *sp2p) Write(msg *KMsg) {
	s.writeTx(msg)
}

func (s *sp2p) SelfNode() string {
	return s.tab.selfNode.string()
}

func (s *sp2p) NodeDumps() []string {
	return s.tab.getRawNodes()
}

func (s *sp2p) NodeUpdate(rawUrl ... string) error {
	for _, url := range rawUrl {
		n, err := NodeParse(url)
		if err != nil {
			return err
		}
		s.tab.updateNode(n)
	}
	return nil
}
func (s *sp2p) NodeDel(nodeID ... string) error {
	for _, id := range nodeID {
		n, err := HexToHash(id)
		if err != nil {
			return err
		}
		s.tab.deleteNode(n)
	}
	return nil
}

func (s *sp2p) FindMinDisNodes(targetID string, n int) (nodes []string, err error) {
	h, err := HexToHash(targetID)
	if err != nil {
		return nil, err
	}

	for _, n := range s.tab.findMinDisNodes(h, n) {
		nodes = append(nodes, n.string())
	}
	return nodes, nil
}

func (s *sp2p) FindRandomNodes(n int) (nodes []string) {
	for _, n := range s.tab.findRandomNodes(n) {
		nodes = append(nodes, n.string())
	}
	return
}

func (s *sp2p) FindNodeWithTargetBySelf(d string) (nodes []string) {
	for _, n := range s.tab.findNodeWithTargetBySelf(StringToHash(d)) {
		nodes = append(nodes, n.string())
	}
	return
}

func (s *sp2p) FindNodeWithTarget(targetId string, measure string) (nodes []string) {
	for _, n := range s.tab.findNodeWithTarget(StringToHash(targetId), StringToHash(measure)) {
		nodes = append(nodes, n.string())
	}
	return
}

func (s *sp2p) PingRandom() {
	go s.pingRandom()
}

func (s *sp2p) FindRandom() {
	go s.findRandom()
}

func (s *sp2p) Broadcast(msg *KMsg) {
	for _, n := range s.tab.getAllNodes() {
		msg.TN = n.string()
		s.writeTx(msg)
	}
}

func (s *sp2p) InitSeeds(seeds []string)  {
	for _,n:=range seeds{
		mustNotErr(s.NodeUpdate(n))
	}
}
