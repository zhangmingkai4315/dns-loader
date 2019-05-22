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

func init() {
	govalidator.SetFieldsRequiredByDefault(true)
}

// Configuration define all config for this app
type Configuration struct {
	sync.RWMutex       `json:"-" valid:"-"`
	IsMaster           bool           `json:"-" valid:"-"`
	Master             string         `json:"master" valid:"ip,optional"`
	Duration           CustomDuration `json:"duration" valid:"-"`
	QPS                int            `json:"qps" valid:"-"`
	Server             string         `json:"server" valid:"ip"`
	Port               string         `json:"port" valid:"port"`
	Domain             string         `json:"domain" valid:"-"`
	EnableEDNS         string         `json:"edns_enable" valid:"-"`
	EnableDNSSEC       string         `json:"dnssec_enable" valid:"-"`
	DomainRandomLength int            `json:"domain_random_length" valid:"-"`
	QueryType          string         `json:"query_type" valid:"-"`
	HTTPServer         string         `json:"web" valid:"-"`
	AgentPort          string         `json:"agent_port" valid:"port,optional"`
	User               string         `json:"-" valid:"-"`
	Password           string         `json:"-" valid:"-"`
	AppSecrect         string         `json:"-" valid:"-"`
	// For current job information
	ID     string `json:"id" valid:"uuid,optional"`
	Status uint32 `json:"-" valid:"-"`

	configFileName    string
	configFileHandler *ini.File
}

var globalConfig *Configuration

// NewConfigurationFromFile load the configuration and save to global variable
func NewConfigurationFromFile(file string) (*Configuration, error) {

	config := &Configuration{
		configFileName: file,
		IsMaster:       true,
		Status:         StatusStopped,
	}
	err := config.LoadConfigurationFromIniFile(file)
	if err != nil {
		return nil, err
	}
	if err = config.ValidateConfiguration(); err != nil {
		return nil, err
	}
	globalConfig = config
	return config, nil
}

// GetGlobalConfig return current configuration
func GetGlobalConfig() *Configuration {
	if globalConfig == nil {
		globalConfig = &Configuration{
			IsMaster: false,
			Status:   StatusStopped,
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

// ValidateConfiguration will check all setting and if no error accure return nil
func (config *Configuration) ValidateConfiguration() error {
	_, err := govalidator.ValidateStruct(config)
	if err != nil {
		return err
	}
	if config.QPS <= 0 || config.DomainRandomLength < 0 {
		return errors.New("number can't set to nagetive")
	}
	if config.ID == "" {
		id, _ := uuid.NewV4()
		config.ID = (*id).String()
	}
	return nil
}

// LoadConfigurationFromIniFile read a config.ini file from local file system
// and return the configuration object
func (config *Configuration) LoadConfigurationFromIniFile(filename string) (err error) {
	if strings.HasSuffix(filename, ".ini") == false {
		return errors.New("Configuration file must be .ini file type")
	}
	cfg, err := ini.Load(filename)
	if err != nil {
		return fmt.Errorf("Configuration file load error:%s", err.Error())
	}
	config.configFileHandler = cfg
	configSectionApp, err := cfg.GetSection("App")
	if err != nil {
		return fmt.Errorf("Configuration file load section [App] error:%s", err.Error())
	}
	if configSectionApp.HasKey("user") {
		config.User = configSectionApp.Key("user").String()
	}
	if configSectionApp.HasKey("password") {
		config.Password = configSectionApp.Key("password").String()
	}
	if configSectionApp.HasKey("app_secrect") {
		config.AppSecrect = configSectionApp.Key("app_secrect").String()
	}
	if configSectionApp.HasKey("control_master") {
		config.Master = configSectionApp.Key("master").String()
	}
	if configSectionApp.HasKey("agent_port") {
		config.AgentPort = configSectionApp.Key("agent_port").String()
	}
	if configSectionApp.HasKey("http_server") {
		config.HTTPServer = configSectionApp.Key("http_server").String()
	}
	configSectionQuery, err := cfg.GetSection("Query")
	if err != nil {
		return fmt.Errorf("Configuration file load section [Query] error:%s", err.Error())
	}
	if configSectionQuery.HasKey("duration") {
		duration, err := configSectionQuery.Key("duration").Duration()
		if err != nil {
			return fmt.Errorf("Configuration file load section [Query] error:%s", err.Error())
		}
		config.Duration = CustomDuration{
			Duration: duration,
		}
	}
	if configSectionQuery.HasKey("qps") {
		config.QPS = configSectionQuery.Key("qps").MustInt()
	}
	if configSectionQuery.HasKey("server") {
		config.Server = configSectionQuery.Key("server").String()
	}
	if configSectionQuery.HasKey("port") {
		config.Port = configSectionQuery.Key("port").String()
	}
	if configSectionQuery.HasKey("domain") {
		config.Domain = configSectionQuery.Key("domain").String()
	}
	if configSectionQuery.HasKey("randomlen") {
		config.DomainRandomLength = configSectionQuery.Key("server").MustInt()
	}
	if configSectionQuery.HasKey("query_type") {
		config.QueryType = configSectionQuery.Key("query_type").String()
	}
	return
}
