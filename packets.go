package sp2p

const (
	PingReqT = byte(0x1)
	PingReqS = "ping req"

	FindNodeReqT = byte(0x2)
	FindNodeReqS = "find node req"

	FindNodeRespT = byte(0x3)
	FindNodeRespS = "find node resp"

	KVSetReqT = byte(0x4)
	KVSetReqS = "kv set req"

	KVGetReqT = byte(0x5)
	KVGetReqS = "kv get req"

	KVGetRespT = byte(0x6)
	KVGetRespS = "kv get resp"
)
