package core

import (
	"errors"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/jinzhu/gorm"
	// for sqlite import
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var dbHander *DBHandler

//DBHandler db manager
type DBHandler struct {
	*gorm.DB
}

// Agent save all agent information in sqlite table
type Agent struct {
	gorm.Model
	IP   string
	Port string
	Live bool
}

// DNSQuery save all query history
type DNSQuery struct {
	gorm.Model
	Server    string `json:"server"`
	Port      string `json:"port"`
	Duration  string `json:"duration"`
	QPS       int    `json:"qps"`
	Domain    string `json:"domain"`
	Length    int    `json:"domain_random_length"`
	QueryType string `json:"query_type"`
}

// NewDatabaseFromFile create database from file
func NewDatabaseFromFile(dbfile string) error {
	if _, err := os.Stat(dbfile); os.IsNotExist(err) {
		dbfile, err := os.Create(dbfile)
		if err != nil {
			return fmt.Errorf("create database file error: %s", err.Error())
		}
		dbfile.Close()
	}
	var err error
	db, err := gorm.Open("sqlite3", dbfile)
	if err != nil {
		log.Fatalf("open dbfile error: %s", err.Error())
	}
	db.AutoMigrate(&Agent{}, &DNSQuery{})
	dbHander = &DBHandler{
		DB: db,
	}
	return nil
}

// GetDBHandler return global db manager
func GetDBHandler() *DBHandler {
	return dbHander
}

// AddAgent dynamic add a new agent in the database
func (dbHander DBHandler) AddAgent(ip string, port string) error {
	agent := Agent{
		IP:   ip,
		Port: port,
	}
	err := dbHander.First(&agent).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			// insert a new agent
			dbHander.Save(&agent)
			return nil
		}
		return fmt.Errorf("save new agent fail: %s", err)
	}
	return errors.New("already exist")
}

// RemoveAgent remove a agent ip from database
func (dbHander DBHandler) RemoveAgent(ip string, port string) error {
	agent := Agent{
		IP:   ip,
		Port: port,
	}
	err := dbHander.Unscoped().Delete(&agent).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return errors.New("agent not in database")
		}
		return fmt.Errorf("delete agent fail: %s", err)
	}
	return nil
}

// CreateDNSQuery save a new dns query info
func (dbHander *DBHandler) CreateDNSQuery(config *Configuration) error {
	dnsQuery := DNSQuery{
		Server:    config.Server,
		Port:      config.Port,
		Duration:  config.Duration.String(),
		QPS:       config.QPS,
		Domain:    config.Domain,
		Length:    config.DomainRandomLength,
		QueryType: config.QueryType,
	}
	return dbHander.Model(&DNSQuery{}).Save(&dnsQuery).Error
}

// GetHistoryInfo return DNSQuery for datatables
func (dbHander *DBHandler) GetHistoryInfo(start, length int, search string) ([]DNSQuery, error) {
	data := []DNSQuery{}
	err := dbHander.Order("id desc").Limit(length).Offset(start).Where("server LIKE ? or domain Like ?", "%"+search+"%", "%"+search+"%").Find(&data).Error
	if err != nil && gorm.IsRecordNotFoundError(err) == false {
		return data, err
	}
	return data, nil
}
