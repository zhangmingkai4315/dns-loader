package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

// NodeManager define the node list
// when new config generated the manager will call the nodes one by one
type NodeManager struct {
	DB             *gorm.DB
	NodeInfos      map[string]NodeInfo
	nodeStatusChan chan NodeInfo
}

// NodeInfo hold current Node baisc info and job info
type NodeInfo struct {
	Agent
	JobID  string `json:"job_id" valid:"-"`
	Status string `json:"status" valid:"-"`
	Error  string `json:"error" valid:"-"`
}

var nodeManager *NodeManager

// GetNodeManager global manager interface
func GetNodeManager() *NodeManager {
	if nodeManager == nil {
		nodeManager = NewNodeManager()
	}
	return nodeManager
}

// NewNodeManager will create a new node manager
func NewNodeManager() *NodeManager {
	manager := NodeManager{
		DB:             GetDBHandler().DB,
		nodeStatusChan: make(chan NodeInfo),
		NodeInfos:      make(map[string]NodeInfo),
	}
	err := manager.SyncDBForAgents()
	if err != nil {
		return nil
	}
	statusCheckTicker := time.NewTicker(time.Second * 5)
	go func() {
		for {
			select {
			case <-statusCheckTicker.C:
				go manager.Call(Status, nil)
			case status := <-manager.nodeStatusChan:
				go manager.statusUpdate(status)
			}
		}
	}()
	return &manager
}

func (manager *NodeManager) statusUpdate(status NodeInfo) {
	statusKey := status.IPAddrWithPort()
	oldStatus, ok := manager.NodeInfos[statusKey]
	if ok == false {
		manager.NodeInfos[statusKey] = status
		return
	}
	if oldStatus.Live != status.Live {
		// update database status
		err := manager.UpdateLiveStatusAgent(status.Agent, status.Live)
		if err != nil {
			log.Errorf("update agent status fail: %s", err)
		}
	}
	oldStatus.Agent.Live = status.Live
	oldStatus.JobID = status.JobID
	oldStatus.Status = status.Status
	manager.NodeInfos[statusKey] = oldStatus
}

// SyncDBForAgents sync db to current list
func (manager *NodeManager) SyncDBForAgents() error {
	agents := []Agent{}
	err := manager.DB.Model(&Agent{}).Find(&agents).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	for _, agent := range agents {
		oldStatus, ok := manager.NodeInfos[agent.IPAddrWithPort()]
		if ok == true {
			oldStatus.Agent = agent
			manager.NodeInfos[agent.IPAddrWithPort()] = oldStatus
			continue
		}
		manager.NodeInfos[agent.IPAddrWithPort()] = NodeInfo{Agent: agent}
	}
	return nil
}

// AddNode append a new ip to this node list
func (manager *NodeManager) AddNode(ip string, port string) error {
	agent := Agent{IP: ip, Port: port}
	// only in check mode
	err := manager.callStatus(agent, true)
	if err != nil {
		return err
	}
	// err = GetDBHandler().AddAgent(ip, port)
	err = manager.DB.First(&agent).Error
	if err == nil {
		return fmt.Errorf("%s already exist", agent.IPAddrWithPort())
	}
	if !gorm.IsRecordNotFoundError(err) {
		return fmt.Errorf("save new agent fail: %s", err)
	}
	// insert a new agent
	agent.Enable = true
	agent.Live = true
	manager.DB.Save(&agent)
	return nil
}

// Agents get all agents info
func (manager *NodeManager) Agents() (agents []Agent) {
	for _, v := range manager.NodeInfos {
		agents = append(agents, v.Agent)
	}
	return
}

// RemoveNode will remove the ip from current list
func (manager *NodeManager) RemoveNode(ip string, port string) error {
	agent := Agent{
		IP:   ip,
		Port: port,
	}
	err := manager.DB.Unscoped().Delete(&agent).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return errors.New("agent not in database")
		}
		return fmt.Errorf("delete agent fail: %s", err)
	}
	return manager.SyncDBForAgents()
}

