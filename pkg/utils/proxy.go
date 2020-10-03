package utils

import (
	"io/ioutil"
	"log"
	"math/rand"
	"strings"
)

//MyProxy holds proxy
type MyProxy struct {
	Proxies []string
}

//NewProxySet returns pointer on MyProxy
//after intit
func NewProxySet(p string) *MyProxy {
	var mp MyProxy
	b, err := ioutil.ReadFile(p)
	if err != nil {
		log.Fatal(err)
	}
	mp.Proxies = strings.Split(string(b), "\n")
	return &mp
}

//GetRandom returns random proxy
func (mp *MyProxy) GetRandom() string {
	return strings.TrimSpace(mp.Proxies[rand.Intn(len(mp.Proxies)-1)])
}
