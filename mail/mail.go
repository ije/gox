package mail

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/ije/aisling/utils/valid"
)

var (
	ls                 = []byte("\r\n")
	ErrEmptySubject    = errors.New("Empty Subject")
	ErrEmptyContent    = errors.New("Empty Content")
	ErrEmptySender     = errors.New("Empty Sender")
	ErrEmptyRecipients = errors.New("Empty Recipients")
)

type MailBody struct {
	from     Contact
	to       Contacts
	subject  string
	text     string
	html     string
	boundary string
	*bytes.Buffer
}

func NewMailBody(from Contact, to Contacts, subject, text, html string) (b *MailBody, err error) {
	if len(subject) == 0 {
		err = ErrEmptySubject
		return
	}
	if len(text) == 0 && len(html) == 0 {
		err = ErrEmptyContent
		return
	}
	if len(to) == 0 || len(to.EmailList()) == 0 {
		err = ErrEmptyRecipients
		return
	}
	if !valid.IsEmail(from.Email) {
		err = ErrEmptySender
		return
	}
	b = &MailBody{
		from:    from,
		to:      to,
		subject: subject,
		text:    text,
		html:    html,
		Buffer:  bytes.NewBuffer(nil),
	}
	return
}

func (b *MailBody) Body() []byte {
	b.writeln("MIME-Version: 1.0")
	b.writeln("Date: ", time.Now().Format(time.RFC1123Z))
	b.writeln("Subject: ", encodeSubject(b.subject))
	b.writeln("From: ", b.from)
	b.writeln("To: ", b.to)
	switch {
	case len(b.text) > 0 && len(b.html) > 0:
		boundary := hash()
		b.writeln("Content-Type: multipart/alternative; boundary=", boundary)
		b.writeln()
		b.writeln("--", boundary)
		b.textBody()
		b.writeln()
		b.writeln()
		b.writeln("--", boundary)
		b.htmlBody()
		b.writeln()
		b.writeln()
		fmt.Fprint(b, "--", boundary, "--")
	case len(b.text) > 0:
		b.textBody()
	case len(b.html) > 0:
		b.htmlBody()
	}
	return b.Bytes()
}

func (b *MailBody) writeln(s ...interface{}) {
	fmt.Fprint(b, s...)
	b.Write(ls)
}

func (b *MailBody) textBody() {
	b.writeln("Content-Type: text/plain; charset=UTF-8")

	for _, c := range b.text {
		if c > 127 {
			b.writeln("Content-Transfer-Encoding: base64")
			b.writeln()
			b.WriteString(base64.StdEncoding.EncodeToString([]byte(b.text)))
			return
		}
	}

	b.writeln()
	b.WriteString(b.text)
}

func (b *MailBody) htmlBody() {
	b.writeln("Content-Type: text/html; charset=UTF-8")

	for _, c := range b.html {
		if c > 127 {
			b.writeln("Content-Transfer-Encoding: quoted-printable")
			b.writeln()
			var c byte
			for i, l := 0, len(b.html); i < l; i++ {
				if c = b.html[i]; c > 127 {
					fmt.Fprintf(b, "=%X", c)
					i++
					fmt.Fprintf(b, "=%X", b.html[i])
					i++
					fmt.Fprintf(b, "=%X", b.html[i])
				} else if c == '=' {
					b.WriteString("=3D")
				} else {
					b.WriteByte(c)
				}
			}
			return
		}
	}

	b.writeln()
	b.WriteString(b.html)
}

func encodeSubject(subject string) string {
	for _, c := range subject {
		if c > 127 {
			return fmt.Sprintf("=?UTF-8?B?%s?=", base64.StdEncoding.EncodeToString([]byte(subject)))
		}
	}
	return subject
}

func hash() string {
	h := md5.New()
	fmt.Fprintln(h, time.Now().UnixNano())
	return hex.EncodeToString(h.Sum(nil))
}
