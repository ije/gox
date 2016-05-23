package smtp

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/smtp"
	"strings"
	"time"

	"github.com/ije/gox/utils"
	"github.com/ije/gox/valid"
)

var CRLF = []byte("\r\n")

var (
	ErrEmptySender     = errors.New("Empty Sender")
	ErrEmptyRecipients = errors.New("Empty Recipients")
	ErrEmptySubject    = errors.New("Empty Subject")
	ErrEmptyContent    = errors.New("Empty Content")
)

type Mail struct {
	to          Contacts
	from        *Contact
	subject     string
	text        []byte
	html        []byte
	attachments []Attachment
	*bytes.Buffer
}

type Attachment struct {
	Name        string
	ContentType string
	io.Reader
}

func NewMail(from, to interface{}, subject, text, html string, attachments []Attachment) (mail *Mail, err error) {
	var sender *Contact
	if from != nil {
		switch a := from.(type) {
		case string:
			name, email := utils.SplitByLastByte(a, ' ')
			if len(email) == 0 {
				email = name
				name = ""
			}
			if email = strings.TrimSpace(email); valid.IsEmail(email) {
				sender = &Contact{Name: strings.TrimSpace(name), Email: email}
			}
		case Contact:
			if a.Email = strings.TrimSpace(a.Email); valid.IsEmail(a.Email) {
				sender = &a
			}
		case *Contact:
			if a.Email = strings.TrimSpace(a.Email); valid.IsEmail(a.Email) {
				sender = a
			}
		}
	}
	if sender == nil {
		err = ErrEmptySender
		return
	}

	var recipients Contacts
	if to != nil {
		switch a := to.(type) {
		case string:
			for _, s := range strings.Split(a, ",") {
				name, email := utils.SplitByLastByte(strings.TrimSpace(s), ' ')
				if len(email) == 0 {
					email = name
					name = ""
				}
				if email = strings.TrimSpace(email); valid.IsEmail(email) {
					recipients = append(recipients, Contact{Email: email, Name: strings.TrimSpace(name)})
				}
			}
		case []string:
			for _, s := range a {
				name, email := utils.SplitByLastByte(strings.TrimSpace(s), ' ')
				if len(email) == 0 {
					email = name
					name = ""
				}
				if email = strings.TrimSpace(email); valid.IsEmail(email) {
					recipients = append(recipients, Contact{Email: email, Name: strings.TrimSpace(name)})
				}
			}
		case map[string]string:
			b := map[string]string{}
			for name, email := range a {
				if email = strings.TrimSpace(email); valid.IsEmail(email) {
					b[email] = strings.TrimSpace(name)
				}
			}
			for email, name := range b {
				recipients = append(recipients, Contact{Email: email, Name: name})
			}
		case Contacts:
			recipients = a
		}
	}
	if recipients == nil {
		err = ErrEmptyRecipients
		return
	}

	if len(subject) == 0 {
		err = ErrEmptySubject
		return
	}

	if len(text) == 0 && len(html) == 0 {
		err = ErrEmptyContent
		return
	}

	if len(text) == 0 {
		text, _ = utils.Html2Text(bytes.NewReader([]byte(html)))
	}

	mail = &Mail{
		from:        sender,
		to:          recipients,
		subject:     subject,
		text:        []byte(text),
		html:        []byte(html),
		attachments: attachments,
		Buffer:      bytes.NewBuffer(nil),
	}
	return
}

func (mail *Mail) Send(s *Smtp) error {
	var boundary, boundary2 string
	mail.writeln("MIME-Version: 1.0")
	mail.writeln("Date: ", time.Now().Format(time.RFC1123Z))
	mail.writeln("Subject: ", encodeSubject(mail.subject))
	mail.writeln("From: ", mail.from)
	mail.writeln("To: ", mail.to)
	if len(mail.attachments) > 0 {
		boundary = bhGen()
		mail.writeln("Content-Type: multipart/mixed; boundary=", boundary)
		mail.writeln()
		mail.writeln("--", boundary)
	}
	if len(mail.text) > 0 && len(mail.html) > 0 {
		boundary2 = bhGen()
		mail.writeln("Content-Type: multipart/alternative; boundary=", boundary2)
		mail.writeln()
		mail.writeln("--", boundary2)
		mail.writeTextBody()
		mail.writeln()
		mail.writeln()
		mail.writeln("--", boundary2)
		mail.writeHtmlBody()
		mail.writeln()
		mail.writeln()
		mail.writeln("--", boundary2, "--")
	} else if len(mail.text) > 0 {
		mail.writeTextBody()
	} else if len(mail.html) > 0 {
		mail.writeHtmlBody()
	}
	if len(mail.attachments) > 0 {
		for _, attchment := range mail.attachments {
			mail.writeln()
			mail.writeln()
			mail.writeln("--", boundary)
			mail.writeln("Content-Type: ", attchment.ContentType, "; name=", attchment.Name, ";")
			mail.writeln("Content-Transfer-Encoding: base64")
			mail.writeln("Content-Disposition: attachment; filename=", attchment.Name, ";")
			mail.writeln()
			encoder := base64.NewEncoder(base64.StdEncoding, mail)
			io.Copy(encoder, attchment)
			encoder.Close()
		}
		mail.writeln()
		mail.writeln()
		mail.writeln("--", boundary, "--")
	}
	return smtp.SendMail(s.addr, s.auth, mail.from.Email, mail.to.EmailList(), mail.Bytes())
}

func (mail *Mail) writeTextBody() {
	mail.writeln("Content-Type: text/plain; charset=UTF-8")

	for _, c := range mail.text {
		if c > 127 {
			mail.writeln("Content-Transfer-Encoding: base64")
			mail.writeln()
			mail.WriteString(base64.StdEncoding.EncodeToString(mail.text))
			return
		}
	}

	mail.writeln()
	mail.Write(mail.text)
}

func (mail *Mail) writeHtmlBody() {
	mail.writeln("Content-Type: text/html; charset=UTF-8")

	for _, c := range mail.html {
		if c > 127 {
			mail.writeln("Content-Transfer-Encoding: quoted-printable")
			mail.writeln()
			var c byte
			for i, l := 0, len(mail.html); i < l; i++ {
				if c = mail.html[i]; c > 127 {
					fmt.Fprintf(mail, "=%X", c)
					i++
					fmt.Fprintf(mail, "=%X", mail.html[i])
					i++
					fmt.Fprintf(mail, "=%X", mail.html[i])
				} else if c == '=' {
					mail.WriteString("=3D")
				} else {
					mail.WriteByte(c)
				}
			}
			return
		}
	}

	mail.writeln()
	mail.Write(mail.html)
}

func (mail *Mail) writeln(s ...interface{}) {
	fmt.Fprint(mail, s...)
	mail.Write(CRLF)
}

func encodeSubject(subject string) string {
	for _, c := range subject {
		if c > 127 {
			return fmt.Sprintf("=?UTF-8?B?%s?=", base64.StdEncoding.EncodeToString([]byte(subject)))
		}
	}
	return subject
}

func bhGen() string {
	h := md5.New()
	fmt.Fprint(h, time.Now().UnixNano(), rand.Int())
	return fmt.Sprintf("--%s--", hex.EncodeToString(h.Sum(nil)))
}
