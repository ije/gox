package form

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
)

func Upload(method string, url string, filepath string, fileFieldName string, extraHeader map[string]string) (ret []byte, err error) {
	// caveat IMO dont use this for large files, create a tmpfile and assemble your multipart from there (not tested)
	buf := bytes.NewBuffer(nil)
	w := multipart.NewWriter(buf)

	// Create file field
	fw, err := w.CreateFormFile(fileFieldName, path.Base(filepath))
	if err != nil {
		return
	}

	fd, err := os.Open(filepath)
	if err != nil {
		return
	}
	defer fd.Close()

	// Write file field from file to upload
	_, err = io.Copy(fw, fd)
	if err != nil {
		return
	}

	for key, val := range extraHeader {
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

	req, err := http.NewRequest(method, url, buf)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	var client http.Client
	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		err = errors.New(res.Status)
		return
	}

	buf = bytes.NewBuffer(nil)
	_, err = io.Copy(buf, res.Body)
	if err != nil {
		return
	}

	ret = buf.Bytes()
	return
}
