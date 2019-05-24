package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	log "github.com/sirupsen/logrus"
	"github.com/unrolled/render"
	"github.com/zhangmingkai4315/dns-loader/core"
)

var currentPath string
var store *sessions.CookieStore

// JSONResponse the response to front end
type JSONResponse struct {
	CurrentMessages []Message       `json:"messages"`
	ID              string          `json:"id"`
	Status          string          `json:"status"`
	Error           string          `json:"error"`
	NodeInfos       []core.NodeInfo `json:"nodes"`
}

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

func login(app *core.AppController) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		session, _ := store.Get(req, "dns-loader")
		if user, ok := session.Values["username"].(string); ok && user != "" {
			http.Redirect(w, req, "/", 302)
			return
		}
		if req.Method == "POST" {
			user := req.FormValue("username")
			password := req.FormValue("password")
			if user == app.AppConfig.User && password == app.AppConfig.Password {
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

func index(w http.ResponseWriter, req *http.Request) {
	nodeManager := core.GetNodeManager()
	data := map[string]interface{}{
		"agents": nodeManager.Agents(),
	}
	r := render.New(render.Options{})
	r.HTML(w, http.StatusOK, "index", data)
}

func startDNSTraffic(w http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})
	app := core.GetGlobalAppController()
	if app.Status != core.StatusStopped {
		r.JSON(w, http.StatusBadRequest, JSONResponse{
			Error:  "benchmark is not ready",
			ID:     app.JobID,
			Status: app.GetCurrentJobStatusString(),
		})
		log.Errorln("start fail: benchmark is not ready")
		return
	}
	job := core.JobConfig{}
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&job)
	if err != nil {
		log.Errorf("decode configuration info fail: %s", err.Error())
		r.JSON(w, http.StatusBadRequest, JSONResponse{
			Error: "decode request infomation fail",
		})
		log.Errorf("decode post request infomation fail:%s", err)
		return
	}
	err = job.ValidateJob()
	if err != nil {
		r.JSON(w, http.StatusBadRequest, JSONResponse{
			Error: err.Error(),
		})
		log.Errorf("validate post infomation fail:%s", err)
		return
	}
	app.JobConfig = &job
	if app.IsMaster == true {
		nodeManager := core.GetNodeManager()
		err := core.GetDBHandler().CreateDNSQueryHistory(app)
		if err != nil {
			log.Errorf("save query histroy fail:%s", err)
		}
		log.Infoln("master send new query job to agents")
		go nodeManager.Call(core.Start, job)
	} else {
		log.Infof("agent receive new query job from %s", req.RemoteAddr)
	}

	go core.GenTrafficFromConfig(app)

	r.JSON(w, http.StatusOK, JSONResponse{
		ID:     app.JobConfig.JobID,
		Status: app.GetCurrentJobStatusString(),
	})
}

func stopDNSTraffic(w http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})
	app := core.GetGlobalAppController()
	if app.LoadManager == nil || app.LoadManager.Status() != core.StatusRunning {
		r.JSON(w, http.StatusBadRequest, JSONResponse{
			Error: "job is already stopped",
		})
		return
	}
	if stopStatus := app.LoadManager.Stop(); true != stopStatus {
		r.JSON(w, http.StatusInternalServerError, JSONResponse{
			Error: "server fail, please try again later",
		})
		return
	}
	if app.IsMaster == true {
		nodeManager := core.GetNodeManager()
		go nodeManager.Call(core.Kill, nil)
	}
	r.JSON(w, http.StatusOK, JSONResponse{
		ID:     app.JobID,
		Status: app.GetCurrentJobStatusString(),
	})
}

func getCurrentStatus(w http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})
	nodeManager := core.GetNodeManager()
	app := core.GetGlobalAppController()
	data, err := MessagesHub.Get()
	if err != nil {
		r.JSON(w, http.StatusServiceUnavailable, JSONResponse{Error: err.Error()})
	}
	messages := []Message{}
	if len(data) != 0 {
		err = json.Unmarshal(data, &messages)
		if err != nil {
			r.JSON(w, http.StatusServiceUnavailable, JSONResponse{Error: err.Error()})
			return
		}
	}
	nodeInfos := []core.NodeInfo{}
	for _, info := range nodeManager.NodeInfos {
		nodeInfos = append(nodeInfos, info)
	}
	r.JSON(w, http.StatusOK, JSONResponse{
		CurrentMessages: messages,
		ID:              app.JobID,
		Status:          app.GetCurrentJobStatusString(),
		NodeInfos:       nodeInfos,
	})
}

