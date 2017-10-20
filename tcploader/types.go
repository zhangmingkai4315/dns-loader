package tcploader

type ServerReq struct {
	ID       int64
	Operands []int
	Operator string
}
type ServerResp struct {
	ID     int64
	Result int
	Err    error
}
