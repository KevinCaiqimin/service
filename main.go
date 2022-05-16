package main

import (
	"context"
	"flag"
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

	log.InitLog(*log_filename, "HOUR", log.LV_DEBUG)

	ctx, cancel_func := context.WithCancel(context.Background())

	switch *app {
	case "client":
		for i := 0; i < *cli_num; i++ {
			cli := client.NewClient(*addr)
			cli.Start(ctx)
		}
		break
	case "service":
		srv := service.NewService(*addr)
		srv.Start(ctx)
		break
	default:
		flag.Usage()
		return
	}

	handler := http.NewServeMux()
	handler.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	srv := &http.Server{Addr: "0.0.0.0:8888", Handler: handler}
	go srv.ListenAndServe()

	exit_ch := make(chan os.Signal, 1)
	signal.Notify(exit_ch, os.Interrupt)
	<-exit_ch
	cancel_func()
}
