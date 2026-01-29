package googleplay

import (
	"context"
	"fmt"
	"os"
	"testing"
)

func TestDelivery(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	home += "/google/play"
	var token Token
	token.Data, err = os.ReadFile(home + "/token.txt")
	if err != nil {
		t.Fatal(err)
	}
	if err := token.Unmarshal(); err != nil {
		t.Fatal(err)
	}
	var auth GoogleAuth
	if err := auth.Auth(context.Background(), token); err != nil {
		t.Fatal(err)
	}
	var checkin Checkin
	checkin.Data, err = os.ReadFile(home + "/x86.bin")
	if err != nil {
		t.Fatal(err)
	}
	if err := checkin.Unmarshal(); err != nil {
		t.Fatal(err)
	}
	deliver, err := checkin.Delivery(context.Background(), auth, "com.google.android.youtube", 1524221376, false)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%#v\n", deliver.m)
}
