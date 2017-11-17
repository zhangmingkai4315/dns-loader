package web

import (
	"encoding/json"
	"errors"
	"sync"
)

type Message struct {
	Level  string `json:"level"`
	Time   string `json:"time"`
	Msg    string `json:"msg"`
	Result bool   `json:"result"`
}
type GlobalMessages struct {
	locker  sync.RWMutex
	MaxSize int
	Events  []Message `json:"events"`
}

var MessagesHub *GlobalMessages

func init() {
	// events := make([]Message, 0)
	MessagesHub = NewGloabalMessages(50)
}

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

func (g *GlobalMessages) Get() ([]byte, error) {
	g.locker.Lock()
	defer g.locker.Unlock()
	result, err := json.Marshal(g.Events)
	if err != nil {
		return nil, err
	}
	g.Events = nil
	return result, nil
}

func NewGloabalMessages(max int) *GlobalMessages {
	events := make([]Message, 0)
	return &GlobalMessages{
		MaxSize: 100,
		Events:  events,
	}
}
