// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package test

import (
	"cmd/go/internal/cache"
	"cmd/go/internal/lockedfile"
	"io"
	"path/filepath"
	"strconv"
	"time"
)

// fsCache wraps cache.Cache lightly, and adds ExpireAll
type fsCache struct {
}

var _ cacheI = &fsCache{}

func (c *fsCache) GetBytes(id cache.ActionID) ([]byte, cache.Entry, error) {
	return cache.Default().GetBytes(id)
}
func (c *fsCache) PutNoVerify(id cache.ActionID, file io.ReadSeeker) (cache.OutputID, int64, error) {
	return cache.Default().PutNoVerify(id, file)
}
func (c *fsCache) FuzzDir() string {
	return cache.Default().FuzzDir()
}

func (c *fsCache) ExpireAll() {
	// Read testcache expiration time, if present.
	// (We implement go clean -testcache by writing an expiration date
	// instead of searching out and deleting test result cache entries.)
	// TODO: skip if remotecache
	if dir := cache.DefaultDir(); dir != "off" {
		if data, _ := lockedfile.Read(filepath.Join(dir, "testexpire.txt")); len(data) > 0 && data[len(data)-1] == '\n' {
			if t, err := strconv.ParseInt(string(data[:len(data)-1]), 10, 64); err == nil {
				testCacheExpire = time.Unix(0, t)
			}
		}
	}
}
