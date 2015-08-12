package streamer

import (
	"io"
	"net/http"
)

type HttpHandler func(StreamID, io.Writer)

func (h HttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := StreamID(r.FormValue(":streamid"))
	w.WriteHeader(http.StatusOK)

	conn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer conn.Close()
	h(id, conn)
}
