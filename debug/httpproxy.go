package debug

const HTTP_PROXY_SERVER_SRC = `
package main

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/ije/gox/utils"
)

var rules map[*rule]string

type rule struct {
	host   string
	regexp *regexp.Regexp
}

func main() {
	if len(rules) == 0 {
		fmt.Println("proxy rules is empty")
		return
	}

	http.ListenAndServe(":80", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var backServer string

		for match, server := range rules {
			if match.host == r.Host {
				backServer = server
				break
			}
		}

		if len(backServer) == 0 {
			for match, server := range rules {
				if match.regexp != nil && match.regexp.MatchString(r.Host) {
					backServer = server
					break
				}
			}
		}

		if len(backServer) == 0 {
			http.Error(w, "Back Server Not Found", 400)
			return
		}

		req, err := http.NewRequest(r.Method, fmt.Sprintf("http://%%s%%s", backServer, r.RequestURI), r.Body)
		if err != nil {
			http.Error(w, "Proxy Server Error", 500)
			return
		}

		req.Host = r.Host
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
}

func init() {
	rules = map[*rule]string{}
	for match, server := range map[string]string%s {
		if match = strings.TrimSpace(match); len(match) > 0 && len(server) > 0 && regexp.MustCompile("^[a-zA-Z0-9\\-\\.\\*]+$").MatchString(match) {
			r := &rule{}
			m := strings.ToLower(match)
			if m[0] == '.' {
				if !strings.ContainsRune(m, '*') {
					r.host = m[1:]
				}
				m = "*" + m
			}
			if strings.ContainsRune(m, '*') {
				r.regexp = regexp.MustCompile("^" + strings.Replace(strings.Replace(strings.Replace(m, ".", "\\.", -1), "-", "\\-", -1), "*", ".*?", -1) + "$")
			} else {
				r.host = m
			}
			rules[r] = server
		}
	}
}
`
