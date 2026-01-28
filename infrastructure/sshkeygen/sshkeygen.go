package sshkeygen

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os/exec"
	"strings"

	"golang.org/x/crypto/ssh"
)

// Execute
// ssh-keygen -s $CA_FILENAME -V $VALID_PERIOD -I $name \
// -O source-address=$IP_ALLOW_LIST \
// -O extension:login@github.com=$name $pub_key_path
// return cert filename
func Execute(ctx context.Context, caFile string, publicKeyFile string, options ...Option) (string, error) {
	args := []string{
		`-s`, caFile,
	}
	for _, option := range options {
		k, v := option()
		args = append(args, k, v)
	}
	args = append(args, publicKeyFile)

	cmd := exec.CommandContext(ctx, `ssh-keygen`, args...)
	err := cmd.Run()
	if err != nil {
		output, outputErr := cmd.Output()
		fmt.Println(cmd.String())
		fmt.Println(err.Error(), string(output), outputErr)
		return ``, err
	}
	return strings.Replace(publicKeyFile, `.pub`, `-cert.pub`, 1), nil
}

type Option func() (key, val string)

func ValidPeriod(period string) Option {
	return func() (key, val string) {
		return `-V`, period
	}
}

func CertificateIdentity(id string) Option {
	return func() (key, val string) {
		return `-I`, id
	}
}

func SourceAddress(address ...string) Option {
	return func() (key, val string) {
		return `-O`, `source-address=` + strings.Join(address, `,`)
	}
}

func GithubUsername(username string) Option {
	return func() (key, val string) {
		return `-O`, `extension:login@github.com=` + username
	}
}

func GenerateKeypair() (string, string, error) {

	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return ``, ``, err
	}

	sshPublicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return ``, ``, err
	}
	sshPublicKey.Type()
	serializesPublicKey := ssh.MarshalAuthorizedKey(sshPublicKey)
	ssh.ParseAuthorizedKey(serializesPublicKey)

	pemPrivateKey := pem.EncodeToMemory(&pem.Block{
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		Type:  "RSA PRIVATE KEY",
	})

	return string(pemPrivateKey), string(serializesPublicKey), nil
}
