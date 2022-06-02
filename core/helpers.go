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

package core

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gammazero/nexus/v3/wamp"
	"golang.org/x/crypto/ed25519"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

func listToWampList(args []string) wamp.List {
	var arguments wamp.List

	if args == nil {
		return wamp.List{}
	}

	for _, value := range args {

		var mapJson map[string]interface{}
		var mapList []map[string]interface{}

		if number, errNumber := strconv.Atoi(value); errNumber == nil {
			arguments = append(arguments, number)
		} else if float, errFloat := strconv.ParseFloat(value, 64); errFloat == nil {
			arguments = append(arguments, float)
		} else if boolean, errBoolean := strconv.ParseBool(value); errBoolean == nil {
			arguments = append(arguments, boolean)
		} else if errJson := json.Unmarshal([]byte(value), &mapJson); errJson == nil {
			arguments = append(arguments, mapJson)
		} else if errList := json.Unmarshal([]byte(value), &mapList); errList == nil {
			arguments = append(arguments, mapList)
		} else {
			arguments = append(arguments, value)
		}
	}

	return arguments
}

func dictToWampDict(kwargs map[string]string) wamp.Dict {
	var keywordArguments wamp.Dict = make(map[string]interface{})

	for key, value := range kwargs {

		var mapJson map[string]interface{}
		var mapList []map[string]interface{}

		if number, errNumber := strconv.Atoi(value); errNumber == nil {
			keywordArguments[key] = number
		} else if float, errFloat := strconv.ParseFloat(value, 64); errFloat == nil {
			keywordArguments[key] = float
		} else if boolean, errBoolean := strconv.ParseBool(value); errBoolean == nil {
			keywordArguments[key] = boolean
		} else if errJson := json.Unmarshal([]byte(value), &mapJson); errJson == nil {
			keywordArguments[key] = mapJson
		} else if errList := json.Unmarshal([]byte(value), &mapList); errList == nil {
			keywordArguments[key] = mapList
		} else {
			keywordArguments[key] = value
		}
	}
	return keywordArguments
}

func argsKWArgs(args wamp.List, kwArgs wamp.Dict, details wamp.Dict) {
	if details != nil {
		logger.Println(details)
	}

	if len(args) != 0 {
		fmt.Println("args:")
		jsonString, err := json.MarshalIndent(args, "", "    ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(jsonString))
	}

	if len(kwArgs) != 0 {
		fmt.Println("kwargs:")
		jsonString, err := json.MarshalIndent(kwArgs, "", "    ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(jsonString))
	}

	if len(args) == 0 && len(kwArgs) == 0 {
		fmt.Println("args: []")
		fmt.Println("kwargs: {}")
	}
}

func shellOut(command string) (error, string, string) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	var cmd *exec.Cmd
	cmd = exec.Command("bash", "-c", command)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return err, stdout.String(), stderr.String()
}

func getKeyPair(privateKeyKex string) (ed25519.PublicKey, ed25519.PrivateKey) {
	privateKeyRaw, _ := hex.DecodeString(privateKeyKex)
	var privateKey ed25519.PrivateKey

	if len(privateKeyRaw) == 32 {
		privateKey = ed25519.NewKeyFromSeed(privateKeyRaw)
	} else if len(privateKeyRaw) == 64 {
		privateKey = ed25519.NewKeyFromSeed(privateKeyRaw[:32])
	} else {
		logger.Fatal("Invalid private key. Cryptosign private key must be either 32 or 64 characters long")
	}

	publicKey := privateKey.Public().(ed25519.PublicKey)

	return publicKey, privateKey
}

func sanitizeURL(url string) string {
	if strings.HasPrefix(url, "rs") {
		return "tcp" + strings.TrimPrefix(url, "rs")
	} else if strings.HasPrefix(url, "rss") {
		return "tcp" + strings.TrimPrefix(url, "rss")
	}
	return url
}
