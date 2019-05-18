package web

import (
	"encoding/json"
	"errors"
	"sync"
)

// Message define one single event message
type Message struct {
	Level  string `json:"level"`
	Time   string `json:"time"`
	Msg    string `json:"msg"`
	Result bool   `json:"result"`
}

// GlobalMessages define the Messages Manager
type GlobalMessages struct {
	locker  sync.RWMutex
	MaxSize int
	Events  []Message `json:"events"`
}

// MessagesHub for global message hub
var MessagesHub *GlobalMessages

func init() {
	MessagesHub = NewGloabalMessages(50)
	SetupLogger()
}

// Len return current message length
func (g *GlobalMessages) Len() int {
	g.locker.RLock()
	defer g.locker.RUnlock()
	return len(g.Events)
}

func (g *GlobalMessages) Write(p []byte) (n int, err error) {
	g.locker.Lock()
	defer g.locker.Unlock()
	event := Message{}
	if err := json.Unmarshal(p, &event); err != nil {
		return 0, err
	}
	if event.Msg == "" {
		return 0, errors.New("Empty Messages")
	}
	g.Events = append(g.Events, event)
	return len(p), nil
}

// Get get current message and delete it from messages manager
func (g *GlobalMessages) Get() ([]byte, error) {
	g.locker.Lock()
	defer g.locker.Unlock()
	if len(g.Events) == 0 {
		return []byte{}, nil
	}
	result, err := json.Marshal(g.Events)
	if err != nil {
		return nil, err
	}
	g.Events = nil
	return result, nil
}

// NewGloabalMessages create a new messages manager
func NewGloabalMessages(max int) *GlobalMessages {
	events := make([]Message, 0)
	return &GlobalMessages{
		MaxSize: max,
		Events:  events,
	}
}
