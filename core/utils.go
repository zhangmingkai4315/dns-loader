package core

// GetDetailInfo
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
	case RET_CALL_ERROR:
		detail = "Call Func Error"
	case RET_RESULT_ERROR:
		detail = "Result Error"
	case RET_CODE_ID_ERROR:
		detail = "ID Not Match"
	case RET_SERVER_ERROR:
		detail = "Server Error"
	default:
		detail = "Unknown"
	}
	return detail
}
