package core

//STATUS hold dns loader running status
const (
	STATUS_INIT     uint32 = 0
	STATUS_STARTING uint32 = 1
	STATUS_RUNNING  uint32 = 2
	STATUS_STOPPING uint32 = 3
	STATUS_STOPPED  uint32 = 4
)
const (
	CALL_NOT_FINISH   uint32 = 0
	CALL_SUCCESS_DONE uint32 = 1
	CALL_TIMEOUT      uint32 = 2
)
