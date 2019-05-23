package web

import (
	"fmt"

	"github.com/asaskevich/govalidator"
)

func init() {
	govalidator.SetFieldsRequiredByDefault(true)
}

// IPWithPort define the posted node info
type IPWithPort struct {
	IPAddress string `json:"ipaddress" valid:"ip"`
	Port      string `json:"port" valid:"port"`
	Enable    bool   `json:"enable" valid:"-"`
}

// NodeInfo for status check response

// Validate if ip and port infomation is valid
func (ipp *IPWithPort) Validate() error {
	_, err := govalidator.ValidateStruct(ipp)
	if err != nil {
		return err
	}
	return nil
}
func (ipp *IPWithPort) toString(defaultPort string) string {
	if ipp.Port == "" {
		ipp.Port = defaultPort
	}
	return fmt.Sprintf("%s:%s", ipp.IPAddress, ipp.Port)
}
