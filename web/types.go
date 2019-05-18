package web

import (
	"fmt"

	"github.com/asaskevich/govalidator"
	uuid "github.com/nu7hatch/gouuid"
	"github.com/zhangmingkai4315/dns-loader/core"
)

func init() {
	govalidator.SetFieldsRequiredByDefault(true)
}

// Event define the event send to all node
type Event uint8

const (
	// Ready usually for listening status
	Ready Event = iota
	// Start and send the new config
	Start
	// Check the status
	Check
	// Running status with some message
	Running
	// Kill the load
	Kill
	// Error Status
	Error
	// Stop the load in normal way
	Stop
	// Ping will do health check the status of node
	Ping
)

// IPWithPort define the posted node info
type IPWithPort struct {
	IPAddress string `json:"ipaddress" valid:"ip"`
	Port      string `json:"port" valid:"port"`
}

// NodeInfo for status check response
type NodeInfo struct {
	IPWithPort
	JobID  string `json:"job_id" valid:"-"`
	Status string `json:"status" valid:"-"`
	Error  string `json:"error" valid:"-"`
}

// Validate if ip and port infomation is valid
func (ipp *IPWithPort) Validate() error {
	_, err := govalidator.ValidateStruct(ipp)
	if err != nil {
		return err
	}
	return nil
}
func (ipp *IPWithPort) toString(defaultPort string) string {
	if ipp.Port == "" {
		ipp.Port = defaultPort
	}
	return fmt.Sprintf("%s:%s", ipp.IPAddress, ipp.Port)
}

// RPCCall define the message send to node
type RPCCall struct {
	ID      uuid.UUID
	Event   Event
	Config  core.Configuration
	Message string
}

// RPCResult define the result send from node
type RPCResult struct {
	ID      uuid.UUID
	Event   Event
	Message string
}
