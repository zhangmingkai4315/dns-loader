package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"github.com/zhangmingkai4315/dns-loader/core"
)

// NodeManager define the node list
// when new config generated the manager will call the nodes one by one
type NodeManager struct {
	DB     *gorm.DB
	IPList []string
}

// NewNodeManager will create a new node manager
func NewNodeManager(c *core.Configuration) *NodeManager {
	dbHander := core.GetDBHandler()
	agents := []core.Agent{}
	dbHander.Model(&core.Agent{}).Find(&agents)
	iplist := []string{}
	for _, agent := range agents {
		iplist = append(iplist, agent.IP+":"+agent.Port)
	}
	return &NodeManager{
		DB:     core.GetDBHandler().DB,
		IPList: iplist,
	}
}

// SyncIPList sync db and memory
func (manager *NodeManager) SyncIPList() {
	agents := []core.Agent{}
	manager.DB.Model(&core.Agent{}).Find(&agents)
	iplist := []string{}
	for _, agent := range agents {
		iplist = append(iplist, agent.IP+":"+agent.Port)
	}
	manager.IPList = iplist
}

// GetAllNodeStatus check all node status
func (manager *NodeManager) GetAllNodeStatus() chan NodeInfo {
	nodeInfoChannel := make(chan NodeInfo, len(manager.IPList))
	for _, ip := range manager.IPList {
		go manager.callStatus(ip, nodeInfoChannel)
	}
	return nodeInfoChannel
}

// Size return how many nodes in managers
func (manager *NodeManager) Size() int {
	return len(manager.IPList)
}

// AddNode append a new ip to this node list
func (manager *NodeManager) AddNode(ip string, port string) error {
	newAgent := ip + ":" + port
	if core.StringInSlice(newAgent, manager.IPList) {
		return errors.New("Already in list")
	}
	err := manager.callPing(newAgent)
	if err != nil {
		return err
	}
	err = core.GetDBHandler().AddAgent(ip, port)
	if err != nil {
		return err
	}
	manager.IPList = append(manager.IPList, ip+":"+port)
	return nil
}

// Remove will remove the ip from current list
func (manager *NodeManager) Remove(ip string, port string) error {
	err := core.GetDBHandler().RemoveAgent(ip, port)
	if err != nil {
		return err
	}
	manager.IPList = core.RemoveStringInSlice(ip+":"+port, manager.IPList)
	return nil
}

// Call function will send data to all node
func (manager *NodeManager) Call(event Event, data interface{}) error {
	for _, ip := range manager.IPList {
		go func(ip string, event Event, data interface{}) {
			switch event {
			case Start:
				log.Printf("send job infomation to agent :%s", ip)
				err := manager.callStart(ip, data)
				if err != nil {
					log.Errorf("send job infomation to agent : %s fail:%s", ip, err.Error())
				}
			case Kill:
				log.Printf("send kill signal to agent :%s", ip)
				err := manager.callKill(ip)
				if err != nil {
					log.Errorf("send kill job command to agent : %s fail:%s", ip, err.Error())
				}
			case Ping:
				log.Printf("send ping signal to agent :%s", ip)
				err := manager.callPing(ip)
				if err != nil {
					log.Errorf("send ping command to agent : %s fail:%s", ip, err.Error())
				}
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
		Timeout: time.Second * 5,
	}
	config, ok := data.(*core.Configuration)
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

// callKill function will kill the process in one node
func (manager *NodeManager) callKill(ip string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic from rpc call[%s]\n", ip)
		}
	}()
	var netClient = &http.Client{
		Timeout: time.Second * 5,
	}
	response, err := netClient.Get(fmt.Sprintf("http://%s/stop", ip))
	if err != nil && response.StatusCode != 200 {
		return err
	}
	return nil
}

func (manager *NodeManager) callStatus(ip string, infoChannel chan NodeInfo) {
	ipAndPort := strings.Split(ip, ":")
	nodeInfo := NodeInfo{
		IPWithPort: IPWithPort{
			IPAddress: ipAndPort[0],
			Port:      ipAndPort[1],
		},
	}
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic from agent status call[%s]\n", ip)
			nodeInfo.Error = fmt.Sprintf("%v", r)
			infoChannel <- nodeInfo
		}
	}()
	var netClient = &http.Client{
		Timeout: time.Second * 1,
	}

	response, err := netClient.Get(fmt.Sprintf("http://%s/status", ip))
	if err != nil && response.StatusCode != 200 {
		nodeInfo.Error = err.Error()
	}
	defer response.Body.Close()
	infoData := &JSONResponse{}
	json.NewDecoder(response.Body).Decode(infoData)
	nodeInfo.Status = infoData.Status
	nodeInfo.JobID = infoData.ID
	nodeInfo.Error = infoData.Error
	infoChannel <- nodeInfo
}

// callCheckStatus function will check all the node with uuid
func (manager *NodeManager) callPing(ip string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic from agent ping call[%s]\n", r)
		}
	}()
	var netClient = &http.Client{
		Timeout: time.Second * 2,
	}
	response, err := netClient.Get(fmt.Sprintf("http://%s/ping", ip))
	if err != nil && response.StatusCode != 200 {
		return err
	}
	return nil
}
