package gwt

import (
	"testing"
	"time"
)

func TestJWT(t *testing.T) {
	gwt := &GWT{"gwt-secret"}

	type User struct {
		UID  uint32
		Name string
		Age  uint8
	}
	rawPayload := User{123, "x", 18}
	tokenString, err := gwt.SignToken(rawPayload, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(tokenString)

	type User2 struct {
		UID   uint32
		Name  string
		Class string
	}
	var payload User2
	err = gwt.ParseToken(tokenString, &payload)
	if err != nil {
		t.Fatal(err)
	}

	if payload.UID != rawPayload.UID {
		t.Fatalf("invalid UID %d want %d", payload.UID, rawPayload.UID)
	}
	if payload.Name != rawPayload.Name {
		t.Fatalf("invalid Name '%s' want '%s'", payload.Name, rawPayload.Name)
	}

	time.Sleep(2 * time.Second)
	err = gwt.ParseToken(tokenString, &payload)
	if err == nil || !IsExpires(err) {
		t.Fatal(err)
	}
}
