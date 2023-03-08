package main

import (
	"fmt"
	"go-redis/config"
	"go-redis/lib/logger"
	"go-redis/resp/handler"
	"go-redis/tcp"
	"os"
)

const configFile string = "redis.conf"

var defaultProperties = &config.ServerProperties{
	Bind: "0.0.0.0",
	Port: 6379,
}

// *3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n
// %2a3%0d%0a$3%0d%0aSET%0d%0a$3%0d%0akey%0d%0a$5%0d%0avalue%0d%0a
// *2\r\n$3\r\nGET\r\n$3\r\nkey\r\n
// %2a2%0d%0a$3%0d%0aGET%0d%0a$3%0d%0akey%0d%0a
// $3\r\nset\r\n
// %24%0d%0aset%0d%0a
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	return err == nil && !info.IsDir()
}

// tcp处理socket -> 交给handler监听连接 -> 将连接交给ParseStream -> 解析的指令放入管道
// -> 交给Database的Exec执行
func main() {
	logger.Setup(&logger.Settings{
		Path:       "logs",
		Name:       "godis",
		Ext:        "log",
		TimeFormat: "2006-01-02",
	})

	if fileExists(configFile) {
		config.SetupConfig(configFile)
	} else {
		config.Properties = defaultProperties
	}

	err := tcp.ListenAndServeWithSignal(
		&tcp.Config{Address: fmt.Sprintf("%s:%d", config.Properties.Bind, config.Properties.Port)},
		handler.MakeHandler())
	if err != nil {
		logger.Error(err)
	}
}
