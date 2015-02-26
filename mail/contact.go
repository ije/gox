package mail

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/ije/go/utils/valid"
)

type Contact struct {
	Name, Addr string
}

func (c Contact) String() string {
	if len(c.Name) > 0 {
		if strings.Contains(c.Name, " ") {
			return fmt.Sprintf(`"%s" <%s>`, c.Name, c.Addr)
		} else {
			return fmt.Sprintf("%s <%s>", c.Name, c.Addr)
		}
	} else {
		return c.Addr
	}
}

type Contacts map[string]string

func (c Contacts) List() []string {
	var i int
	list := make([]string, len(c))
	for _, addr := range c {
		if valid.IsEmail(addr) {
			list[i] = addr
			i++
		}
	}
	return list[:i]
}

func (c Contacts) String() string {
	buf := bytes.NewBuffer(nil)
	for name, addr := range c {
		if valid.IsEmail(addr) {
			if len(name) > 0 {
				if strings.Contains(name, " ") {
					fmt.Fprintf(buf, `"%s" <%s>, `, name, addr)
				} else {
					fmt.Fprintf(buf, "%s <%s>, ", name, addr)
				}
			} else {
				fmt.Fprint(buf, addr, ", ")
			}
		}
	}
	if l := buf.Len(); l >= 2 {
		return buf.String()[:l-2]
	}
	return ""
}
