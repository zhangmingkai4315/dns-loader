package web

import (
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

func getAgentStatus(w http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})
	config := core.GetGlobalConfig()
	r.JSON(w, http.StatusOK, JSONResponse{
		ID:     config.ID,
		Status: config.GetCurrentJobStatusString(),
	})
}

// NewAgentServer function create the http api
func NewAgentServer(host, port string) {
	agent = NewAgent()
	r := mux.NewRouter()
	r.HandleFunc("/ping", ping).Methods("GET")
	r.HandleFunc("/start", startDNSTraffic).Methods("POST")
	r.HandleFunc("/status", getAgentStatus).Methods("GET")
	r.HandleFunc("/stop", stopDNSTraffic).Methods("GET")
	log.Println("agent server route init success")
	http.ListenAndServe(fmt.Sprintf("%s:%s", host, port), http.TimeoutHandler(r, time.Second*10, "timeout"))
}
