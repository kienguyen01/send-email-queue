package elk

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
)

type Message struct {
	SenderEmail   string
	SenderName    string
	ReceiverEmail string
	ReceiverName  string
	Body          string
	Subject       string
	Timestamp     time.Time
}

type ELKClient struct {
	client *elasticsearch.Client
}

func NewELKClient(host string, port string) (*ELKClient, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{
			fmt.Sprintf("https://%s:%s", host, port),
		},
		Username: os.Getenv("Username"),
		Password: os.Getenv("Password"),
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &ELKClient{client: client}, nil
}

func (c *ELKClient) SendLog(index string, message interface{}) error {
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		return err
	}
	res, err := c.client.Index(
		index,
		strings.NewReader(string(jsonMessage)),
		c.client.Index.WithContext(context.Background()),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return nil
}

func SendMessageToELK(client *ELKClient, message *Message, index string) error {
	return client.SendLog(index, message)
}
