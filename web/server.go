package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	log "github.com/sirupsen/logrus"
	"github.com/unrolled/render"
	"github.com/zhangmingkai4315/dns-loader/core"
)

var currentPath string
var store *sessions.CookieStore
var nodeManager *NodeManager

// NewServer function

func auth(f func(w http.ResponseWriter, req *http.Request)) func(w http.ResponseWriter, req *http.Request) {

	return func(w http.ResponseWriter, req *http.Request) {
		session, _ := store.Get(req, "dns-loader")
		if user, ok := session.Values["username"].(string); !ok || user == "" {
			http.Redirect(w, req, "/login", 302)
			return
		}
		f(w, req)
	}

}

func index(w http.ResponseWriter, req *http.Request) {

	data := map[string]interface{}{
		"iplist": nodeManager.IPList,
	}
	r := render.New(render.Options{})
	r.HTML(w, http.StatusOK, "index", data)
}

func addNode(w http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})
	decoder := json.NewDecoder(req.Body)
	var ipinfo IPWithPort
	err := decoder.Decode(&ipinfo)
	if err != nil {
		r.JSON(w, http.StatusBadRequest, map[string]string{"status": "error", "message": "decode data fail"})
		return
	}
	err = nodeManager.AddNode(ipinfo.IPAddress, ipinfo.Port)
	if err != nil {
		r.JSON(w, http.StatusBadRequest, map[string]string{"status": "error", "message": err.Error()})
		return
	}
	r.JSON(w, http.StatusOK, map[string]string{"status": "success"})
}

func pingNode(w http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})
	decoder := json.NewDecoder(req.Body)
	var ipinfo IPWithPort
	err := decoder.Decode(&ipinfo)
	if err != nil {
		r.JSON(w, http.StatusBadRequest, map[string]string{"status": "error", "message": err.Error()})
		return
	}
	ip := fmt.Sprintf("%s:%d", ipinfo.IPAddress, ipinfo.Port)
	err = nodeManager.callPing(ip)
	if err != nil {
		r.JSON(w, http.StatusBadRequest, map[string]string{"status": "error", "message": err.Error()})
		return
	}
	r.JSON(w, http.StatusOK, map[string]string{"status": "success"})
}

func deleteNode(w http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})
	decoder := json.NewDecoder(req.Body)
	var ipinfo IPWithPort
	err := decoder.Decode(&ipinfo)
	if err != nil {
		r.JSON(w, http.StatusBadRequest, map[string]string{"status": "error", "message": err.Error()})
		return
	}
	pending := fmt.Sprintf("%s:%d", ipinfo.IPAddress, ipinfo.Port)
	err = nodeManager.Remove(pending)
	if err != nil {
		r.JSON(w, http.StatusBadRequest, map[string]string{"status": "error", "message": err.Error()})
		return
	}
	r.JSON(w, http.StatusOK, map[string]string{"status": "success"})
}

func startDNSTraffic(w http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})
	config := core.GetGlobalConfig()
	if config.Status != core.StatusStopped {
		r.JSON(w, http.StatusBadRequest, map[string]string{"status": "error", "message": "please make sure no job is running"})
		return
	}
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&config)
	if err != nil {
		log.Errorf("decode configuration info fail: %s", err.Error())
		r.JSON(w, http.StatusBadRequest, map[string]string{"status": "error", "message": "decode config fail"})
		return
	}

	err = config.ValidateConfiguration()
	if err != nil {
		log.Println(err)
		r.JSON(w, http.StatusBadRequest, map[string]string{"status": "error", "message": "validate config fail"})
		return
	}
	go core.GenTrafficFromConfig(config)
	go nodeManager.Call(Start, config)
	r.JSON(w, http.StatusOK, map[string]string{"id": config.ID, "status": "success"})

}

func stopDNSTraffic(w http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})
	if core.GloablGenerator == nil || core.GloablGenerator.Status() != core.StatusRunning {
		r.JSON(w, http.StatusInternalServerError, map[string]string{"status": "error", "message": "Not Running"})
		return
	}
	if stopStatus := core.GloablGenerator.Stop(); true != stopStatus {
		r.JSON(w, http.StatusInternalServerError, map[string]string{"status": "error", "message": "Server Fail"})
		return
	}
	go nodeManager.Call(Kill, nil)
	r.JSON(w, http.StatusOK, map[string]string{"status": "success"})
}

type StatusResponse struct {
	CurrentMessages []Message `json:"messages"`
	ID              string    `json:"id"`
	Status          string    `json:"status"`
	Error           string    `json:"error"`
}

func getCurrentStatus(w http.ResponseWriter, req *http.Request) {
	config := core.GetGlobalConfig()
	r := render.New(render.Options{})
	data, err := MessagesHub.Get()
	if err != nil {
		r.JSON(w, http.StatusServiceUnavailable, StatusResponse{Error: err.Error()})
	}
	messages := []Message{}
	if len(data) != 0 {
		err = json.Unmarshal(data, &messages)
		if err != nil {
			r.JSON(w, http.StatusServiceUnavailable, StatusResponse{Error: err.Error()})
			return
		}
	}
	r.JSON(w, http.StatusOK, StatusResponse{
		CurrentMessages: messages,
		ID:              config.ID,
		Status:          config.GetCurrentJobStatusString(),
	})
}

func login(config *core.Configuration) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		session, _ := store.Get(req, "dns-loader")
		if user, ok := session.Values["username"].(string); ok && user != "" {
			http.Redirect(w, req, "/", 302)
			return
		}
		if req.Method == "POST" {
			user := req.FormValue("username")
			password := req.FormValue("password")
			if user == config.User && password == config.Password {
				session.Values["username"] = user
				session.Save(req, w)
				http.Redirect(w, req, "/", 302)
			} else {
				http.Redirect(w, req, "/login", 401)
			}
		}
		if req.Method == "GET" {
			r := render.New(render.Options{})
			r.HTML(w, http.StatusOK, "login", nil)
		}
	}
}
func logout(w http.ResponseWriter, req *http.Request) {
	session, _ := store.Get(req, "dns-loader")
	if user, ok := session.Values["username"].(string); ok && user != "" {
		session.Values["username"] = ""
		session.Save(req, w)
		http.Redirect(w, req, "/login", 302)
		return
	}
}

// NewServer function create the http api
func NewServer() error {
	config := core.GetGlobalConfig()
	key := []byte(config.AppSecrect)
	nodeManager = NewNodeManager(config)
	store = sessions.NewCookieStore(key)
	r := mux.NewRouter()
	r.HandleFunc("/", auth(index)).Methods("GET")
	r.HandleFunc("/logout", logout).Methods("POST", "GET")
	r.HandleFunc("/login", login(config)).Methods("GET", "POST")
	r.HandleFunc("/nodes", auth(addNode)).Methods("POST")
	r.HandleFunc("/nodes", auth(deleteNode)).Methods("DELETE")
	r.HandleFunc("/ping", auth(pingNode)).Methods("POST")
	r.HandleFunc("/start", auth(startDNSTraffic)).Methods("POST")
	r.HandleFunc("/stop", auth(stopDNSTraffic)).Methods("GET")
	r.HandleFunc("/status", (getCurrentStatus)).Methods("GET")
	r.PathPrefix("/public/").Handler(http.StripPrefix("/public", http.FileServer(http.Dir("./web/assets"))))
	err := http.ListenAndServe(config.HTTPServer, http.TimeoutHandler(r, time.Second*10, "timeout"))
	return err
}
