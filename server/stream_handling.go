package server

import (
	"io"
	"net/http"
	"strconv"

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
	streamID, err := strconv.Atoi(r.FormValue(":streamid"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return 0, nil
	}
	stream := streamer.StreamID(streamID)

	w.WriteHeader(http.StatusOK)

	conn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return 0, nil
	}

	return stream, conn
}
