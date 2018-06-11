package sp2p

func init() {
	hm := GetHManager()

	MustNotErr(hm.Registry(
		PingReq{},
		FindNodeReq{},
		FindNodeResp{},

		KVSetReq{},
		KVGetReq{},
		KVGetResp{},

		GKVSetReq{},
		GKVGetReq{},
		GKVGetResp{},
	))
}
