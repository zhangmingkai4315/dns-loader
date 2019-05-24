package core

//Status define the current dns loader running status
const (
	StatusStart    uint32 = 1
	StatusRunning  uint32 = 2
	StatusStopping uint32 = 3
	StatusStopped  uint32 = 4
)

// StatusToString store the status code to string info
var StatusToString = map[uint32]string{
	StatusStart:    "start",
	StatusRunning:  "running",
	StatusStopping: "stopping",
	StatusStopped:  "stopped",
}
