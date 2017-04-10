package service

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
)

const (
	maxHooks         = 4
	maxHookURLLength = 256
)

var (
	validHookURLSchemes = map[string]bool{
		"http":  true,
		"https": true,
	}
)

func (c *chain) postHookHandler(w http.ResponseWriter, r *http.Request) {

	if !contentTypeHeaderFound(w, r) {
		return
	}

	pl := struct {
		URL string `json:"url"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&pl); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(pl.URL) > maxHookURLLength {
		http.Error(w, "url too long", http.StatusBadRequest)
		return
	}

	url, err := url.ParseRequestURI(pl.URL)
	if err != nil {
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
	}

	if len(c.hooks) >= maxHooks {
		http.Error(w, "max hooks reached", http.StatusBadRequest)
		return
	}

	if !validHookURLSchemes[url.Scheme] {
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
	}

	if url.Host == "" {
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
	}

	if err := c.CreateHook(pl.URL); err == errHookExists {
		sendError(w, http.StatusBadRequest, "hook exists")
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (c *chain) getHooksHandler(w http.ResponseWriter, r *http.Request) {
	if !acceptHeaderFound(w, r) {
		return
	}
	hooks := c.Hooks()
	pl := make([]struct {
		URL string `json:"url"`
	}, len(hooks))
	for i, h := range hooks {
		pl[i] = struct {
			URL string `json:"url"`
		}{
			URL: h,
		}
	}
	sendPayload(w, http.StatusOK, "hooks", "", pl)
}

func (c *chain) deleteHookHandler(w http.ResponseWriter, r *http.Request) {

	encodedURL := mux.Vars(r)["url"]
	if encodedURL == "" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	urlBytes, err := base64.URLEncoding.DecodeString(encodedURL)
	if err != nil {
		http.Error(w, "url not base64 encoded", http.StatusBadRequest)
		return
	}
	url := string(urlBytes)
	c.DeleteHook(url)
}
