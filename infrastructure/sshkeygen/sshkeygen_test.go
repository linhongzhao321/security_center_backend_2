package sshkeygen

import (
	"context"
	"fmt"
	"os"
)

func ExampleExecute() {
	caFilename, err := Execute(context.Background(),
		`test_ca`,
		`test_user.pub`,
		ValidPeriod(`+7d`),
		CertificateIdentity(`funcolin`),
		SourceAddress(`127.0.0.1`),
		GithubUsername(`funcolin`),
	)
	if err != nil {
		panic(err)
	}
	_, err = os.Stat(caFilename)
	if err != nil {
		panic(err)
	}
	fmt.Println(!os.IsNotExist(err))
	_ = os.Remove(caFilename)
	// Output:
	// true
}
