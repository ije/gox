package filehash

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/ije/gox/crypto/rs"
)

func TestHash(t *testing.T) {
	tmpFile := path.Join(os.TempDir(), rs.Hex.String(16)+".bin")
	err := ioutil.WriteFile(tmpFile, []byte("hello world!"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	hash, err := Hash(tmpFile)
	if err != nil {
		t.Fatal(err)
	}

	want := "fc3ff98e8c6a0d3087d515c0473f8677" + "03b4c26d"
	if hash != want {
		t.Fatalf("bad hash '%s' want '%s'", hash, want)
	}
}
