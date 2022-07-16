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
	"testing"

	"github.com/gammazero/nexus/v3/transport/serialize"
)

func TestSerializerSelect(t *testing.T) {
	serializerID := getSerializerByName("json")
	if serializerID != serialize.JSON {
		t.Errorf("invalid serializer id for json, expected=%d, got=%d", serialize.JSON, serializerID)
	}

	serializerID = getSerializerByName("cbor")
	if serializerID != serialize.CBOR {
		t.Errorf("invalid serializer id for cbor, expected=%d, got=%d", serialize.CBOR, serializerID)
	}

	serializerID = getSerializerByName("msgpack")
	if serializerID != serialize.MSGPACK {
		t.Errorf("invalid serializer id for msgpack, expected=%d, got=%d", serialize.MSGPACK, serializerID)
	}

	serializerID = getSerializerByName("halo")
	if serializerID != -1 {
		t.Errorf("should not accept as only anonymous,ticket,wampcra,cryptosign are allowed")
	}
}

func TestSelectCryptosignAuthMethod(t *testing.T) {
	method := selectAuthMethod("b99067e6e271ae300f3f5d9809fa09288e96f2bcef8dd54b7aabeb4e579d37ef", "", "")
	if method != "cryptosign" {
		t.Error("problem in choosing auth method")
	}
}

func TestSelectTicketAuthMethod(t *testing.T) {
	method := selectAuthMethod("", "williamsburg", "")
	if method != "ticket" {
		t.Error("problem in choosing auth method")
	}
}

func TestSelectWampCRAAuthMethod(t *testing.T) {
	method := selectAuthMethod("", "", "williamsburg")
	if method != "wampcra" {
		t.Error("problem in choosing auth method")
	}
}

func TestAutoSelectMethodAnony(t *testing.T) {
	method := selectAuthMethod("", "", "")
	if method != "anonymous" {
		t.Error("default authmethod must be anonymous if no credentials provided")
	}
}
