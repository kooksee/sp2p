package sp2p

func init() {
	GetHManager().Registry(
		pingReq{},
		findNodeReq{},
		findNodeResp{},
	)
}
