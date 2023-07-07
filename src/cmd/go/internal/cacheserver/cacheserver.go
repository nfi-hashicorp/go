// TODO: this feels more like a go tool cacheserver, but I want to access internal/cache
package cacheserver

import (
	"bytes"
	"cmd/go/internal/base"
	"cmd/go/internal/cache"
	"context"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

// Break init loop.
func init() {
	CmdCacheserver.Run = runCacheserver
}

var CmdCacheserver = &base.Command{
	UsageLine: `go cacheserver [addr (default: ":2601")]`,
	Short:     "runs build cache server",
}

func runCacheserver(ctx context.Context, cmd *base.Command, args []string) {
	addr := ":2601"
	if len(args) > 0 {
		addr = args[0]
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(r.URL.Path, "/")
		pb, err := hex.DecodeString(p)
		if len(pb) != cache.HashSize {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "path must be exactly %d bytes long; is %d", cache.HashSize, len(pb))
			return
		}
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "unabled to decode path %q to action ID: %s", p, err)
			return
		}
		aid := cache.ActionID(pb)

		switch r.Method {
		case http.MethodGet:
			// TODO: this should "only be used for data that can be expected to fit in memory",
			// so we probably shouldn't use it with aid coming from the client
			b, entry, err := cache.Default().GetBytes(aid)
			if err != nil {
				// TODO: clean this error?
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "error: unable to get item %q: %s", p, err)
				log.Printf("GET %q -> miss: %s", p, err)
				return
			}
			log.Printf("GET %q -> hit", p)
			http.ServeContent(w, r, p, entry.Time, bytes.NewReader(b))
		case http.MethodPost:
			b, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, "error: reading body: %s", err)
				return
			}
			out, size, err := cache.Default().PutNoVerify(aid, bytes.NewReader(b))
			if err != nil {
				log.Printf("POST %q -> error: PutNoVerify: %s", p, err)
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "error: PutNoVerify: %s", err)
				return
			}
			// TODO: return outputID?
			_ = size
			log.Printf("POST %q -> %x", p, out)
		}
	})
	log.Printf("serving on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
