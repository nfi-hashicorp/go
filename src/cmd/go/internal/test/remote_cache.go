// go_bootstrap builds can't use net (or net/http)
//go:build !cmd_go_bootstrap

package test

import (
	"cmd/go/internal/cache"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

// A remoteCache is a remote cache minimially compatible with cache.Cache.
type remoteCache struct {
	base string
	c    *http.Client
}

// assert remoteCache implements cacheI
var _ cacheI = &remoteCache{}

func newRemote(url string) *remoteCache {
	return &remoteCache{
		base: url,
		c:    http.DefaultClient,
	}
}

// GetBytes looks up the action ID in the cache and returns
// the corresponding output bytes.
// GetBytes should only be used for data that can be expected to fit in memory.
func (c *remoteCache) GetBytes(id cache.ActionID) ([]byte, cache.Entry, error) {
	url := fmt.Sprintf("%s/%s", c.base, hex.EncodeToString([]byte(id[:])))
	resp, err := c.c.Get(url)
	if err != nil {
		return nil, cache.Entry{}, fmt.Errorf("GET %q: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, cache.Entry{}, fmt.Errorf("GET %q: %d: reading body: %s", url, resp.StatusCode, err)
		}
		return nil, cache.Entry{}, fmt.Errorf("GET %q: %d: %s", url, resp.StatusCode, body)
	}
	entry := cache.Entry{}
	// TODO: should it really be an error if this field is missing? do we need it?
	mtimeS := resp.Header.Get("Last-Modified")
	if mtimeS == "" {
		return nil, entry, errors.New("missing Last-Modified")
	}
	mtime, err := time.Parse(http.TimeFormat, mtimeS)
	if err != nil {
		return nil, entry, fmt.Errorf("parsing Last-Modified: %w", err)
	}
	entry.Time = mtime

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, entry, fmt.Errorf("reading body: %s", err)
	}
	// pretty sure http checks that Content-Length and body length align
	entry.Size = int64(len(body))
	// TODO: cache.OutputID
	return body, entry, nil
}

func (c *remoteCache) PutNoVerify(id cache.ActionID, file io.ReadSeeker) (cache.OutputID, int64, error) {
	url := fmt.Sprintf("%s/%s", c.base, hex.EncodeToString([]byte(id[:])))
	// TODO: content type?
	// TODO: maybe this should be PUT
	resp, err := c.c.Post(url, "", file)
	if err != nil {
		return cache.OutputID{}, 0, fmt.Errorf("POST %q: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		// TODO: limit
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return cache.OutputID{}, 0, fmt.Errorf("POST: %d: reading body : %s", resp.StatusCode, err)
		}
		return cache.OutputID{}, 0, fmt.Errorf("POST: %d: body : %q", resp.StatusCode, body)
	}
	// TODO: real cache.OutputID, real size
	return cache.OutputID{}, 0, nil
}

// Get looks up the action ID in the cache,
// returning the corresponding output ID and file size, if any.
// Note that finding an output ID does not guarantee that the
// saved file for that output ID is still available.
func (c *remoteCache) Get(id cache.ActionID) (cache.Entry, error) {
	panic("TODO?")
}

func (c *remoteCache) ExpireAll() {
	panic("TODO?")
}

// OutputFile returns the name of the cache file storing output with the given cache.OutputID.
func (c *remoteCache) OutputFile(out cache.OutputID) string {
	panic("TODO?")
}

// Trim removes old cache entries that are likely not to be reused.
func (c *remoteCache) Trim() {
	panic("TODO?")
}

// Put stores the given output in the cache as the output for the action ID.
// It may read file twice. The content of file must not change between the two passes.
func (c *remoteCache) Put(id cache.ActionID, file io.ReadSeeker) (cache.OutputID, int64, error) {
	panic("TODO?")
}

// PutBytes stores the given bytes in the cache as the output for the action ID.
func (c *remoteCache) PutBytes(id cache.ActionID, data []byte) error {
	panic("TODO")
}

// FuzzDir returns a subdirectory within the cache for storing fuzzing data.
// The subdirectory may not exist.
//
// This directory is managed by the internal/fuzz package. Files in this
// directory aren't removed by the 'go clean -cache' command or by Trim.
// They may be removed with 'go clean -fuzzcache'.
//
// TODO(#48526): make Trim remove unused files from this directory.
func (c *remoteCache) FuzzDir() string {
	panic("TODO?")
}

func getCache() cacheI {
	// TODO: Once, blah blah blah
	// TODO: actually get this from the env, etc.
	const remoteURL = "http://localhost:2601"
	return newRemote(remoteURL)
}
