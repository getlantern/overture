// Copyright (c) 2016 shawn1m. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be
// found in the LICENSE file.

package core

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/getlantern/golog"
	"github.com/getlantern/overture/core/cache"
	"github.com/getlantern/overture/core/common"
	"github.com/getlantern/overture/core/hosts"
	"github.com/getlantern/overture/core/outbound"
)

var (
	log = golog.LoggerFor("overture.core")
)

type Config struct {
	BindAddress        string `json:"BindAddress"`
	PrimaryDNS         []*common.DNSUpstream
	AlternativeDNS     []*common.DNSUpstream
	OnlyPrimaryDNS     bool
	RedirectIPv6Record bool
	IPNetworkFile      string
	DomainFile         string
	DomainBase64Decode bool
	HostsFile          string
	MinimumTTL         int
	CacheSize          int
	RejectQtype        []uint16

	DomainList    []string
	IPNetworkList []*net.IPNet
	Hosts         *hosts.Hosts
	Cache         *cache.Cache
}

// New config with json file and do some other initiate works
func NewConfig(configFile string) *Config {

	log.Debugf("Trying to load overture config from %s", configFile)
	config := parseJson(configFile)

	config.getIPNetworkList()
	config.getDomainList()

	if config.MinimumTTL > 0 {
		log.Debugf("Minimum TTL is %s", strconv.Itoa(config.MinimumTTL))
	} else {
		log.Debug("Minimum TTL is disabled")
	}

	config.Cache = cache.New(config.CacheSize)
	if config.CacheSize > 0 {
		log.Debugf("CacheSize is %s", strconv.Itoa(config.CacheSize))
	} else {
		log.Debug("Cache is disabled")
	}

	h, err := hosts.New(config.HostsFile)
	if err != nil {
		log.Debugf("Load hosts file failed: %v", err)
	} else {
		config.Hosts = h
		log.Debug("Load hosts file successful")
	}

	return config
}

func parseFromString(configStr string) *Config {
	j := new(Config)
	err := json.Unmarshal([]byte(configStr), j)
	if err != nil {
		log.Fatalf("Json syntax error: %v", err)
		os.Exit(1)
	}

	log.Debug(configStr)

	return j
}

func parseJson(path string) *Config {

	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("Open config file failed: %v", err)
		os.Exit(1)
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatalf("Read config file failed: %v", err)
		os.Exit(1)
	}

	j := new(Config)
	err = json.Unmarshal(b, j)
	if err != nil {
		log.Fatalf("Json syntax error: %v", err)
		os.Exit(1)
	}

	log.Debug(string(b))

	return j
}

func (c *Config) getDomainList() {

	log.Debugf("Attempting to read domain list from file %s",
		c.DomainFile)

	var dl []string
	f, err := ioutil.ReadFile(c.DomainFile)
	if err != nil {
		log.Errorf("Open Custom domain file failed: %v", err)
		return
	}

	re := regexp.MustCompile(`([\w\-\_]+\.[\w\.\-\_]+)[\/\*]*`)
	if c.DomainBase64Decode {
		fd, err := base64.StdEncoding.DecodeString(string(f))
		if err != nil {
			log.Errorf("Decode Custom domain failed: %v", err)
			return
		}
		fds := string(fd)
		n := strings.Index(fds, "Whitelist Start")
		dl = re.FindAllString(fds[:n], -1)
	} else {
		dl = re.FindAllString(string(f), -1)
	}

	if len(dl) > 0 {
		log.Debug("Load domain file successful")
	} else {
		log.Debug("There is no element in domain file")
	}
	c.DomainList = dl
}

func (c *Config) getIPNetworkList() {

	ipnl := make([]*net.IPNet, 0)
	f, err := os.Open(c.IPNetworkFile)
	if err != nil {
		log.Errorf("Open IP network file failed: %v", err)
		return
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		_, ip_net, err := net.ParseCIDR(s.Text())
		if err != nil {
			break
		}
		ipnl = append(ipnl, ip_net)
	}
	if len(ipnl) > 0 {
		log.Debug("Load IP network file successful")
	} else {
		log.Debug("There is no element in IP network file")
	}

	c.IPNetworkList = ipnl
}
