package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/lyuangg/umyproxy/internal/proxy"
)

var (
	appname     string = "umyproxy"
	version     string = "0.0.1"
	showversion bool
	host        string
	port        int
	socketfile  string
	poolsize    int
	maxlife     int
	waittimeout int
	debug       bool
)

const (
	logstr = `
   __  ____  ___      ____
  / / / /  |/  /_  __/ __ \_________  _  ____  __
 / / / / /|_/ / / / / /_/ / ___/ __ \| |/_/ / / /
/ /_/ / /  / / /_/ / ____/ /  / /_/ />  </ /_/ /
\____/_/  /_/\__, /_/   /_/   \____/_/|_|\__, /
            /____/                      /____/
    `
)

func init() {
	flag.BoolVar(&showversion, "version", false, "show version")
	flag.StringVar(&host, "host", "127.0.0.1", "mysql host")
	flag.IntVar(&port, "port", 3306, "mysql port")
	flag.StringVar(&socketfile, "socket", "/tmp/"+appname+".socket", "socket file path")
	flag.IntVar(&poolsize, "size", runtime.NumCPU()*10, "pool size")
	flag.IntVar(&maxlife, "life", 3600, "mysql connection max life time")
	flag.IntVar(&waittimeout, "wait", 3000, "wait mysql connection timeout")
	flag.BoolVar(&debug, "debug", false, "set debug mode")
}

func main() {
	flag.Parse()

	if showversion {
		fmt.Println(appname, version)
		os.Exit(0)
	}
	fmt.Println(appname, version)
	fmt.Println(logstr)

	option := proxy.PoolOption{
		Host:        host,
		Port:        port,
		MaxLifetime: time.Second * time.Duration(maxlife),
		PoolMaxSize: poolsize,
		WaitTimeout: time.Millisecond * time.Duration(waittimeout),
	}

	p := proxy.NewProxy(proxy.NewPool(option), socketfile)
	if debug {
		p.SetDebug()
	}

	processed := make(chan struct{})
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGUSR2, syscall.SIGINT, syscall.SIGTERM)
		<-c
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := p.Shutdown(ctx); nil != err {
			log.Fatalf("shutdown failed, err: %v\n", err)
		}
		close(processed)
	}()

	p.Run()

	<-processed
	log.Println("exit")
}
