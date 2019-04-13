package debug

import (
	"fmt"
)

const HTTP_PROXY_SERVER_SRC = `
package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/ije/gox/utils"
)

func main() {
	http.ListenAndServe(":%d", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		req, err := http.NewRequest(r.Method, fmt.Sprintf("%s://%s%%s", r.RequestURI), r.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		req.Host = r.Host
		for key, values := range r.Header {
			req.Header[key] = values
		}
		remoteIp, _ := utils.SplitByLastByte(r.RemoteAddr, ':')
		req.Header.Set("X-Forwarded-For", remoteIp)

		client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return fmt.Errorf("Redirected!")
		}}
		resp, err := client.Do(req)
		if err != nil {
			if resp != nil && resp.StatusCode > 300 && resp.StatusCode < 400 {
				http.Redirect(w, r, resp.Header.Get("Location"), resp.StatusCode)
			} else {
				http.Error(w, err.Error(), 500)
			}
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
}
`

func UseHttpProxy(port uint16, sudo bool, backServer string, backServerProtocol string) (err error) {
	if backServer == "" {
		return fmt.Errorf("bad back server '%s'", backServer)
	}

	if backServerProtocol != "http" && backServerProtocol != "https" {
		return fmt.Errorf("bad back protocol '%s'", backServerProtocol)
	}

	return AddProcess(&Process{
		Sudo:   sudo,
		Name:   "http-proxy",
		GoCode: fmt.Sprintf(HTTP_PROXY_SERVER_SRC, port, backServerProtocol, backServer),
	})
}
