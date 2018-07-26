package sp2p

func init() {
	getHManager().registry(
		pingReq{},
		findNodeReq{},
		findNodeResp{},
	)
}
