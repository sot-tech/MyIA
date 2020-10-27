// Simple http server which returns requester IP address
package main

import (
	"errors"
	"flag"
	"log"
	"net/http"
	"net/netip"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var errNoAddress = errors.New("no listen address provided")

const requestTimeout = time.Millisecond * 500

func main() {
	var listenAddr, listenPath string
	flag.StringVar(&listenAddr, "l", "127.0.0.1:1234", "Listen address")
	flag.StringVar(&listenPath, "p", "/", "Listen path")
	flag.Parse()
	if len(listenAddr) == 0 {
		log.Fatal(errNoAddress)
	}

	if !strings.HasSuffix(listenPath, "/") {
		listenPath = "/" + listenPath
	}
	http.HandleFunc(listenPath, myIP)
	srv := &http.Server{
		Addr:         listenAddr,
		ReadTimeout:  requestTimeout,
		WriteTimeout: requestTimeout,
	}
	defer func(srv *http.Server) {
		err := srv.Close()
		if err != nil {
			log.Println(err)
		}
	}(srv)
	log.Println("Starting serving " + listenAddr + listenPath)
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

func myIP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		if addrPort, err := netip.ParseAddrPort(req.RemoteAddr); err == nil {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(addrPort.Addr().String()))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
		}
	case http.MethodOptions:
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
