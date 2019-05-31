package core

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	uuid "github.com/nu7hatch/gouuid"

	"github.com/asaskevich/govalidator"
	"gopkg.in/ini.v1"
)

// Default value for benchmark
const (
	DefaultPort         = "53"
	DefaultRandomLength = 0
	DefaultQPS          = 100
	DefaultMaxQuery     = 0
	DefaultProtocol     = "udp"
)

func init() {
	govalidator.SetFieldsRequiredByDefault(true)
}

// AppConfig hold the configration from ini file
type AppConfig struct {
	User              string
	Password          string
	AppSecrect        string
	HTTPServer        string
	ConfigFileName    string
	ConfigFileHandler *ini.File
}

// NewAppConfigFromFile read a appConfig.ini file from local file system
// and return the AppConfig object
func NewAppConfigFromFile(filename string) (*AppConfig, error) {
	appConfig := AppConfig{}
	if strings.HasSuffix(filename, ".ini") == false {
		return nil, errors.New("AppController file must be .ini file type")
	}
	cfg, err := ini.Load(filename)
	if err != nil {
		return nil, fmt.Errorf("AppController file load error:%s", err.Error())
	}
	appConfigSectionApp, err := cfg.GetSection("App")
	if err != nil {
		return nil, fmt.Errorf("Load app appAppController file section [App] error:%s", err.Error())
	}
	if appConfigSectionApp.HasKey("user") {
		appConfig.User = appConfigSectionApp.Key("user").String()
	}
	if appConfigSectionApp.HasKey("password") {
		appConfig.Password = appConfigSectionApp.Key("password").String()
	}
	if appConfigSectionApp.HasKey("app_secrect") {
		appConfig.AppSecrect = appConfigSectionApp.Key("app_secrect").String()
	}
	if appConfigSectionApp.HasKey("http_server") {
		appConfig.HTTPServer = appConfigSectionApp.Key("http_server").String()
	}
	return &appConfig, nil
}

// JobConfig hold the job appAppController
type JobConfig struct {
	JobID              string `json:"job_id" valid:"uuid,optional"`
	Duration           string `json:"duration" valid:"-"`
	Protocol           string `json:"protocol" valid:"in(tcp|udp),optional"`
	QPS                uint32 `json:"qps" valid:"-"`
	ClientNumber       int    `json:"client_number" valid:"-"`
	MaxQuery           uint64 `json:"max_query" valid:"-"`
	Server             string `json:"server" valid:"ip,optional"`
	Port               string `json:"port" valid:"port,optional"`
	Domain             string `json:"domain" valid:"-"`
	EnableEDNS         string `json:"edns_enable" valid:"-"`
	EnableDNSSEC       string `json:"dnssec_enable" valid:"-"`
	DomainRandomLength int    `json:"domain_random_length" valid:"-"`
	QueryType          string `json:"query_type" valid:"-"`
}

//NewDefaultJobConfig create a init job for appConfigration
func NewDefaultJobConfig() *JobConfig {
	return &JobConfig{
		Port:               DefaultPort,
		QPS:                DefaultQPS,
		DomainRandomLength: DefaultRandomLength,
		MaxQuery:           DefaultMaxQuery,
		Protocol:           DefaultProtocol,
	}
}

// ValidateJob validate the job config
func (jobConfig *JobConfig) ValidateJob() error {
	_, err := govalidator.ValidateStruct(jobConfig)
	if err != nil {
		return err
	}
	if jobConfig.QPS < 0 {
		return errors.New("qps number can't set to nagetive")
	}
	if jobConfig.DomainRandomLength < 0 {
		return errors.New("domain random length can't set to nagetive")
	}
	if jobConfig.MaxQuery < 0 {
		return errors.New("maximum number of queries can't set to nagetive")
	}
	if jobConfig.JobID == "" {
		id, _ := uuid.NewV4()
		jobConfig.JobID = (*id).String()
	}
	return nil
}

// AppController hold all infomation and control interface for this app
type AppController struct {
	sync.RWMutex
	*JobConfig
	*AppConfig
	LoadManager
	LoadCaller
	Status   uint32
	IsMaster bool
}

var appController *AppController

// NewAppControllerFromFile load the AppController and save to global variable
func NewAppControllerFromFile(file string) (*AppController, error) {
	appConfig, err := NewAppConfigFromFile(file)
	if err != nil {
		return nil, err
	}
	controller := &AppController{
		AppConfig: appConfig,
		JobConfig: NewDefaultJobConfig(),
		IsMaster:  true,
		Status:    StatusStopped,
	}
	if err = controller.ValidateJob(); err != nil {
		return nil, err
	}
	appController = controller
	return controller, nil
}

// GetGlobalAppController return current appAppController
func GetGlobalAppController() *AppController {
	if appController == nil {
		appController = &AppController{
			JobConfig: NewDefaultJobConfig(),
			IsMaster:  false,
			Status:    StatusStopped,
		}
	}
	return appController
}

// GetCurrentJobStatus return current task running status
func (config *AppController) GetCurrentJobStatus() uint32 {
	config.RLock()
	defer config.RUnlock()
	return config.Status
}

// GetCurrentJobStatusString return the readable string
func (config *AppController) GetCurrentJobStatusString() string {
	code := config.GetCurrentJobStatus()
	codeString, _ := StatusToString[code]
	return codeString
}

// SetCurrentJobStatus change the status of current task
func (config *AppController) SetCurrentJobStatus(status uint32) error {
	config.Lock()
	defer config.Unlock()
	config.Status = status
	return nil
}
