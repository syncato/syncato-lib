// Copyright 2015 The Syncato Authors.  All rights reserved.
// Use of this source code is governed by a AGPL
// license that can be found in the LICENSE file.

// Package mux defines the API multiplexer to route requests to the proper API.
package mux

import (
	"github.com/syncato/lib/api"
	"github.com/syncato/lib/logger"
	"golang.org/x/net/context"
	"net/http"
	"strings"
)

// APIMux is the multiplexer responsible for routing request to a specific API.
// It keeps a map with all the APIs.
type APIMux struct {
	apis map[string]api.APIProvider
	log  *logger.Logger
}

// NewAPIMux creates a new APIMux object or return an error
func NewAPIMux(logger *logger.Logger) (*APIMux, error) {
	apimux := &APIMux{map[string]api.APIProvider{}, logger}
	return apimux, nil
}

// RegisterAPI register an API into the APIMux so it can be used.
func (apimux *APIMux) RegisterApi(api api.APIProvider) {
	apimux.apis[api.GetID()] = api
}

// GetAPI returns an API object by its ID
func (apimux *APIMux) GetAPI(apiID string) api.APIProvider {
	return apimux.apis[apiID]
}

// HandleRequest routes a general request to the specific API or returns 404 if the API
// asked is not registered.
func (apimux *APIMux) HandleRequest(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	urlParts := strings.Split(path, "/")
	// a correct url will be /api/files/something, so the len of the urlParts should be at least 3
	if len(urlParts) < 3 {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	apiID := urlParts[2]
	api := apimux.GetAPI(apiID)
	if api == nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	api.HandleRequest(ctx, w, r)
}
