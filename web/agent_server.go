package web

import (
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/zhangmingkai4315/dns-loader/core"

	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

func getAgentStatus(w http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})
	config := core.GetGlobalConfig()
	r.JSON(w, http.StatusOK, JSONResponse{
		ID:     config.JobID,
		Status: config.GetCurrentJobStatusString(),
	})
}

// NewAgentServer function create the http api
func NewAgentServer(host, port string) {
	r := mux.NewRouter()
	r.HandleFunc("/start", startDNSTraffic).Methods("POST")
	r.HandleFunc("/status", getAgentStatus).Methods("GET")
	r.HandleFunc("/stop", stopDNSTraffic).Methods("GET")
	err := http.ListenAndServe(fmt.Sprintf("%s:%s", host, port), http.TimeoutHandler(r, time.Second*10, "timeout"))
	if err != nil {
		log.Errorf("start agent server fail: %s", err)
	}
}
