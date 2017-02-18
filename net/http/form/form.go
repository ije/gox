package form

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	URL "net/url"
	"os"
	"path"
	"strings"
)

func Form(url string, method string, params map[string]string, files map[string]string) (resp *http.Response, err error) {
	var buf *bytes.Buffer
	var contentType string
	var contentSize int

	if len(files) == 0 {
		if len(params) > 0 {
			values := URL.Values{}
			for key, value := range params {
				values.Add(key, value)
			}
			if strings.ContainsRune(url, '?') {
				url += "&"
			} else {
				url += "?"
			}

			if method == "GET" || method == "HEAD" || method == "DELETE" {
				url += values.Encode()
			} else {
				buf = bytes.NewBufferString(values.Encode())
				contentSize = buf.Len()
				contentType = "application/x-www-form-urlencoded"
			}
		}
	} else {
		if method == "GET" || method == "HEAD" || method == "DELETE" {
			err = errors.New("Invalid Method")
			return
		}

		// caveat IMO dont use this for large files, create a tmpfile and assemble your multipart from there (not tested)
		buf = bytes.NewBuffer(nil)
		w := multipart.NewWriter(buf)

		for fieldName, filePath := range files {
			var fw io.Writer
			var fd *os.File

			fw, err = w.CreateFormFile(fieldName, path.Base(filePath))
			if err != nil {
				return
			}

			fd, err = os.Open(filePath)
			if err != nil {
				return
			}

			_, err = io.Copy(fw, fd)
			fd.Close()
			if err != nil {
				return
			}
		}

		for key, val := range params {
			var fw io.Writer
			fw, err = w.CreateFormField(key)
			if err != nil {
				return
			}
			_, err = fw.Write([]byte(val))
			if err != nil {
				return
			}
		}

		// Important if you do not close the multipart writer you will not have a terminating boundry
		w.Close()
		contentType = w.FormDataContentType()
	}

	req, err := http.NewRequest(method, url, buf)
	if err != nil {
		return
	}
	if len(contentType) > 0 {
		req.Header.Set("Content-Type", contentType)
	}
	if contentSize > 0 {
		req.ContentLength = int64(contentSize)
	}

	var client http.Client
	resp, err = client.Do(req)
	return
}