// Call function will send data to all agents
func (manager *NodeManager) Call(event Event, data interface{}) error {
	for _, nodeInfo := range manager.NodeInfos {
		agent := nodeInfo.Agent
		if event != Status && agent.Enable == false {
			log.Infof("skip agent :%s because it not enabled", agent.IPAddrWithPort())
			continue
		}
		switch event {
		case Start:
			log.Infof("send job infomation to agent :%s", agent.IPAddrWithPort())
			err := manager.callStart(agent, data)
			if err != nil {
				log.Errorf("send job infomation to agent : %s fail:%s", agent.IPAddrWithPort(), err.Error())
			}
		case Kill:
			log.Printf("send kill signal to agent :%s", agent.IPAddrWithPort())
			err := manager.callKill(agent)
			if err != nil {
				log.Errorf("send kill job command to agent : %s fail:%s", agent.IPAddrWithPort(), err.Error())
			}
		case Status:
			manager.callStatus(agent, false)
		}
	}
	return nil
}

// callStart function will send data to all node with start evnet and config
func (manager *NodeManager) callStart(agent Agent, data interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic for agent start job call[%s]\n", r)
		}
	}()
	var ip = agent.IPAddrWithPort()
	var netClient = &http.Client{
		Timeout: time.Second * 5,
	}
	config, ok := data.(JobConfig)
	if ok != true {
		return errors.New("config data fail to send")
	}
	log.Infof("%+v", config)
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
func (manager *NodeManager) callKill(agent Agent) (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic for kill job on agent [%s]\n", agent.IP+":"+agent.Port)
		}
	}()
	var netClient = &http.Client{
		Timeout: time.Second * 5,
	}
	response, err := netClient.Get(fmt.Sprintf("http://%s/stop", agent.IP+":"+agent.Port))
	if err != nil && response.StatusCode != 200 {
		return err
	}
	return nil
}

func (manager *NodeManager) callStatus(agent Agent, checkOnly bool) error {
	nodeInfo := NodeInfo{
		Agent: agent,
	}
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("panic from agent status call[%s]\n", agent.IP+":"+agent.Port)
			nodeInfo.Error = fmt.Sprintf("%v", r)
			if checkOnly == false {
				nodeInfo.Live = false
				manager.nodeStatusChan <- nodeInfo
			}
		}
	}()
	var netClient = &http.Client{
		Timeout: time.Second * 1,
	}

	response, err := netClient.Get(fmt.Sprintf("http://%s/status", agent.IP+":"+agent.Port))
	if err != nil || response.StatusCode != 200 {
		nodeInfo.Error = err.Error()
		nodeInfo.Live = false
		log.Errorf("get status from [%s] fail:%s", agent.IPAddrWithPort(), err)
		if checkOnly == true {
			return err
		}
		manager.nodeStatusChan <- nodeInfo
		return nil
	}

	defer response.Body.Close()
	infoData := &AgentStatusJSONResponse{}
	json.NewDecoder(response.Body).Decode(infoData)
	nodeInfo.Status = infoData.Status
	nodeInfo.JobID = infoData.ID
	nodeInfo.Live = true
	nodeInfo.Error = infoData.Error
	if checkOnly == false {
		manager.nodeStatusChan <- nodeInfo
	}
	return nil
}

// UpdateEnableStatusAgent will enable or disable one agent when using benchmark
func (manager *NodeManager) UpdateEnableStatusAgent(ip string, port string, enable bool) error {
	agent := Agent{
		IP:   ip,
		Port: port,
	}
	err := manager.DB.First(&agent).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return errors.New("agent not in database")
		}
		return fmt.Errorf("disable agent fail: %s", err)
	}
	agent.Enable = enable
	manager.DB.Save(&agent)
	manager.SyncDBForAgents()
	log.Infof("update enable status for %s to [%v]", agent.IPAddrWithPort(), agent.Enable)
	return nil
}

// GetEnabledStatusAgent get all enabled status
func (manager *NodeManager) GetEnabledStatusAgent() ([]Agent, error) {
	agents := []Agent{}
	err := manager.DB.Where("enable = ?", true).Find(&agents).Error
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, fmt.Errorf("get enabled agents fail: %s", err)
	}
	return agents, err
}

// UpdateLiveStatusAgent will set live or dead status on one agent when using benchmark
func (manager *NodeManager) UpdateLiveStatusAgent(agent Agent, live bool) error {
	err := manager.DB.First(&agent).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return errors.New("agent not in database")
		}
		return fmt.Errorf("disable agent fail: %s", err)
	}
	agent.Live = live
	log.Infof("update agent %s :live status to %v", agent.IPAddrWithPort(), live)
	manager.DB.Save(&agent)
	return nil
}