func updateNodeEnableStatus(w http.ResponseWriter, req *http.Request) {
	var ipinfo IPWithPort
	r := render.New(render.Options{})
	decoder := json.NewDecoder(req.Body)
	nodeManager := core.GetNodeManager()
	err := decoder.Decode(&ipinfo)
	if err != nil {
		r.JSON(w, http.StatusBadRequest, JSONResponse{Error: "decode post data fail"})
		return
	}
	log.Infof("%+v", ipinfo)
	err = nodeManager.UpdateEnableStatusAgent(ipinfo.IPAddress, ipinfo.Port, ipinfo.Enable)
	if err != nil {
		r.JSON(w, http.StatusBadRequest, JSONResponse{Error: "decode post data fail"})
		return
	}
	r.JSON(w, http.StatusOK, JSONResponse{})
}

func addNode(w http.ResponseWriter, req *http.Request) {
	var ipinfo IPWithPort
	r := render.New(render.Options{})
	decoder := json.NewDecoder(req.Body)
	nodeManager := core.GetNodeManager()
	err := decoder.Decode(&ipinfo)
	if err != nil {
		r.JSON(w, http.StatusBadRequest, JSONResponse{Error: "decode post data fail"})
		return
	}
	if err := ipinfo.Validate(); err != nil {
		r.JSON(w, http.StatusBadRequest, JSONResponse{Error: err.Error()})
		return
	}
	err = nodeManager.AddNode(ipinfo.IPAddress, ipinfo.Port)
	if err != nil {
		r.JSON(w, http.StatusBadRequest, JSONResponse{Error: "add node fail:" + err.Error()})
		return
	}
	r.JSON(w, http.StatusOK, JSONResponse{})
}

func deleteNode(w http.ResponseWriter, req *http.Request) {
	var ipinfo IPWithPort
	r := render.New(render.Options{})
	decoder := json.NewDecoder(req.Body)
	nodeManager := core.GetNodeManager()
	err := decoder.Decode(&ipinfo)
	if err != nil {
		r.JSON(w, http.StatusBadRequest, JSONResponse{Error: "decode post data fail"})
		return
	}
	if err := ipinfo.Validate(); err != nil {
		r.JSON(w, http.StatusBadRequest, JSONResponse{Error: err.Error()})
		return
	}
	err = nodeManager.RemoveNode(ipinfo.IPAddress, ipinfo.Port)
	if err != nil {
		r.JSON(w, http.StatusBadRequest, JSONResponse{Error: "delete node fail:" + err.Error()})
		return
	}
	r.JSON(w, http.StatusOK, JSONResponse{})
}

// HistoryResponse return data for history query

func getQueryHistory(w http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{})
	dbHandler := core.GetDBHandler()
	search := req.URL.Query().Get("search[value]")
	start, _ := strconv.Atoi(req.URL.Query().Get("start"))
	draw, _ := strconv.Atoi(req.URL.Query().Get("draw"))
	length, _ := strconv.Atoi(req.URL.Query().Get("length"))
	response, err := dbHandler.GetDNSQueryHistory(start, length, search)
	if err != nil {
		r.JSON(w, http.StatusOK, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}
	r.JSON(w, http.StatusOK, map[string]interface{}{
		"draw": draw,
		"data": response,
	})
}

// NewServer function create the http api
func NewServer() error {
	app := core.GetGlobalAppController()
	key := []byte(app.AppSecrect)

	store = sessions.NewCookieStore(key)
	r := mux.NewRouter()
	r.HandleFunc("/", auth(index)).Methods("GET")
	r.HandleFunc("/logout", logout).Methods("POST", "GET")
	r.HandleFunc("/history", auth(getQueryHistory)).Methods("GET")
	r.HandleFunc("/login", login(app)).Methods("GET", "POST")
	r.HandleFunc("/nodes", auth(addNode)).Methods("POST")
	r.HandleFunc("/update-node", auth(updateNodeEnableStatus)).Methods("POST")
	r.HandleFunc("/nodes", auth(deleteNode)).Methods("DELETE")
	r.HandleFunc("/start", auth(startDNSTraffic)).Methods("POST")
	r.HandleFunc("/stop", auth(stopDNSTraffic)).Methods("GET")
	r.HandleFunc("/status", auth(getCurrentStatus)).Methods("GET")
	r.PathPrefix("/public/").Handler(http.StripPrefix("/public", http.FileServer(http.Dir("./web/assets"))))
	err := http.ListenAndServe(app.AppConfig.HTTPServer, http.TimeoutHandler(r, time.Second*10, "timeout"))
	return err
}
