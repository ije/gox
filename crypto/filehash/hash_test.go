package filehash

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestHash(t *testing.T) {
	tmpFile := path.Join(os.TempDir(), "gox-filehash-test.bin")
	ioutil.WriteFile(tmpFile, []byte("hello world!"), 0644)
	t.Log(Hash(tmpFile))
}
