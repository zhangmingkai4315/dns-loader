package core

type ReturnCode int

// Return Code
const (
	RET_NO_ERROR   ReturnCode = 0
	RET_SERVFAIL              = 0x2
	RET_FORMERR               = 0x1
	RET_NXDOMAIN              = 0x3
	RET_NOTIMP                = 0x4
	RET_REFUSED               = 0x5
	RET_TIMEOUT               = 0xf0
	RET_CALL_ERROR            = 0xf1
)

//dns loader running status
const (
	STATUS_ORIGINAL uint32 = 0
	STATUS_STARTING uint32 = 0
	STATUS_STARTED  uint32 = 0
	STATUS_STOPPING uint32 = 0
	STATUS_STOPPED  uint32 = 0
)

const (
	CALL_NOT_FINISH   uint32 = 0
	CALL_SUCCESS_DONE uint32 = 1
	CALL_TIMEOUT      uint32 = 2
)
