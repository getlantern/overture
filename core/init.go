// Copyright (c) 2016 shawn1m. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be
// found in the LICENSE file.

// Package core implements the essential features.
package core

import (
	"github.com/getlantern/overture/core/inbound"
	"github.com/getlantern/overture/core/outbound"
)

// Initiate the server with config file
func InitServer(configFilePath, socksAddr string) {

	config := NewConfig(configFilePath)

	config.PrimaryDNS[0].Protocol = "tcp"
	config.PrimaryDNS[0].SOCKS5Address = socksAddr
	config.AlternativeDNS[0].SOCKS5Address = socksAddr
	log.Debugf("Primary DNS : %q", config.PrimaryDNS)

	// New dispatcher without ClientBundle, ClientBundle must be initiated when server is running
	d := &outbound.Dispatcher{
		PrimaryDNS:         config.PrimaryDNS,
		AlternativeDNS:     config.AlternativeDNS,
		OnlyPrimaryDNS:     config.OnlyPrimaryDNS,
		IPNetworkList:      config.IPNetworkList,
		DomainList:         config.DomainList,
		RedirectIPv6Record: config.RedirectIPv6Record,
	}

	s := &inbound.Server{
		BindAddress: config.BindAddress,
		Dispatcher:  d,
		MinimumTTL:  config.MinimumTTL,
		RejectQtype: config.RejectQtype,
		//Hosts:       "127.0.0.1 localhost",
		Hosts: config.Hosts,
		Cache: config.Cache,
	}

	s.Run()
}
