// Copyright 2015 The Syncato Authors.  All rights reserved.
// Use of this source code is governed by a AGPL
// license that can be found in the LICENSE file.

// Package api defines the API interface that every API should implement.
package api

import (
	"net/http"

	"golang.org/x/net/context"
)

// APIProvider is the interface that APIs should implement to be served from the daemon.
// An API is defined by an ID, so for example the APIFiles will have the ID 'files'.
type APIProvider interface {
	GetID() string                                                     // returns the ID of the API.
	HandleRequest(context.Context, http.ResponseWriter, *http.Request) // handle the requestd.
}
