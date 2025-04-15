package sftp

import "batched-orders/pkg"

type SFTP interface {
	Upload(orders []pkg.Order) error
	Download(orders []pkg.Order) error
}

type Client struct {
}

func NewSFTPClient() SFTP {
	return &Client{}
}

func (s *Client) Upload(orders []pkg.Order) error {
	return nil
}

func (s *Client) Download(orders []pkg.Order) error {
	return nil
}
