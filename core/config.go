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

// LoadAppConfigurationFromFile read a appConfig.ini file from local file system
// and return the AppConfig object
func LoadAppConfigurationFromFile(filename string) (*AppConfig, error) {
	appConfig := AppConfig{}
	if strings.HasSuffix(filename, ".ini") == false {
		return nil, errors.New("Configuration file must be .ini file type")
	}
	cfg, err := ini.Load(filename)
	if err != nil {
		return nil, fmt.Errorf("Configuration file load error:%s", err.Error())
	}
	appConfigSectionApp, err := cfg.GetSection("App")
	if err != nil {
		return nil, fmt.Errorf("Load app appConfiguration file section [App] error:%s", err.Error())
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

// JobConfig hold the job appConfiguration
type JobConfig struct {
	JobID              string `json:"job_id" valid:"uuid,optional"`
	Duration           string `json:"duration" valid:"-"`
	QPS                int    `json:"qps" valid:"-"`
	Server             string `json:"server" valid:"ip,optional"`
	Port               string `json:"port" valid:"port,optional"`
	Domain             string `json:"domain" valid:"-"`
	EnableEDNS         string `json:"edns_enable" valid:"-"`
	EnableDNSSEC       string `json:"dnssec_enable" valid:"-"`
	DomainRandomLength int    `json:"domain_random_length" valid:"-"`
	QueryType          string `json:"query_type" valid:"-"`
}

//NewJobConfig create a init job for appConfigration
func NewJobConfig() *JobConfig {
	return &JobConfig{
		Port:               DefaultPort,
		QPS:                DefaultQPS,
		DomainRandomLength: DefaultRandomLength,
	}
}

// ValidateJobConfiguration validate the job config
func (jobConfig *JobConfig) ValidateJobConfiguration() error {
	_, err := govalidator.ValidateStruct(jobConfig)
	if err != nil {
		return err
	}
	if jobConfig.QPS < 0 || jobConfig.DomainRandomLength < 0 {
		return errors.New("number can't set to nagetive")
	}
	if jobConfig.JobID == "" {
		id, _ := uuid.NewV4()
		jobConfig.JobID = (*id).String()
	}
	return nil
}

// Configuration define all appConfig for this app
type Configuration struct {
	sync.RWMutex
	*JobConfig
	*AppConfig
	Status   uint32
	IsMaster bool
}

var globalConfig *Configuration

// NewConfigurationFromFile load the appConfiguration and save to global variable
func NewConfigurationFromFile(file string) (*Configuration, error) {
	appConfig, err := LoadAppConfigurationFromFile(file)
	if err != nil {
		return nil, err
	}
	config := &Configuration{
		AppConfig: appConfig,
		JobConfig: NewJobConfig(),
		IsMaster:  true,
		Status:    StatusStopped,
	}
	if err = config.ValidateJobConfiguration(); err != nil {
		return nil, err
	}
	globalConfig = config
	return config, nil
}

// GetGlobalConfig return current appConfiguration
func GetGlobalConfig() *Configuration {
	if globalConfig == nil {
		globalConfig = &Configuration{
			JobConfig: NewJobConfig(),
			IsMaster:  false,
		}
	}
	return globalConfig
}

// GetCurrentJobStatus return current task running status
func (config *Configuration) GetCurrentJobStatus() uint32 {
	config.RLock()
	defer config.RUnlock()
	return config.Status
}

// GetCurrentJobStatusString return the readable string
func (config *Configuration) GetCurrentJobStatusString() string {
	code := config.GetCurrentJobStatus()
	codeString, _ := StatusToString[code]
	return codeString
}

// SetCurrentJobStatus change the status of current task
func (config *Configuration) SetCurrentJobStatus(status uint32) error {
	config.Lock()
	defer config.Unlock()
	config.Status = status
	return nil
}
