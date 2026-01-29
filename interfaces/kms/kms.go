package kms

import (
	"fmt"

	"github.com/spf13/cobra"

	"infrastructure/aws"
)

func NewCommand() *cobra.Command {
	command := &cobra.Command{
		Use:     `kms`,
		Short:   `encrypt secret data`,
		Long:    `using third-party tools such as KMS to encrypt secrets for better data protection`,
		Example: `secret -k <kms-key-key>  <original text 1> <plain text 2> ...<plain text >`,
		RunE: func(cmd *cobra.Command, plainTexts []string) error {
			if !aws.IsInitialized() {
				return aws.ErrorUninitialized
			}

			// encrypt and output
			ciphers := make([]string, len(plainTexts))
			for i, txt := range plainTexts {
				cipher, err := aws.EncryptString(txt)
				if err != nil {
					return err
				}
				ciphers[i] = cipher
			}
			fmt.Println(`Please remember ciphers:`)
			fmt.Printf("\t⚠️%s\n\t⚠️%s\n",
				`If it is lost, it can only be regenerated and cannot be retrieved`,
				`The order of ciphers is the same as the inputted plains`,
			)
			for i, cipher := range ciphers {
				fmt.Printf("%d. %s\n", i+1, cipher)
			}
			return nil
		},
	}

	return command
}
