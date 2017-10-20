package dnsloader

import (
	"strings"
)

// GetDNSTypeFromString return the true code of dns type
func GetDNSTypeFromString(typeString string) int {
	queryString := "dnsType" + strings.ToUpper(typeString)
	code, ok := DNSType[queryString]
	if ok {
		return code
	} else {
		return 0
	}
}
