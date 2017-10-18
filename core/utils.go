package core

func GetDetailInfo(code ReturnCode) string {
	var detail string
	switch code {
	case RET_NO_ERROR:
		detail = "Success"
	case RET_FORMERR:
		detail = "Query Format Error"
	case RET_NOTIMP:
		detail = "Query Type Not Support"
	case RET_NXDOMAIN:
		detail = "Query Not Exist"
	case RET_REFUSED:
		detail = "Query Refused"
	case RET_SERVFAIL:
		detail = "Server Fail"
	default:
		detail = "Unknown"
	}
	return detail
}
