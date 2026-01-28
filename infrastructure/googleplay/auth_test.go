package googleplay

import (
	"testing"
)

func FuzzParseQuery(f *testing.F) {
	f.Add(`name=foo`)
	f.Fuzz(func(t *testing.T, data string) {
		_, err := ParseQuery(data)
		if err != nil {
			t.Fatal(err)
		}
	})
}
