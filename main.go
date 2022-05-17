package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"

	"github.com/KevinCaiqimin/log"
	"github.com/KevinCaiqimin/service/client"
	"github.com/KevinCaiqimin/service/service"
)

func main() {
	app := flag.String("app", "service", "[required] start app name")
	log_filename := flag.String("log", "debug.log", "[optional] log file name")
	addr := flag.String("addr", ":8888", "[optional] log file name")
	cli_num := flag.Int("cli_num", 10, "[optional] client number")

	flag.Parse()

	ctx, cancel_func := context.WithCancel(context.Background())

	// login_mod := &login.LoginMod{}
	// router.Reg(login_mod)
	// m := make(map[string]interface{})
	// m["acc"] = "123"
	// m["pwd"] = "123"
	// js_body, _ := json.Marshal(m)
	// router.Call("Login", string(js_body))

	pprof_port := 8889

	switch *app {
	case "client":
		log.InitLog(*log_filename, "HOUR", log.LV_ERROR)
		pprof_port = 8890
		for i := 0; i < *cli_num; i++ {
			cli := client.NewClient(*addr)
			cli.Start(ctx)
		}
		break
	case "service":
		log.InitLog(*log_filename, "HOUR", log.LV_ERROR)
		srv := service.NewService(*addr)
		srv.Start(ctx)
		break
	default:
		flag.Usage()
		return
	}

	pprof_addr := fmt.Sprintf("0.0.0.0:%d", pprof_port)
	handler := http.NewServeMux()
	handler.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	srv := &http.Server{Addr: pprof_addr, Handler: handler}
	go srv.ListenAndServe()

	exit_ch := make(chan os.Signal, 1)
	signal.Notify(exit_ch, os.Interrupt)
	<-exit_ch
	cancel_func()
}
