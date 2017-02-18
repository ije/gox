package form

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func Test(t *testing.T) {
	tmp := path.Join(os.TempDir(), "hello.txt")
	defer os.Remove(tmp)

	ioutil.WriteFile(tmp, []byte("Hello World!"), 0644)
	ret, err := Form("http://x-stud.io/api/file", "PUT", map[string]string{
		"saveas":   "test/hello.txt",
		"email":    "",
		"password": "",
	}, map[string]string{
		"file": tmp,
	})
	t.Log(string(ret), err)
}
