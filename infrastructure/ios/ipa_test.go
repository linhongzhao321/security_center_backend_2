package ios

import (
	"fmt"
	"os"
)

func ExampleNewIPA() {
	testIPA, err := os.Open(`test_ipa`)
	if err != nil {
		panic(err)
	}
	testIPAStat, err := testIPA.Stat()
	if err != nil {
		panic(err)
	}
	got, err := NewIPA(testIPA, testIPAStat.Size(), `Payload/CoinExchange_iOS.app/CoinExchange_iOS`)
	if err != nil {
		panic(err)
	}
	fmt.Println(got.CheckSums())
	// Output:
	// 02BBD5A6-A860-39D3-9B6F-68B2D18B119F
}
