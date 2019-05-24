package core

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
	"time"
)

var letters = []byte("1234567890abcdefghijklmnopqrstuvwxyz")

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Assert fails the test if the condition is false.
func Assert(t *testing.T, condition bool, msg string, v ...interface{}) {
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
		t.FailNow()
	}
}

// OK fails the test if an err is not nil.
func OK(t *testing.T, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		t.FailNow()
	}
}

// Equals fails the test if exp is not equal to act.
func Equals(t *testing.T, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
		t.FailNow()
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
	if length == 0 {
		return domain
	}
	b := make([]byte, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b) + "." + domain
}
