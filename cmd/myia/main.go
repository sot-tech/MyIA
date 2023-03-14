// Package main starts http.Server with MyIA handler
package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sot-te.ch/myia"
)

var errNoAddress = errors.New("no listen address provided")

const requestTimeout = time.Millisecond * 500

func main() {
	var addr, path, acao, net, header string
	flag.StringVar(&addr, "l", "127.0.0.1:1234", "Listen address")
	flag.StringVar(&path, "p", "/", "Listen path")
	flag.StringVar(&net, "n", "", "Filter retrieved IP with provided network")
	flag.StringVar(&acao, "o", "", "Set provided value to 'Access-Control-Allow-Origin' header")
	flag.StringVar(&header, "r", "", "Get value as client IP from provided HTTP header instead of HTTP remote address")
	flag.Parse()

	if len(addr) == 0 {
		fmt.Println(errNoAddress)
		os.Exit(1)
	}

	h, err := myia.NewHandler(path, acao, net, header)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	srv := &http.Server{
		Addr:         addr,
		ReadTimeout:  requestTimeout,
		WriteTimeout: requestTimeout,
		Handler:      h,
	}
	defer func(srv *http.Server) {
		err := srv.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(srv)
	ch := make(chan os.Signal, 2)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	fmt.Println("Staring server", srv.Addr)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Println(err)
			ch <- syscall.SIGABRT
		}
	}()
	sig := <-ch
	if sig == syscall.SIGABRT {
		os.Exit(1)
	}
}
