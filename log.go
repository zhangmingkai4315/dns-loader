package main

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/zhangmingkai4315/dns-loader/web"
)

//MyHook 定义logrus的hook类型
type MyHook struct{}

var fmter = new(log.TextFormatter)

// Levels 必须实施的接口类型，返回所有的打印级别信息
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

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	// 将所有日志同时写入messagehub(传递到前端的console)
	log.SetOutput(web.MessagesHub)
	log.AddHook(&MyHook{})
}
