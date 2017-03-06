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

	"github.com/ije/gox/valid"
)

var CRLF = []byte("\r\n")

type Mail struct {
	Subject     string
	PlainText   []byte
	Html        []byte
	Attachments []*Attachment
}

type Attachment struct {
	Name        string
	ContentType string
	io.Reader
}

func (mail *Mail) MakeBody(from *mail.Address, to AddressList) []byte {
	buf := &buffer{bytes.NewBuffer(nil)}
	buf.writeln("MIME-Version: 1.0")
	buf.writeln("Date: ", time.Now().Format(time.RFC1123Z))
	buf.writeln("Subject: ", encodeSubject(mail.Subject))
	buf.writeln("From: ", from)
	buf.writeln("To: ", to)
	var boundary string
	if len(mail.Attachments) > 0 {
		boundary = boundaryGen()
		buf.writeln("Content-Type: multipart/mixed; boundary=", boundary)
		buf.writeln()
		buf.writeln("--", boundary)
	}
	if len(mail.PlainText) > 0 && len(mail.Html) > 0 {
		cBoundary := boundaryGen()
		buf.writeln("Content-Type: multipart/alternative; boundary=", cBoundary)
		buf.writeln()
		buf.writeln("--", cBoundary)
		buf.writeTextBody(mail.PlainText)
		buf.writeln()
		buf.writeln()
		buf.writeln("--", cBoundary)
		buf.writeHtmlBody(mail.Html)
		buf.writeln()
		buf.writeln()
		buf.writeln("--", cBoundary, "--")
	} else if len(mail.PlainText) > 0 {
		buf.writeTextBody(mail.PlainText)
	} else if len(mail.Html) > 0 {
		buf.writeHtmlBody(mail.Html)
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

func Address(a ...string) *mail.Address {
	var name string
	var address string
	if len(a) == 1 {
		if len(a[0]) > 0 {
			addr, err := mail.ParseAddress(a[0])
			if err == nil {
				return addr
			}
		}
	} else if len(a) > 1 {
		name = a[0]
		address = a[1]
	}

	if len(address) == 0 || !valid.IsEmail(address) {
		return nil
	}

	return &mail.Address{
		Name:    name,
		Address: address,
	}
}

type AddressList []*mail.Address

func (list AddressList) String() string {
	var ss []string
	for _, addr := range list {
		ss = append(ss, addr.String())
	}
	return strings.Join(ss, ", ")
}

func (list AddressList) List() []string {
	var ss []string
	for _, addr := range list {
		ss = append(ss, addr.Address)
	}
	return ss
}

type buffer struct {
	*bytes.Buffer
}

func (buf *buffer) writeln(s ...interface{}) {
	fmt.Fprint(buf, s...)
	buf.Write(CRLF)
}

func (buf *buffer) writeTextBody(text []byte) {
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

func (buf *buffer) writeHtmlBody(html []byte) {
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

func boundaryGen() string {
	h := md5.New()
	fmt.Fprint(h, time.Now().UnixNano(), rand.Int())
	return fmt.Sprintf("--%s--", hex.EncodeToString(h.Sum(nil)))
}
