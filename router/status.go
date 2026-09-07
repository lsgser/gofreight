package router

import "net/http"

// Status forces a fixed HTTP status code for this route.
func (reg *RouteRegistrar) Status(code int) *RouteRegistrar {
	return reg.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(&forcedStatusWriter{ResponseWriter: w, code: code}, r)
		})
	})
}

// StatusName forces a fixed HTTP status using a symbolic name (e.g. "no_content").
func (reg *RouteRegistrar) StatusName(name string) *RouteRegistrar {
	code, ok := statusFromName(name)
	if !ok {
		return reg
	}
	return reg.Status(code)
}

type forcedStatusWriter struct {
	http.ResponseWriter
	code  int
	wrote bool
}

func (w *forcedStatusWriter) WriteHeader(status int) {
	w.wrote = true
	w.ResponseWriter.WriteHeader(w.code)
}

func (w *forcedStatusWriter) Write(b []byte) (int, error) {
	if !w.wrote {
		w.WriteHeader(w.code)
	}
	return w.ResponseWriter.Write(b)
}

func statusFromName(name string) (int, bool) {
	switch name {
	case "ok":
		return http.StatusOK, true
	case "created":
		return http.StatusCreated, true
	case "accepted":
		return http.StatusAccepted, true
	case "no_content", "no-content":
		return http.StatusNoContent, true
	case "not_found", "not-found":
		return http.StatusNotFound, true
	case "gone":
		return http.StatusGone, true
	case "found":
		return http.StatusFound, true
	case "see_other", "see-other":
		return http.StatusSeeOther, true
	case "moved_permanently", "moved-permanently":
		return http.StatusMovedPermanently, true
	default:
		return 0, false
	}
}
