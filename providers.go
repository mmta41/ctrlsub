package main

import (
	"encoding/json"
	"io/ioutil"
	"strings"
)

type Provider struct {
	Name     string   `json:"name"`
	Cname    []string `json:"cname"`
	Response []string `json:"response"`
}

var providers []*Provider

func InitializeProviders() error {
	raw, err := ioutil.ReadFile(config.Provider)
	if err != nil {
		return err
	}

	providers = make([]*Provider, 0, 50)
	err = json.Unmarshal(raw, &providers)
	if err != nil {
		return err
	}
	return nil
}

func FindProvider(target string) *Provider {
	for _, p := range providers {
		for _, cname := range p.Cname {
			if strings.Contains(target, cname) {
				return p
			}
		}
	}
	return nil
}
