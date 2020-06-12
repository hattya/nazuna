//
// nazuna/internal/test :: test.go
//
//   Copyright (c) 2013-2020 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

package test

import (
	"crypto/tls"
	"net"
	"net/http"
)

func NewHTTPClient(addr string) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Dial: func(network, _ string) (net.Conn, error) {
				return net.Dial(network, addr)
			},
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
}
