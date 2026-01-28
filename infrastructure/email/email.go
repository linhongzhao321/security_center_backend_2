package email

import (
	"gopkg.in/gomail.v2"
)

type Client struct {
	dialer *gomail.Dialer
}

func NewClient(sender string, password string) *Client {
	dialer := gomail.NewDialer(`smtp.gmail.com`, 587, sender, password)
	return &Client{dialer: dialer}
}
