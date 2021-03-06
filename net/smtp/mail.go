package smtp

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"net/mail"
	"strings"
	"time"
)

// CRLF defines the newline
var CRLF = []byte("\r\n")

// Mail body for smtp
type Mail struct {
	Subject     string
	PlainText   []byte
	HTML        []byte
	Attachments []*Attachment
}

// Attachment for mail
type Attachment struct {
	Name        string
	ContentType string
	io.Reader
}

// AddressList defines a list of mail.Address
type AddressList []*mail.Address

// List returns the list of mail address
func (list AddressList) List() []string {
	var ss []string
	for _, addr := range list {
		ss = append(ss, addr.Address)
	}
	return ss
}

func (list AddressList) String() string {
	var ss []string
	for _, addr := range list {
		ss = append(ss, addr.String())
	}
	return strings.Join(ss, ", ")
}

// AppendAttachment appends an attachment
func (mail *Mail) AppendAttachment(attachment *Attachment) {
	if attachment != nil {
		mail.Attachments = append(mail.Attachments, attachment)
	}
}

// Encode encodes mail to mime bytes
// TODO: add cc and bcc support
func (mail *Mail) Encode(from *mail.Address, to AddressList) []byte {
	buf := &mailBuffer{bytes.NewBuffer(nil)}
	buf.writeln("MIME-Version: 1.0")
	buf.writeln("Date: ", time.Now().Format(time.RFC1123Z))
	buf.writeln("Subject: ", encodeSubject(mail.Subject))
	buf.writeln("From: ", from)
	buf.writeln("To: ", to)
	var boundary string
	if len(mail.Attachments) > 0 {
		boundary = newBoundary()
		buf.writeln("Content-Type: multipart/mixed; boundary=", boundary)
		buf.writeln()
		buf.writeln("--", boundary)
	}
	if len(mail.PlainText) > 0 && len(mail.HTML) > 0 {
		cBoundary := newBoundary()
		buf.writeln("Content-Type: multipart/alternative; boundary=", cBoundary)
		buf.writeln()
		buf.writeln("--", cBoundary)
		buf.writeTextBody(mail.PlainText)
		buf.writeln()
		buf.writeln()
		buf.writeln("--", cBoundary)
		buf.writeHTMLBody(mail.HTML)
		buf.writeln()
		buf.writeln()
		buf.writeln("--", cBoundary, "--")
	} else if len(mail.PlainText) > 0 {
		buf.writeTextBody(mail.PlainText)
	} else if len(mail.HTML) > 0 {
		buf.writeHTMLBody(mail.HTML)
	}
	if len(mail.Attachments) > 0 {
		for _, attchment := range mail.Attachments {
			buf.writeln()
			buf.writeln()
			buf.writeln("--", boundary)
			buf.writeln("Content-Type: ", attchment.ContentType, "; name=", attchment.Name, ";")
			buf.writeln("Content-Transfer-Encoding: base64")
			buf.writeln("Content-Disposition: attachment; filename=", attchment.Name, ";")
			buf.writeln()
			encoder := base64.NewEncoder(base64.StdEncoding, buf)
			io.Copy(encoder, attchment)
			encoder.Close()
		}
		buf.writeln()
		buf.writeln()
		buf.writeln("--", boundary, "--")
	}
	return buf.Bytes()
}

type mailBuffer struct {
	*bytes.Buffer
}

func (buf *mailBuffer) writeln(s ...interface{}) {
	fmt.Fprint(buf, s...)
	buf.Write(CRLF)
}

func (buf *mailBuffer) writeTextBody(text []byte) {
	buf.writeln("Content-Type: text/plain; charset=UTF-8")

	for _, c := range text {
		if c > 127 {
			buf.writeln("Content-Transfer-Encoding: base64")
			buf.writeln()
			buf.WriteString(base64.StdEncoding.EncodeToString(text))
			return
		}
	}

	buf.writeln()
	buf.Write(text)
}

func (buf *mailBuffer) writeHTMLBody(html []byte) {
	buf.writeln("Content-Type: text/html; charset=UTF-8")

	for _, c := range html {
		if c > 127 {
			buf.writeln("Content-Transfer-Encoding: quoted-printable")
			buf.writeln()
			var c byte
			for i, l := 0, len(html); i < l; i++ {
				if c = html[i]; c > 127 {
					fmt.Fprintf(buf, "=%X", c)
					i++
					fmt.Fprintf(buf, "=%X", html[i])
					i++
					fmt.Fprintf(buf, "=%X", html[i])
				} else if c == '=' {
					buf.WriteString("=3D")
				} else {
					buf.WriteByte(c)
				}
			}
			return
		}
	}

	buf.writeln()
	buf.Write(html)
}

func encodeSubject(subject string) string {
	for _, c := range subject {
		if c > 127 {
			return fmt.Sprintf("=?UTF-8?B?%s?=", base64.StdEncoding.EncodeToString([]byte(subject)))
		}
	}
	return subject
}

func newBoundary() string {
	h := md5.New()
	fmt.Fprint(h, time.Now().UnixNano(), rand.Int())
	return fmt.Sprintf("--%s--", hex.EncodeToString(h.Sum(nil)))
}
