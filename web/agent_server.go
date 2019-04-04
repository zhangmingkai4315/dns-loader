package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/zhangmingkai4315/dns-loader/core"

	"github.com/gorilla/mux"
	uuid "github.com/nu7hatch/gouuid"
	"github.com/unrolled/render"
)

// Agent define the agent  object
type Agent struct {
	ID     string
	Status Event
	config core.Configuration
}

var agent *Agent

// NewAgent create new agent
func NewAgent() *Agent {
	id, _ := uuid.NewV4()
	return &Agent{
		ID:     id.String(),
		Status: Ready,
	}
}
func ping(w http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})
	r.JSON(w, http.StatusOK, map[string]string{"id": agent.ID, "status": "success", "message": "pong"})
	log.Printf("ping request from %s\n", req.RemoteAddr)
	return
}

func startAgentTraffic(w http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})
	decoder := json.NewDecoder(req.Body)
	// 判断是否有在工作的发包程序
	status := core.GetGlobalStatus()
	if status != core.STATUS_STOPPED && status != core.STATUS_INIT {
		r.JSON(w, http.StatusBadRequest, map[string]string{"status": "error", "message": "please make sure no job is running"})
		return
	}
	var config core.Configuration
	err := decoder.Decode(&config)
	if err != nil {
		r.JSON(w, http.StatusBadRequest, map[string]string{"status": "error", "message": "decode config fail"})
		return
	}
	// localTraffic
	err = config.Valid()
	if err != nil {
		log.Println(err)
		r.JSON(w, http.StatusBadRequest, map[string]string{"status": "error", "message": "validate config fail"})
		return
	}
	log.Printf("receive new job id:%s\n", config.ID)
	go core.GenTrafficFromConfig(&config)
	r.JSON(w, http.StatusOK, map[string]string{"status": "success"})

}

func killAgentTraffic(w http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})
	if stopStatus := core.GloablGenerator.Stop(); true != stopStatus {
		r.JSON(w, http.StatusInternalServerError, map[string]string{"status": "error", "message": "ServerFail"})
		return
	}
	r.JSON(w, http.StatusOK, map[string]string{"status": "success"})
}

// NewAgentServer function create the http api
func NewAgentServer(config *core.Configuration) {
	agent = NewAgent()
	r := mux.NewRouter()
	r.HandleFunc("/ping", ping).Methods("GET")
	r.HandleFunc("/start", startAgentTraffic).Methods("POST")
	r.HandleFunc("/kill", killAgentTraffic).Methods("GET")
	log.Println("agent server route init success")
	http.ListenAndServe(fmt.Sprintf(":%d", config.AgentPort), http.TimeoutHandler(r, time.Second*10, "timeout"))
}
