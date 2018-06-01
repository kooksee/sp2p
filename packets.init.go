package sp2p

func init() {
	hm := GetHManager()

	MustNotErr(hm.Registry(PingReqT, &PingReq{}))

	MustNotErr(hm.Registry(FindNodeReqT, &FindNodeReq{}))
	MustNotErr(hm.Registry(FindNodeRespT, &FindNodeResp{}))

	MustNotErr(hm.Registry(KVSetReqT, &KVSetReq{}))
	MustNotErr(hm.Registry(KVGetReqT, &KVGetReq{}))
	MustNotErr(hm.Registry(KVGetRespT, &KVGetResp{}))

	MustNotErr(hm.Registry(GKVSetReqT, &GKVSetReq{}))
	MustNotErr(hm.Registry(GKVGetReqT, &GKVGetReq{}))
	MustNotErr(hm.Registry(GKVGetRespT, &GKVGetResp{}))
}
