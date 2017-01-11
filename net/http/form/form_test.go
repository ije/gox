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
	ret, err := Upload("PUT", "http://x-stud.io/api/file", tmp, "file", map[string]string{
		"saveas":   "test/hello.txt",
		"email":    "",
		"password": "",
	})
	t.Log(string(ret), err)
}
