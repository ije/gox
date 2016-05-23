package smtp

import (
	"fmt"
	"strings"
)

type Contact struct {
	Name, Email string
}

func (contact *Contact) String() string {
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
	list := make([]string, len(contacts))
	for i, contact := range contacts {
		list[i] = contact.Email
	}
	return list
}

func (contacts Contacts) String() string {
	cs := make([]string, len(contacts))
	for i, contact := range contacts {
		cs[i] = contact.String()
	}
	return strings.Join(cs, ",")
}
