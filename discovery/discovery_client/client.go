// Package discovery_client allows for communication with the Discovery service.
// TODO: Migrate away from net/http; Investigate the usage of fasthttp.

package discovery_client

import (
	"encoding/json"
	"fmt"
	"gitlab.com/george/shoya-go/models"
	"io"
	"net/http"
)

type Discovery struct {
	c      *http.Client
	Url    string
	ApiKey string
}

func NewDiscovery(url, apiKey string) *Discovery {
	return &Discovery{
		c:      &http.Client{},
		Url:    url,
		ApiKey: apiKey,
	}
}

// GetInstance retrieves live information about an instance.
func (d *Discovery) GetInstance(instance string) *models.WorldInstance {
	var i *models.WorldInstance

	b, err := d.doRequest(http.MethodGet, fmt.Sprintf("%s/%s", d.Url, instance))
	err = json.Unmarshal(b, i)
	if err != nil {
		return nil
	}

	return i
}

// GetInstancesForWorld retrieves a list of instances for a specified world id.
func (d *Discovery) GetInstancesForWorld(world string) []*models.WorldInstance {
	var i []*models.WorldInstance

	b, err := d.doRequest(http.MethodGet, fmt.Sprintf("%s/world/%s", d.Url, world))
	err = json.Unmarshal(b, &i)
	if err != nil {
		return nil
	}

	return i
}

// RegisterInstance registers an instance in Redis.
func (d *Discovery) RegisterInstance(instance string) {
	d.c.Post(fmt.Sprintf("%s/register/%s", d.Url, instance), "application/json", nil)
}

// UnregisterInstance removes an instance from Redis.
func (d *Discovery) UnregisterInstance(instance string) {
	d.c.Post(fmt.Sprintf("%s/unregister/%s", d.Url, instance), "application/json", nil)
}

// FindPlayer finds what instance(s) a player is in.
func (d *Discovery) FindPlayer(player string) []*models.WorldInstance {
	var i []*models.WorldInstance

	b, err := d.doRequest(http.MethodGet, fmt.Sprintf("%s/player/%s", d.Url, player))
	err = json.Unmarshal(b, &i)
	if err != nil {
		return nil
	}

	return i
}

// AddPlayerToInstance adds a player to an instance in Redis.
func (d *Discovery) AddPlayerToInstance(player, instance string) {
	d.doRequest(http.MethodPut, fmt.Sprintf("%s/player/%s/%s", d.Url, instance, player))
}

// RemovePlayerFromInstance removes a player from an instance in Redis.
func (d *Discovery) RemovePlayerFromInstance(player, instance string) {
	d.doRequest(http.MethodDelete, fmt.Sprintf("%s/player/%s/%s", d.Url, instance, player))
}

func (d *Discovery) doRequest(method, url string) ([]byte, error) {
	r, _ := http.NewRequest(method, url, nil)
	r.Header.Add("Authorization", d.ApiKey)

	do, err := d.c.Do(r)
	if err != nil {
		return nil, err
	}

	if do.StatusCode != 200 {
		return nil, err
	}

	b, err := io.ReadAll(do.Body)
	if err != nil {
		return nil, err
	}

	return b, nil
}
