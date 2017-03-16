package form

import (
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"testing"
	"time"
)

func Test(t *testing.T) {
	tmp := path.Join(os.TempDir(), "hello.txt")
	ioutil.WriteFile(tmp, []byte("Hello World!"), 0644)
	defer os.Remove(tmp)

	go func() {
		http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Log("Authorization(header):", r.Header.Get("Authorization"))
			t.Log("key:", r.FormValue("key"))
			_, h, err := r.FormFile("file")
			if err != nil {
				w.WriteHeader(500)
			}
			t.Log("file:", h)
			w.WriteHeader(200)
		}))
	}()
	time.Sleep(time.Second)

	ret, err := Form("http://127.0.0.1:8080", "POST", map[string]string{
		"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte("user:password")),
	}, map[string]string{
		"key": "value",
	}, map[string]string{
		"file": tmp,
	})
	if err != nil {
		t.Fatal(err)
		return
	}

	data, _ := ioutil.ReadAll(ret.Body)
	t.Log(ret, string(data))
}
