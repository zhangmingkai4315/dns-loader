package dnsloader

import (
	"testing"
)

func TestGetDNSTypeCodeFromString(t *testing.T) {
	var testCase = []struct {
		input  string
		expect int
	}{{
		input:  "a",
		expect: 1,
	}, {
		input:  "A",
		expect: 1,
	}, {
		input:  "AAAA",
		expect: 28,
	}, {
		input:  "aaaa",
		expect: 28,
	}, {
		input:  "not-exist",
		expect: 0,
	},
	}
	for _, obj := range testCase {
		output := GetDNSTypeCodeFromString(obj.input)
		if output != obj.expect {
			t.Errorf("Expect %d: Got %d", obj.expect, output)
		}
	}
}
