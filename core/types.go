package core

// Event define the event send to all node
type Event uint8

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

const (
	// Ready usually for listening status
	Ready Event = iota
	// Start and send the new config
	Start
	// Status the status
	Status
	// Running status with some message
	Running
	// Kill the load
	Kill
	// Error Status
	Error
	// Stop the load in normal way
	Stop
)
