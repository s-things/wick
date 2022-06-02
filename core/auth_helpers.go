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
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"github.com/gammazero/nexus/v3/wamp"
	"github.com/gammazero/nexus/v3/wamp/crsign"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/pbkdf2"
)

func handleCRAAuth(secret string) func(c *wamp.Challenge) (string, wamp.Dict) {
	callable := func(c *wamp.Challenge) (string, wamp.Dict) {
		ch, _ := wamp.AsString(c.Extra["challenge"])
		// If the client needed to lookup a user's key, this would require decoding
		// the JSON-encoded challenge string and getting the authid.  For this
		// example assume that client only operates as one user and knows the key
		// to use.

		var rawSecret []byte
		saltStr, _ := wamp.AsString(c.Extra["salt"])
		// If no salt given, use raw password as key.
		if saltStr != "" {
			// If salting info give, then compute a derived key using PBKDF2.
			iters, _ := wamp.AsInt64(c.Extra["iterations"])
			keylen, _ := wamp.AsInt64(c.Extra["keylen"])

			rawSecret = deriveKey(saltStr, secret, int(iters), int(keylen))
		} else {
			rawSecret = []byte(secret)
		}

		return crsign.SignChallenge(ch, rawSecret), wamp.Dict{}
	}

	return callable
}

func deriveKey(saltStr string, secret string, iterations int, keyLength int) []byte {
	// If salting info give, then compute a derived key using PBKDF2.
	salt := []byte(saltStr)
	//iters, _ := wamp.AsInt64(c.Extra["iterations"])
	//keylen, _ := wamp.AsInt64(c.Extra["keylen"])

	if iterations == 0 {
		iterations = 1000
	}
	if keyLength == 0 {
		keyLength = 32
	}

	// Compute derived key.
	dk := pbkdf2.Key([]byte(secret), salt, iterations, keyLength, sha256.New)
	// Get base64 bytes. see https://github.com/gammazero/nexus/issues/252
	derivedKey := []byte(base64.StdEncoding.EncodeToString(dk))

	return derivedKey
}

func handleCryptosign(pvk ed25519.PrivateKey) func(c *wamp.Challenge) (string, wamp.Dict) {
	callable := func(c *wamp.Challenge) (string, wamp.Dict) {
		challengeHex, _ := wamp.AsString(c.Extra["challenge"])
		challengeBytes, _ := hex.DecodeString(challengeHex)

		signed := ed25519.Sign(pvk, challengeBytes)
		signedHex := hex.EncodeToString(signed)
		result := signedHex + challengeHex
		return result, wamp.Dict{}
	}

	return callable
}

func getBaseHello(authid string, authrole string) wamp.Dict {
	helloDict := wamp.Dict{}
	if authid != "" {
		helloDict["authid"] = authid
	}

	if authrole != "" {
		helloDict["authrole"] = authrole
	}

	return helloDict
}
