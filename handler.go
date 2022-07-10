// Package myia contains http.ServeMux which returns requester IP address
package myia

import (
	"log"
	"net/http"
	"net/netip"
	"path"
)

const (
	httpAllowHeader = "Allow"
	httpACAOHeader  = "Access-Control-Allow-Origin"
	httpACAMHeader  = "Access-Control-Allow-Methods"
)

var httpMethods = []string{
	http.MethodGet + ", " + http.MethodOptions + ", " + http.MethodHead,
}

type handler struct {
	path   string
	acao   []string
	prefix netip.Prefix
}

// NewHandler creates new IP detection http.Handler.
// If acao parameter not provided, it will be set to `*`.
// Non-empty net parameter will filter returned IP with provided
// network prefix (i.e. 10.0.0.0/8)
func NewHandler(uPath, acao, net string) (http.Handler, error) {
	if !path.IsAbs(uPath) {
		uPath = "/" + uPath
	}
	uPath = path.Clean(uPath)

	var prefix netip.Prefix
	if len(net) > 0 {
		var err error
		if prefix, err = netip.ParsePrefix(net); err != nil {
			return nil, err
		}
	}
	return &handler{
		path:   uPath,
		acao:   []string{acao},
		prefix: prefix,
	}, nil

}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "*" {
		if r.ProtoAtLeast(1, 1) {
			w.Header()["Connection"] = []string{"close"}
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if path.Clean(r.URL.Path) != h.path {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodOptions:
		w.Header()[httpAllowHeader] = httpMethods
		w.Header()[httpACAMHeader] = httpMethods
		w.Header()[httpACAOHeader] = h.acao
		w.WriteHeader(http.StatusNoContent)
	case http.MethodHead:
		w.Header()[httpACAOHeader] = h.acao
		w.WriteHeader(http.StatusNoContent)
	case http.MethodGet:
		w.Header()[httpACAOHeader] = h.acao
		w.WriteHeader(http.StatusOK)
		if addrPort, err := netip.ParseAddrPort(r.RemoteAddr); err != nil {
			log.Println(addrPort, err)
		} else if !h.prefix.IsValid() || h.prefix.Contains(addrPort.Addr()) {
			_, _ = w.Write([]byte(addrPort.Addr().String()))
		}
	default:
		w.Header()[httpAllowHeader] = h.acao
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
