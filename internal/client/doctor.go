package client

import (
	"errors"
	"net/http"
	"time"
)

var ErrDocNotFound = errors.New("doctor not found")

type DoctorClient struct {
	BaseUrl string
	client  *http.Client
}

func NewDoctorClient(baseUrl string) *DoctorClient {
	return &DoctorClient{
		BaseUrl: baseUrl,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *DoctorClient) ExistsDoctor(id string) (bool, error) {
	resp, err := c.client.Get(c.BaseUrl + "/doctors/" + id)

	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	}

	return false, ErrDocNotFound
}
