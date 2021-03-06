package sp2p

func (s *sp2p) Write(msg *KMsg) {
	go s.writeTx(msg)
}

func (s *sp2p) GetSelfNode() string {
	return s.tab.selfNode.string()
}

func (s *sp2p) GetNodes() []string {
	return s.tab.getRawNodes()
}

func (s *sp2p) TableSize() int {
	return s.tab.size()
}

func (s *sp2p) UpdateNode(rawUrl string) error {
	n, err := NodeParse(rawUrl)
	if err != nil {
		return err
	}
	s.tab.updateNode(n)
	return nil
}
func (s *sp2p) DeleteNode(id string) error {
	n, err := HexToHash(id)
	if err != nil {
		return err
	}
	s.tab.deleteNode(n)
	return nil
}

func (s *sp2p) AddNode(rawUrl string) error {
	n, err := NodeParse(rawUrl)
	if err != nil {
		return err
	}
	s.tab.addNode(n)
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

func (s *sp2p) PingN() {
	go s.pingN()
}

func (s *sp2p) FindN() {
	go s.findN()
}

func (s *sp2p) Broadcast(msg *KMsg) {
	for _, n := range s.tab.getAllNodes() {
		msg.TAddr = n.addrString()
		msg.TID = n.ID.Hex()
		s.writeTx(msg)
	}
}
