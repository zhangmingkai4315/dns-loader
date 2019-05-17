package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/rpc"
	"time"

	uuid "github.com/nu7hatch/gouuid"
	log "github.com/sirupsen/logrus"
	"github.com/zhangmingkai4315/dns-loader/core"
)

// NodeStatus define the one of node status
type NodeStatus struct {
	Status    Event
	TimeStamp time.Time
	Message   string
}

// NewNodeStatus create the status from message
func NewNodeStatus(status Event, message string) *NodeStatus {
	return &NodeStatus{
		Status:    status,
		Message:   message,
		TimeStamp: time.Now(),
	}
}

// NodeManager define the rpc node list
// when new config generated the manager will call the nodes one by one
type NodeManager struct {
	IPList     []string
	IPStatus   map[string][]NodeStatus
	TaskStatus []string
	config     *core.Configuration
}

// NewNodeManager will create a new node manager
func NewNodeManager(c *core.Configuration) *NodeManager {
	ipstatus := make(map[string][]NodeStatus)
	return &NodeManager{
		IPList:     c.Agents,
		TaskStatus: []string{},
		IPStatus:   ipstatus,
		config:     c,
	}
}

// AddNode append a new ip to this node list
func (manager *NodeManager) AddNode(ip string, port string) error {
	if port == "" {
		port = manager.config.AgentPort
	}
	ip = fmt.Sprintf("%s:%s", ip, port)
	log.Printf("send ping request to node:%s", ip)
	err := manager.callPing(ip)
	if err != nil {
		return err
	}
	log.Println("ping agent success")
	if !core.StringInSlice(ip, manager.IPList) {
		manager.IPList = append(manager.IPList, ip)
		config := core.GetGlobalConfig()
		if err != nil {
			return err
		}
		return config.AddAgent(ip)
	} else {
		return errors.New("Already in list")
	}
	return nil
}

// AddStatus append a new ip to this node list
func (manager *NodeManager) AddStatus(ip string, status Event, message string) error {
	statuslist, ok := manager.IPStatus[ip]
	if ok != true {
		return errors.New("ip not exist")
	}
	nodeStatus := NewNodeStatus(status, message)
	statuslist = append(statuslist, *nodeStatus)
	return nil
}

// Remove will remove the ip from current list
func (manager *NodeManager) Remove(deleteip string) (err error) {
	manager.IPList = core.RemoveStringInSlice(deleteip, manager.IPList)
	delete(manager.IPStatus, deleteip)
	config := core.GetGlobalConfig()
	if err != nil {
		return err
	}
	return config.RemoveAgent(deleteip)
}

// Call function will send data to all node
func (manager *NodeManager) Call(event Event, data interface{}) error {

	for _, ip := range manager.IPList {
		go func(ip string, event Event, data interface{}) {
			switch event {
			case Start:
				log.Printf("send configuration to agent :%s", ip)
				manager.callStart(ip, data)
			case Check:
				log.Printf("send check signal to agent :%s", ip)
				manager.callCheckStatus(ip, event, data)
			case Kill:
				log.Printf("send kill signal to agent :%s", ip)
				manager.callKill(ip)
			case Ping:
				log.Printf("send ping signal to agent :%s", ip)
				manager.callPing(ip)
			}
		}(ip, event, data)
	}
	return nil
}

// callStart function will send data to all node with start evnet and config
func (manager *NodeManager) callStart(ip string, data interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic from agent ping call[%s]\n", r)
		}
	}()
	var netClient = &http.Client{
		Timeout: time.Second * 10,
	}

	config, ok := data.(core.Configuration)
	if ok != true {
		return errors.New("config data fail to send")
	}
	jsonData, err := json.Marshal(config)
	if err != nil {
		return err
	}
	response, err := netClient.Post(fmt.Sprintf("http://%s/start", ip), "application/json", bytes.NewBuffer(jsonData))
	if err != nil && response.StatusCode != 200 {
		return err
	}
	return nil
}

// callCheckStatus function will check all the node with uuid
func (manager *NodeManager) callCheckStatus(ip string, event Event, data interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic from rpc call[%s]\n", ip)
		}
	}()
	client, err := rpc.DialHTTP("tcp", ip)
	defer client.Close()
	if err != nil {
		log.Printf("call remote node rpc failed:[%s] %s\n", ip, err.Error())
		return
	}
	id, ok := data.(uuid.UUID)
	if ok != true {
		log.Println("config data fail to send")
		return
	}
	var result *RPCResult
	args := &RPCCall{
		Event: event,
		ID:    id,
	}
	err = client.Call("Manager.CheckStatus", args, &result)
	if err != nil {
		log.Printf("send check status to node faile:[%s] %s", ip, err.Error())
		return
	}
	return nil
}

// callCheckStatus function will check all the node with uuid
func (manager *NodeManager) callKill(ip string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic from rpc call[%s]\n", ip)
		}
	}()
	var netClient = &http.Client{
		Timeout: time.Second * 10,
	}
	response, err := netClient.Get(fmt.Sprintf("http://%s/kill", ip))
	if err != nil && response.StatusCode != 200 {
		return err
	}
	return nil
}

// callCheckStatus function will check all the node with uuid
func (manager *NodeManager) callPing(ip string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic from agent ping call[%s]\n", r)
		}
	}()
	var netClient = &http.Client{
		Timeout: time.Second * 10,
	}
	response, err := netClient.Get(fmt.Sprintf("http://%s/ping", ip))
	if err != nil && response.StatusCode != 200 {
		return err
	}
	return nil
}
