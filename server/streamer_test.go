package server_test

import (
	"net/http"
	"net/url"
	"runtime/pprof"
	"time"

	"github.com/cloudfoundry-incubator/garden/server"
	"github.com/cloudfoundry-incubator/garden/server/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Streamer", func() {
	Context("when where is a single stream still running", func() {
		It("does not leak goroutines", func() {
			streamer := server.NewStreamServer(500 * time.Millisecond)

			initialGoRoutines := pprof.Lookup("goroutine").Count()

			w := new(fakes.FakeHijackerResponseWriter)
			w.HijackReturns(new(fakes.FakeConn), nil, nil)

			for i := 0; i < 100; i++ {
				stdout := make(chan []byte, 0)
				id := streamer.Stream(stdout, nil)

				go streamer.HandleStream(w,
					&http.Request{Form: url.Values{":streamid": []string{id}}},
					server.Stdout)

				if i == 0 { // simulate not every stream having stopped
					continue
				}

				streamer.Stop(id)
			}

			profile := pprof.Lookup("goroutine")
			Eventually(profile.Count).Should(BeNumerically("<", initialGoRoutines+20)) // leave some wiggle room for system goroutines
		})
	})
})
