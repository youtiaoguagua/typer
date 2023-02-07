package main

import (
	"flag"
	"fmt"
	_ "github.com/youtiaoguagua/typer/log"
	"github.com/youtiaoguagua/typer/server"
	"github.com/youtiaoguagua/typer/util"
	"net/http"
	_ "net/http/pprof"
)

var commandInfo = server.CommandInfo{}

func init() {
	// init command info
	h := util.GetEnv("SERVER_HOST", "0.0.0.0")
	p := util.GetEnvInt("SERVER_PORT", 7788)

	path := util.GetEnv("SERVER_KEY_PATH", "data/typer_key")

	flag.StringVar(&commandInfo.Host, "h", h, "server host")
	flag.IntVar(&commandInfo.Port, "p", p, "server port")
	flag.StringVar(&commandInfo.ServerKeyPath, "path", path, "server key path")
}

func main() {
	go func() {
		ip := "0.0.0.0:6060"
		if err := http.ListenAndServe(ip, nil); err != nil {
			fmt.Printf("start pprof failed on %s\n", ip)
		}
	}()

	flag.Parse()
	// 开启服务
	server.Start(commandInfo)
}
