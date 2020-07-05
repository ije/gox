package utils

import (
	"testing"
)

func TestGetLocalIPList(t *testing.T) {
	list, err := GetLocalIPList()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(list)
}
