package core

import (
	"testing"
)

func TestGetDetailInfo(t *testing.T) {
	var detail string
	var testCase = []struct {
		code   ReturnCode
		detail string
	}{{
		code:   RET_FORMERR,
		detail: "Query Format Error",
	}, {
		code:   RET_NO_ERROR,
		detail: "Success",
	}, {
		code:   RET_NOTIMP,
		detail: "Query Type Not Support",
	}, {
		code:   RET_NXDOMAIN,
		detail: "Query Not Exist",
	}, {
		code:   RET_REFUSED,
		detail: "Query Refused",
	}, {
		code:   RET_SERVFAIL,
		detail: "Server Fail",
	}, {
		code:   100001,
		detail: "Unknown",
	}}
	for _, obj := range testCase {
		detail = GetDetailInfo(obj.code)
		if detail != obj.detail {
			t.Errorf("Expect %s: Got %s", obj.detail, detail)
		}
	}
}
