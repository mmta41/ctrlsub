package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

type Client struct {
	*http.Client
	isInitialized bool
}

var pool = sync.Pool{
	New: func() interface{} {
		return &Client{&http.Client{}, false}
	},
}

func GetClient(timeout time.Duration) *Client {
	c := pool.Get().(*Client)
	if !c.isInitialized {
		t := http.DefaultTransport.(*http.Transport).Clone()
		t.MaxIdleConns = 100
		t.MaxConnsPerHost = 100
		t.MaxIdleConnsPerHost = 100
		t.TLSClientConfig.InsecureSkipVerify = true
		c.Transport = t
		c.isInitialized = true
	}
	c.Timeout = timeout
	return c
}

func ReleaseClient(client *Client) {
	pool.Put(client)
}

func Request(target string, timeout time.Duration, https bool, fallback bool) (resp *http.Response, body []byte, err error) {
	var url string
	if https == true {
		url = fmt.Sprintf("https://%s/", target)
	} else {
		url = fmt.Sprintf("http://%s/", target)
	}

	c := GetClient(timeout)
	defer ReleaseClient(c)

	var req *http.Request
	req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:60.0) Gecko/20100101 Firefox/60.0")

	resp, err = c.Do(req)
	if err != nil {
		if fallback {
			return Request(target, timeout, !https, false)
		}
		return nil, nil, err
	}

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}
	return resp, body, nil
}
