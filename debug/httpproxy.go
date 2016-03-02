package debug

const HTTP_PROXY_SERVER_SRC = `
package main

import (
	"net/http"
	"strings"
	"fmt"
	"io"

	"github.com/ije/gox/utils"
)

var rules = map[string]string%s

func main() {
	http.ListenAndServe(":80", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var backServer string

		host, _ := utils.SplitByLastByte(r.Host, ':')
		for match, server := range rules {
			if len(server) > 0 && len(match) > 0 && (match == "*" || strings.Contains(match, host)) {
				backServer = server
				break
			}
		}

		if len(backServer) == 0 {
			http.Error(w, "BackServer Not Found", 400)
			return
		}

		req, err := http.NewRequest(r.Method, fmt.Sprintf("http://%%s%%s", backServer, r.RequestURI), r.Body)
		if err != nil {
			http.Error(w, "Proxy Server Error", 500)
			return
		}

		req.Host = host
		for key, values := range r.Header {
			req.Header[key] = values
		}
		remoteIp, _ := utils.SplitByLastByte(r.RemoteAddr, ':')
		req.Header.Set("X-Forwarded-For", remoteIp)

		resp, err := new(http.Client).Do(req)
		if err != nil {
			http.Error(w, "Proxy Server Error", 500)
			return
		}
		defer resp.Body.Close()

		header := w.Header()
		for key, values := range resp.Header {
			header[key] = values
		}
		w.WriteHeader(resp.StatusCode)

		io.Copy(w, resp.Body)
	}))
}`
