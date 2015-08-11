package server

import (
	"io"
	"net/http"

	"github.com/cloudfoundry-incubator/garden/server/streamer"
)

func (s *GardenServer) handleStdout(w http.ResponseWriter, r *http.Request) {
	streamID, conn := streamInfo(w, r)
	if conn != nil {
		defer conn.Close()
		s.streamer.StreamStdout(streamID, conn)
	}
}

func (s *GardenServer) handleStderr(w http.ResponseWriter, r *http.Request) {
	streamID, conn := streamInfo(w, r)
	if conn != nil {
		defer conn.Close()
		s.streamer.StreamStderr(streamID, conn)
	}
}

func streamInfo(w http.ResponseWriter, r *http.Request) (streamer.StreamID, io.WriteCloser) {
	stream := streamer.StreamID(r.FormValue(":streamid"))
	w.WriteHeader(http.StatusOK)

	conn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return "", nil
	}

	return stream, conn
}
