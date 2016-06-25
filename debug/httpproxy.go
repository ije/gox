package debug

const HTTP_PROXY_SERVER_SRC = `
package main

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"

	"github.com/ije/gox/utils"
)

var (
	hostRules   map[string]string
	regexpRules []RegexpRule
)

type RegexpRule struct {
	Regexp *regexp.Regexp
	Server string
}

type RegexpRules []RegexpRule

func (a RegexpRules) Len() int           { return len(a) }
func (a RegexpRules) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a RegexpRules) Less(i, j int) bool { return len(a[j].Regexp.String()) < len(a[i].Regexp.String()) }

func main() {
	if len(hostRules) == 0 && len(regexpRules) == 0 {
		fmt.Println("proxy rules is empty")
		return
	}

	http.ListenAndServe(":80", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var backServer string

		for host, server := range hostRules {
			if host == r.Host {
				backServer = server
				break
			}
		}

		if len(backServer) == 0 {
			for _, rule := range regexpRules {
				if rule.Regexp != nil && rule.Regexp.MatchString(r.Host) {
					backServer = rule.Server
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

func init() {
	hostRules = map[string]string{}
	for match, server := range map[string]string%s {
		if match = strings.TrimSpace(match); len(match) > 0 && len(server) > 0 && regexp.MustCompile("^[a-zA-Z0-9\\-\\.\\*]+$").MatchString(match) {
			m := strings.ToLower(match)
			if m[0] == '.' {
				if !strings.ContainsRune(m, '*') {
					hostRules[m[1:]] = server
				}
				m = "*" + m
			}
			if strings.ContainsRune(m, '*') {
				regexpRules = append(regexpRules, RegexpRule{regexp.MustCompile("^" + strings.Replace(strings.Replace(strings.Replace(m, ".", "\\.", -1), "-", "\\-", -1), "*", ".*?", -1) + "$"), server})
			} else {
				hostRules[m] = server
			}
		}
	}
	if len(regexpRules) > 0 {
		sort.Sort(RegexpRules(regexpRules))
	}
}
`
