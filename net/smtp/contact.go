package smtp

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/ije/gox/valid"
)

type Contact struct {
	Name, Email string
}

func (contact Contact) String() string {
	if len(contact.Name) > 0 {
		if strings.Contains(contact.Name, " ") {
			return fmt.Sprintf(`"%s" <%s>`, contact.Name, contact.Email)
		} else {
			return fmt.Sprintf("%s <%s>", contact.Name, contact.Email)
		}
	}
	return contact.Email
}

type Contacts []Contact

func (contacts Contacts) EmailList() []string {
	var i int
	list := make([]string, len(contacts))
	for _, contact := range contacts {
		if valid.IsEmail(contact.Email) {
			list[i] = contact.Email
			i++
		}
	}
	return list[:i]
}

func (contacts Contacts) String() string {
	buf := bytes.NewBuffer(nil)
	for _, contact := range contacts {
		if valid.IsEmail(contact.Email) {
			fmt.Fprint(buf, contact, ", ")
		}
	}
	if l := buf.Len(); l > 2 {
		return buf.String()[:l-2]
	}
	return ""
}
