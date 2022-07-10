package main

import (
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sot-te.ch/myia"
	"syscall"
	"time"
)

var errNoAddress = errors.New("no listen address provided")

const requestTimeout = time.Millisecond * 500

func main() {
	var addr, path, acao, net string
	flag.StringVar(&addr, "l", "127.0.0.1:1234", "Listen address")
	flag.StringVar(&path, "p", "/", "Listen path")
	flag.StringVar(&net, "n", "", "Filter retrieved IP with provided network (i.e. return only address from 10.0.0.0/8)")
	flag.StringVar(&acao, "o", "*", "Set provided value to `Access-Control-Allow-Origin` header")
	flag.Parse()

	if len(addr) == 0 {
		log.Fatal(errNoAddress)
	}

	h, err := myia.NewHandler(path, acao, net)
	if err != nil {
		log.Fatal(err)
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
			log.Println(err)
		}
	}(srv)
	log.Println("Serving " + srv.Addr + " path " + path)
	ch := make(chan os.Signal, 2)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Println(err)
			ch <- syscall.SIGABRT
		}
	}()
	sig := <-ch
	if sig == syscall.SIGABRT {
		os.Exit(1)
	}
}
