package presence_client

import (
	"encoding/json"
	"errors"
	"fmt"
	"gitlab.com/george/shoya-go/models/service_types"
	"io"
	"net/http"
	"time"
)

var PresenceService *Client
var ServerErr = errors.New("presence: service faced error, check logs")

type Client struct {
	c      *http.Client
	Url    string
	ApiKey string
}

func NewClient(url, apiKey string) *Client {
	return &Client{
		c:      &http.Client{},
		Url:    url,
		ApiKey: apiKey,
	}
}

func (c *Client) GetPresenceForUser(id string) (*service_types.UserPresence, error) {
	b, err := c.doRequest(http.MethodGet, fmt.Sprintf("%s/%s", c.Url, id))
	if err != nil {
		return nil, err
	}

	var p = &service_types.UserPresence{}
	err = json.Unmarshal(b, p)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (c *Client) UpdateStatusForUser(id string, status service_types.UserStatus) error {
	_, err := c.doRequest(http.MethodPut, fmt.Sprintf("%s/%s/status/%s", c.Url, id, status))
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) UpdateStateForUser(id string, state service_types.UserState) error {
	_, err := c.doRequest(http.MethodPut, fmt.Sprintf("%s/%s/state/%s", c.Url, id, state))
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) UpdateLastSeenForUser(id string, t time.Time) error {
	_, err := c.doRequest(http.MethodPut, fmt.Sprintf("%s/%s/lastSeen/%d", c.Url, id, t.Unix()))
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) UpdateInstanceForUser(id string, instance string) error {
	_, err := c.doRequest(http.MethodPut, fmt.Sprintf("%s/%s/instance/%s", c.Url, id, instance))
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) doRequest(method, url string) ([]byte, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", c.ApiKey)
	resp, err := c.c.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, ServerErr
	}

	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}
