package core

// LoadManager define the behavior of dns benchmark loader
type LoadManager interface {
	Start() bool
	Stop() bool
	Status() uint32
	CallCount() uint64
}

// LoadCaller define the behavior of call processor
type LoadCaller interface {
	BuildReq(job *JobConfig) []byte
	Call(req []byte)
}
