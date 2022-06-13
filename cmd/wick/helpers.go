/*
*
* Copyright 2021-2022 Simple Things Inc.
*
* Permission is hereby granted, free of charge, to any person obtaining a copy
* of this software and associated documentation files (the "Software"), to deal
* in the Software without restriction, including without limitation the rights
* to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
* copies of the Software, and to permit persons to whom the Software is
* furnished to do so, subject to the following conditions:
*
* The above copyright notice and this permission notice shall be included in all
* copies or substantial portions of the Software.
*
* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
* IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
* FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
* LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
* OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
* SOFTWARE.
*
 */

package main

import (
	"fmt"
	"github.com/gammazero/nexus/v3/transport/serialize"
	"github.com/sirupsen/logrus"
	"gopkg.in/ini.v1"
	"os"
	"runtime"
)

func getSerializerByName(name string) serialize.Serialization {

	switch name {
	case "json":
		return serialize.JSON
	case "msgpack":
		return serialize.MSGPACK
	case "cbor":
		return serialize.CBOR
	}
	return -1
}

func selectAuthMethod(privateKey string, ticket string, secret string) string {
	if privateKey != "" && (ticket == "" && secret == "") {
		return "cryptosign"
	} else if ticket != "" && (privateKey == "" && secret == "") {
		return "ticket"
	} else if secret != "" && (privateKey == "" && ticket == "") {
		return "wampcra"
	}

	return "anonymous"
}

func userHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	} else if runtime.GOOS == "linux" {
		home := os.Getenv("XDG_CONFIG_HOME")
		if home != "" {
			return home
		}
	}
	return os.Getenv("HOME")
}

func readFromProfile(logger *logrus.Logger) {
	cfg, err := ini.Load(fmt.Sprintf("%s/.wick/config", userHomeDir()))
	if err != nil {
		logger.Fatalf("Fail to read config: %v", err)
	}

	section, err := cfg.GetSection(*profile)
	if err != nil {
		logger.Fatalf("Error in getting section: %s", err)
	}

	*url = section.Key("url").Validate(func(s string) string {
		if len(s) == 0 {
			return "ws://localhost:8080/ws"
		}
		return s
	})
	*realm = section.Key("realm").Validate(func(s string) string {
		if len(s) == 0 {
			return "realm1"
		}
		return s
	})
	*authid = section.Key("authid").String()
	*authrole = section.Key("authrole").String()
	*authMethod = section.Key("authmethod").String()
	if *authMethod == "cryptosign" {
		*privateKey = section.Key("private-key").String()
	} else if *authMethod == "ticket" {
		*ticket = section.Key("ticket").String()
	} else if *authMethod == "wampcra" {
		*secret = section.Key("secret").String()
	}
}
