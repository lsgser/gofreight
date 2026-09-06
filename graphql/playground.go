package graphql

import (
	_ "embed"
	"net/http"
)

//go:embed playground.html
var playgroundHTML []byte

func (s *Server) playgroundHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(playgroundHTML)
	}
}
