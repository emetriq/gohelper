package net

import (
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"golang.org/x/net/proxy"
)

func CheckSocks5Proxy(proxyAddr string, testUrl string) error {
	// create a socks5 dialer
	dialer, err := proxy.SOCKS5("tcp", proxyAddr, nil, proxy.Direct)
	if err != nil {
		return err
	}
	// setup a http client
	httpTransport := &http.Transport{
		Dial: dialer.Dial,
	}
	httpClient := &http.Client{Transport: httpTransport}
	// create a request
	req, err := http.NewRequest("GET", testUrl, nil)
	if err != nil {
		return err
	}
	// use the http client to fetch the page
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return nil
}

func WaitForSocks5Proxy(proxUrl string, testUrl string, waitTimeSeconds int, maxCheckCount int) error {
	for i := 0; i < maxCheckCount; i++ {
		err := CheckSocks5Proxy(proxUrl, testUrl)
		time.Sleep(time.Second * time.Duration(waitTimeSeconds))
		if err == nil {
			return nil
		}
	}
	return errors.New("proxy check failed")
}
