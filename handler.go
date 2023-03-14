// Package myia contains http.Handler which returns requester IP address
package myia

import (
	"errors"
	"log"
	"net/http"
	"net/netip"
	"net/textproto"
	"net/url"
	"path"
	"strings"
)

var errMalformedACAOURL = errors.New("malformed Access-Control-Allow-Origin URL")

const (
	httpAllowHeader = "Allow"
	httpACAOHeader  = "Access-Control-Allow-Origin"
	httpACAMHeader  = "Access-Control-Allow-Methods"
	httpAll         = "*"
)

var httpMethods = []string{
	http.MethodGet + ", " + http.MethodOptions + ", " + http.MethodHead,
}

type handler struct {
	path, ipHeader string
	acao           []string
	prefix         netip.Prefix
}

// NewHandler creates new IP detection http.Handler.
// If acao parameter not provided, it will be set to `*`.
// Non-empty net parameter will filter returned IP with provided
// network prefix (i.e. 10.0.0.0/8).
// Non-empty ipHeader parameter will fetch client IP from provided HTTP header.
// If header value contains several values (or comma-separated), only first address is used.
func NewHandler(uPath, acao, net, ipHeader string) (http.Handler, error) {
	if !path.IsAbs(uPath) {
		uPath = "/" + uPath
	}
	uPath = path.Clean(uPath)

	if len(acao) > 0 && acao != httpAll {
		if u, err := url.Parse(acao); err != nil {
			return nil, err
		} else if !u.IsAbs() || len(u.Host) == 0 {
			return nil, errMalformedACAOURL
		}
	}

	var prefix netip.Prefix
	if len(net) > 0 {
		var err error
		if prefix, err = netip.ParsePrefix(net); err != nil {
			return nil, err
		}
	}
	return &handler{
		path:     uPath,
		ipHeader: textproto.CanonicalMIMEHeaderKey(ipHeader),
		acao:     []string{acao},
		prefix:   prefix,
	}, nil
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == httpAll {
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

	if len(h.acao) > 0 && len(h.acao[0]) > 0 {
		w.Header()[httpACAOHeader] = h.acao
		w.Header()[httpACAMHeader] = httpMethods
	}
	switch r.Method {
	case http.MethodOptions:
		w.Header()[httpAllowHeader] = httpMethods
		w.WriteHeader(http.StatusNoContent)
	case http.MethodHead:
		w.WriteHeader(http.StatusNoContent)
	case http.MethodGet:
		w.WriteHeader(http.StatusOK)
		var clientAddress string
		if len(h.ipHeader) > 0 {
			addrs := r.Header[h.ipHeader]
			if len(addrs) > 0 {
				clientAddress = addrs[0]
				// some headers contain addresses separated by comma,
				// we will get only first (nearest to client i.e. x-forwarded-for)
				if commaIndex := strings.IndexRune(clientAddress, ','); commaIndex >= 0 {
					clientAddress = clientAddress[:commaIndex]
				}
				clientAddress = strings.TrimSpace(clientAddress)
			}
		} else {
			clientAddress = r.RemoteAddr
		}
		var addr netip.Addr
		if len(clientAddress) > 0 {
			if addrPort, err := netip.ParseAddrPort(clientAddress); err == nil {
				addr = addrPort.Addr()
			} else if addr, err = netip.ParseAddr(clientAddress); err != nil {
				log.Println(clientAddress, err)
			}
			if addr.IsValid() && (!h.prefix.IsValid() || h.prefix.Contains(addr)) {
				_, _ = w.Write([]byte(addr.String()))
			}
		}
	default:
		w.Header()[httpAllowHeader] = httpMethods
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
