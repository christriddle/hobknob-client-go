package hobknob

import (
	"strings"
	"time"

	"github.com/coreos/go-etcd/etcd"
)

// Client : the client
type Client struct {
	CacheInterval time.Duration
	AppName       string
	OnUpdate      chan error
	etcd          *etcd.Client
	cache         map[string]bool
}

// NewClient creates a new instance of the client and returns it
func NewClient(etcdHosts []string, appName string, cacheInterval int) *Client {
	client := &Client{
		cache:         make(map[string]bool),
		etcd:          etcd.NewClient(etcdHosts),
		CacheInterval: time.Duration(cacheInterval) * time.Second,
		AppName:       appName,
		OnUpdate:      make(chan error),
	}

	return client
}

// Initialise the client
func (c *Client) Initialise() error {
	err := c.update()
	if err == nil {
		c.schedule()
	}
	return err
}

func (c *Client) schedule() {
	ticker := time.NewTicker(c.CacheInterval)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				c.OnUpdate <- c.update()
				return
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func parseValue(val string) bool {
	if val == "true" {
		return true
	}

	return false
}

func parseResponse(resp *etcd.Response) map[string]bool {
	m := make(map[string]bool)
	for _, element := range resp.Node.Nodes {
		ks := strings.Split(element.Key, "/")
		m[ks[len(ks)-1]] = parseValue(element.Value)
	}
	return m
}

func (c *Client) update() error {
	resp, err := c.etcd.Get("/v1/toggles/"+c.AppName, false, true)
	if err != nil {
		return err
	}

	c.cache = parseResponse(resp)
	return nil
}

// Get a toggle state from the cache
func (c *Client) Get(toggle string) (bool, bool) {
	val, ok := c.cache[toggle]
	return val, ok
}

// GetOrDefault get a toggle and supply a default value
func (c *Client) GetOrDefault(toggle string, defaultVal bool) bool {
	if val, ok := c.cache[toggle]; ok {
		return val
	}
	return defaultVal
}
