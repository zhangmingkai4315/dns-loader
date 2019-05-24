package core

import (
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
	UUID   string `json:"uuid"`
	IP     string `json:"ip"`
	Port   string `json:"port"`
	Live   bool   `json:"live"`
	Enable bool   `json:"enable"`
}

//IPAddrWithPort return connect ip and port combination
func (agent Agent) IPAddrWithPort() string {
	return agent.IP + ":" + agent.Port
}

// DNSQuery save all query history
type DNSQuery struct {
	gorm.Model
	JobConfig
}

// NewDatabaseConnectionFromFile create database from file
func NewDatabaseConnectionFromFile(dbfile string) error {
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

// CreateDNSQueryHistory save a new dns query info
func (dbHander *DBHandler) CreateDNSQueryHistory(appController *AppController) error {
	jobConfig := appController.JobConfig
	dnsQuery := DNSQuery{
		JobConfig: *jobConfig,
	}
	return dbHander.Model(&DNSQuery{}).Save(&dnsQuery).Error
}

// GetDNSQueryHistory return DNSQuery for datatables
func (dbHander *DBHandler) GetDNSQueryHistory(start, length int, search string) ([]DNSQuery, error) {
	data := []DNSQuery{}
	err := dbHander.Order("id desc").Limit(length).Offset(start).Where("server LIKE ? or domain Like ?", "%"+search+"%", "%"+search+"%").Find(&data).Error
	if err != nil && gorm.IsRecordNotFoundError(err) == false {
		return data, err
	}
	return data, nil
}
