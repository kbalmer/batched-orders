package sftp

type SFTP interface {
	Upload(data []byte) error
	Download(ids []string) ([]byte, error)
}

type Client struct {
}

func NewSFTPClient() SFTP {
	return &Client{}
}

func (s *Client) Upload(data []byte) error {
	return nil
}

func (s *Client) Download(ids []string) ([]byte, error) {
	return nil, nil
}
