package utils

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/proxy"
)

//MyClient is http client
type MyClient struct {
	client *http.Client
}

//ErrRetry - when client rich maximum  retries
var ErrRetry = errors.New("Maxium retry reached")

//CreateClient creates new http client
func CreateClient() *MyClient {
	var m MyClient
	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 260 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 260 * time.Second,
	}
	var netClient = &http.Client{
		Timeout:   time.Second * 300,
		Transport: netTransport,
	}
	m.client = netClient
	return &m
}

//CreateClientWithProxy creates new http client with SOCKS5 proxy
func CreateClientWithProxy(proxyIP string) *MyClient {
	var m MyClient

	dialSocksProxy, err := proxy.SOCKS5("tcp", proxyIP, nil, proxy.Direct)
	if err != nil {
		fmt.Println("Error connecting to proxy:", err)
	}

	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 60 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 60 * time.Second,
	}

	netTransport.Dial = dialSocksProxy.Dial

	var netClient = &http.Client{
		Timeout:   time.Second * 60,
		Transport: netTransport,
	}
	m.client = netClient
	return &m
}

//Get is Get request
func (m *MyClient) Get(link string) (doc *goquery.Document, err error) {
	counter := 0
	for {
		if counter > 2 {
			return doc, fmt.Errorf("%s: %w", link, ErrRetry)
		}
		counter++
		resp, err := m.client.Get(link)
		if err != nil {
			log.Println(err)
			time.Sleep(10 * time.Second)
			continue
		}

		doc, err = goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Println(err)

			time.Sleep(10 * time.Second)
			resp.Body.Close()
			continue
		}
		resp.Body.Close()
		break
	}
	return

}

// GetWithByte returns Goquery document and raw byte
func (m *MyClient) GetWithByte(link string) (doc *goquery.Document, b []byte, err error) {
	counter := 0
	var resp *http.Response
	for {
		if counter > 2 {
			return doc, []byte{}, fmt.Errorf("%s: %w", link, ErrRetry)
		}
		counter++
		resp, err = m.client.Get(link)
		if err != nil {
			log.Println(err)
			time.Sleep(10 * time.Second)
			continue
		}

		/*doc, err = goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Println(err)

			time.Sleep(10 * time.Second)
			resp.Body.Close()
			continue
		}*/
		b, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return
		}
		resp.Body.Close()
		doc, _ = goquery.NewDocumentFromReader(bytes.NewReader(b))
		break
	}
	return

}

//CreateClientWithTOR creates new http client with SOCKS5 TOR proxy
func CreateClientWithTOR() *MyClient {
	var m MyClient

	dialSocksProxy, err := proxy.SOCKS5("tcp", "127.0.0.1:9150", nil, proxy.Direct)
	if err != nil {
		fmt.Println("Error connecting to proxy:", err)
	}

	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 60 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 60 * time.Second,
	}

	netTransport.Dial = dialSocksProxy.Dial

	var netClient = &http.Client{
		Timeout:   time.Second * 60,
		Transport: netTransport,
	}
	m.client = netClient
	return &m
}
