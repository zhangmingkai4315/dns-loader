package core

import (
	"strings"
	"testing"
)

func TestStringInSlice(t *testing.T) {
	testCases := []struct {
		inputArray []string
		inputStr   string
		expect     bool
	}{
		{
			inputArray: []string{"dns", "query", "domain"},
			inputStr:   "dns",
			expect:     true,
		},
		{
			inputArray: []string{"dns", "query", "domain"},
			inputStr:   "rfc",
			expect:     false,
		},
		{
			inputArray: []string{"dns", "query", "domain"},
			inputStr:   "",
			expect:     false,
		},
		{
			inputArray: []string{},
			inputStr:   "dns",
			expect:     false,
		},
		{
			inputArray: []string{},
			inputStr:   "",
			expect:     false,
		},
	}
	for _, test := range testCases {
		Equals(t, test.expect, StringInSlice(test.inputStr, test.inputArray))
	}
}

func TestRemoveStringInSlice(t *testing.T) {
	testCases := []struct {
		inputArray []string
		inputStr   string
		expect     []string
	}{
		{
			inputArray: []string{"dns", "query", "domain"},
			inputStr:   "dns",
			expect:     []string{"query", "domain"},
		},
		{
			inputArray: []string{"dns", "query", "domain"},
			inputStr:   "rfc",
			expect:     []string{"dns", "query", "domain"},
		},
		{
			inputArray: []string{"dns", "query", "domain"},
			inputStr:   "",
			expect:     []string{"dns", "query", "domain"},
		},
		{
			inputArray: []string{},
			inputStr:   "dns",
			expect:     []string{},
		},
		{
			inputArray: []string{},
			inputStr:   "",
			expect:     []string{},
		},
	}
	for _, test := range testCases {
		Equals(t, test.expect, RemoveStringInSlice(test.inputStr, test.inputArray))
	}
}

func TestGenRandomDomain(t *testing.T) {
	testCases := []struct {
		inputLength   int
		inputDomain   string
		expectLength  int
		expectEndWith string
	}{
		{
			inputLength:   5,
			inputDomain:   "cn",
			expectLength:  8,
			expectEndWith: ".cn",
		},
		{
			inputLength:   0,
			inputDomain:   "cn",
			expectLength:  2,
			expectEndWith: "cn",
		},
	}
	for _, test := range testCases {
		domain := GenRandomDomain(test.inputLength, test.inputDomain)
		Equals(t, test.expectLength, len(domain))
		Assert(t, strings.HasSuffix(domain, test.expectEndWith), "test suffix fail")
	}
}
