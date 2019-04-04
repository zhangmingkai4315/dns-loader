package core

import (
	"fmt"
	"math/rand"
	"os"
	"time"
)

var letters = []byte("1234567890abcdefghijklmnopqrstuvwxyz")

func init() {
	rand.Seed(time.Now().UnixNano())
}

func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
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
	b := make([]byte, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b) + "." + domain
}

// LoadConfigFile func
func LoadConfigFile(file string) error {
	return nil
}
