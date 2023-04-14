package xbee

const (
	// commands sent to the xbee s3b
	ATCmd                    = 0x08
	ATCmdQueueParameterValue = 0x09
	TxReq                    = 0x10
	TxReqExpl                = 0x11
	RemoteCmdReq             = 0x17
	// commands recieved from the xbee
	ATCmdResponse = 0x88
	ModemStatus   = 0x8A
	TxStatus      = 0x8B
	RouteInfoPkt  = 0x8D
	AddrUpdate    = 0x8E
	RxPkt         = 0x90
	RxPktExpl     = 0x91
	IOSample      = 0x92
)
