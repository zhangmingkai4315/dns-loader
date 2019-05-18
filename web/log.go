package web

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

//MyHook 定义logrus的hook类型
type MyHook struct{}

var fmter = new(log.TextFormatter)

// Levels return all log level for Hooks
func (h *MyHook) Levels() []log.Level {
	return []log.Level{
		log.InfoLevel,
		log.WarnLevel,
		log.ErrorLevel,
		log.FatalLevel,
		log.PanicLevel,
	}
}

// Fire 必须实施的接口类型，将打印的信息进行格式化，此处使用text格式化
func (h *MyHook) Fire(entry *log.Entry) (err error) {
	line, err := fmter.Format(entry)
	if err == nil {
		fmt.Fprintf(os.Stderr, string(line))
	}
	return
}

func SetupLogger() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(MessagesHub)
	log.AddHook(&MyHook{})
}
