package aws

import (
	"fmt"
)

func Example_decryptString() {
	planText := `this is test`
	cipherBlob, err := EncryptString(planText)
	if err != nil {
		panic(err)
	}
	planText, err = DecryptString(cipherBlob)
	if err != nil {
		panic(err)
	}
	fmt.Println(planText)
	// Output:
	// this is test
}
