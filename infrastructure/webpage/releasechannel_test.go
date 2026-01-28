package webpage

import (
	"context"
	"fmt"
)

func ExampleReleaseChannel_GetLastRelease() {

	channel := MustNewReleaseChannel(`www.coinex.com`)
	got, err := channel.GetLastRelease(context.Background(), nil)
	fmt.Println(got, err)
	// Output:
	//
}
