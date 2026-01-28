// Package aws
// refer:
//  1. what is aws kms: https://docs.aws.amazon.com/kms/latest/developerguide/overview.html
//  2. aws sdk for go: https://aws.github.io/aws-sdk-go-v2/docs/
package aws

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/kms"
)

var kmsClient *kms.Client
var kmsKeyID string

func InitKMS(ctx context.Context, region string, key string, secret string, id string) error {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(key, secret, ``)),
	)
	if err != nil {
		return err
	}

	// Create an Amazon S3 service client
	kmsClient = kms.NewFromConfig(cfg)
	kmsKeyID = id
	return nil
}

var ErrorUninitialized = fmt.Errorf(`kms client is uninitialized`)

func IsInitialized() bool {
	return kmsClient != nil
}

func EncryptString(plainText string) (string, error) {
	if !IsInitialized() {
		return ``, ErrorUninitialized
	}
	if plainText == `` {
		return ``, nil
	}
	result, err := kmsClient.Encrypt(context.TODO(), &kms.EncryptInput{KeyId: &kmsKeyID, Plaintext: []byte(plainText)})
	if err != nil {
		return ``, err
	}
	ciphertextBlob := base64.StdEncoding.EncodeToString(result.CiphertextBlob)
	return ciphertextBlob, nil
}

func DecryptString(ciphertextB64 string) (string, error) {
	if !IsInitialized() {
		return ``, ErrorUninitialized
	}
	if ciphertextB64 == `` {
		return ``, nil
	}
	ciphertextBlob, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return ``, err
	}
	result, err := kmsClient.Decrypt(context.TODO(), &kms.DecryptInput{CiphertextBlob: ciphertextBlob, KeyId: &kmsKeyID})
	if err != nil {
		return ``, err
	}

	return string(result.Plaintext), err
}

func DecryptStrings(ciphers ...*string) error {
	for _, cipher := range ciphers {
		if *cipher == `` {
			continue
		}
		plain, err := DecryptString(*cipher)
		if err != nil {
			return err
		}
		*cipher = plain
	}
	return nil
}
