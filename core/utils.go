package core

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

var letters = []byte("1234567890abcdefghijklmnopqrstuvwxyz")

func init() {
	rand.Seed(time.Now().UnixNano())
}

// StringInSlice find if string type object a in a string list
func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// RemoveStringInSlice find if string type object and remove it
func RemoveStringInSlice(a string, list []string) []string {
	for i, b := range list {
		if b == a {
			list = append(list[:i], list[i+1:]...)
			break
		}
	}
	return list
}

//GenRandomDomain will generate the random domain name with the fix length
func GenRandomDomain(length int, domain string) string {
	if length == 0 {
		return domain
	}
	b := make([]byte, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b) + "." + domain
}

// CustomDuration define a custom time duration for easy json serial
type CustomDuration struct {
	time.Duration
}

// UnmarshalJSON for json unmarshal
func (cd *CustomDuration) UnmarshalJSON(b []byte) (err error) {
	cd.Duration, err = time.ParseDuration(strings.Trim(string(b), `"`))
	return
}

// MarshalJSON for json marshal
func (cd *CustomDuration) MarshalJSON() (b []byte, err error) {
	return []byte(fmt.Sprintf(`"%s"`, cd.String())), nil
}
