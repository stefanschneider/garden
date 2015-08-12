package streamer_test

import (
	"bytes"
	"errors"
	"io"
	"sync"
	"time"

	"runtime"

	"github.com/cloudfoundry-incubator/garden/server/streamer"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Streamer", func() {

	const testString = "x"

	var (
		graceTime time.Duration
		str       *streamer.Streamer

		stdoutChan        chan []byte
		stderrChan        chan []byte
		testByteSlice     []byte
		channelBufferSize int
	)

	JustBeforeEach(func() {
		str = streamer.New(graceTime)
		stdoutChan = make(chan []byte, channelBufferSize)
		stderrChan = make(chan []byte, channelBufferSize)
	})

	BeforeEach(func() {
		graceTime = 10 * time.Second
		channelBufferSize = 1

		testByteSlice = []byte(testString)
	})

	It("should stream standard output until it is stopped", func() {
		sid := str.Stream(stdoutChan, stderrChan)
		w := &syncBuffer{
			Buffer: new(bytes.Buffer),
		}
		go str.ServeStdout(sid, w)
		stdoutChan <- testByteSlice
		stdoutChan <- testByteSlice
		Eventually(w.String).Should(Equal("xx"))
		str.Stop(sid)
	})

	It("should stream just one standard output message after being stopped", func() {
		sid := str.Stream(stdoutChan, stderrChan)
		str.Stop(sid)
		w := new(bytes.Buffer)
		stdoutChan <- testByteSlice
		str.ServeStdout(sid, w)
		stdoutChan <- testByteSlice
		Consistently(w.String).Should(Equal(testString))
	})

	It("should return and not panic when asked to stream output with an invalid stream ID", func() {
		w := new(bytes.Buffer)
		str.ServeStdout("", w)
	})

	It("should stream standard error until it is stopped", func() {
		sid := str.Stream(stdoutChan, stderrChan)
		w := &syncBuffer{
			Buffer: new(bytes.Buffer),
		}
		go str.ServeStderr(sid, w)
		stderrChan <- testByteSlice
		stderrChan <- testByteSlice
		Eventually(w.String).Should(Equal("xx"))
		str.Stop(sid)
	})

	It("should stream just one standard error message after being stopped", func() {
		sid := str.Stream(stdoutChan, stderrChan)
		str.Stop(sid)
		w := new(bytes.Buffer)
		stderrChan <- testByteSlice
		str.ServeStderr(sid, w)
		stderrChan <- testByteSlice
		Consistently(w.String).Should(Equal(testString))
	})

	It("should return and not panic when asked to stream errors with an invalid stream ID", func() {
		w := new(bytes.Buffer)
		str.ServeStderr("", w)
	})

	It("should return and not panic when asked to stream errors with a nil writer", func() {
		var w io.Writer
		sid := str.Stream(stdoutChan, stderrChan)
		str.Stop(sid)
		stdoutChan <- testByteSlice
		str.ServeStderr(sid, w)
	})

	It("should panic when asked to stop with an invalid stream ID", func() {
		Expect(func() { str.Stop("") }).To(Panic())
	})

	Context("when using channels with buffer size greater than one", func() {
		BeforeEach(func() {
			channelBufferSize = 2
		})

		It("should finish streaming standard output after being stopped", func() {
			sid := str.Stream(stdoutChan, stderrChan)
			str.Stop(sid)
			w := new(bytes.Buffer)
			stdoutChan <- testByteSlice
			stdoutChan <- testByteSlice
			str.ServeStdout(sid, w)
			Consistently(w.String).Should(Equal("xx"))
		})

		It("should finish streaming standard error after being stopped", func() {
			sid := str.Stream(stdoutChan, stderrChan)
			str.Stop(sid)
			w := new(bytes.Buffer)
			stderrChan <- testByteSlice
			stderrChan <- testByteSlice
			str.ServeStderr(sid, w)
			Consistently(w.String).Should(Equal("xx"))
		})
	})

	Context("when a grace time has been set", func() {
		BeforeEach(func() {
			graceTime = 100 * time.Millisecond

		})

		It("should not leak unused streams for longer than the grace time after streaming has been stopped", func() {
			sid := str.Stream(stdoutChan, stderrChan)
			str.Stop(sid)
			time.Sleep(200 * time.Millisecond)
			Expect(func() { str.Stop(sid) }).To(Panic(), "stream was not removed")
		})

		It("should not leak goroutines for longer than the grace time after streaming has been stopped", func() {
			initialNumGoroutine := runtime.NumGoroutine()

			sid := str.Stream(stdoutChan, stderrChan)
			str.Stop(sid)

			Eventually(func() int {
				return runtime.NumGoroutine()
			}, "200ms").Should(Equal(initialNumGoroutine))
		})
	})

	It("should terminate streaming output after a write error has occurred", func() {
		sid := str.Stream(stdoutChan, stderrChan)
		w := &syncBuffer{
			Buffer: new(bytes.Buffer),
			fail:   true,
		}
		go str.ServeStdout(sid, w)
		stdoutChan <- testByteSlice
		stdoutChan <- testByteSlice
		Consistently(w.String).Should(Equal(""))
		str.Stop(sid)
	})

	It("should terminate streaming errors after a write error has occurred", func() {
		sid := str.Stream(stdoutChan, stderrChan)
		w := &syncBuffer{
			Buffer: new(bytes.Buffer),
			fail:   true,
		}
		go str.ServeStderr(sid, w)
		stderrChan <- testByteSlice
		stderrChan <- testByteSlice
		Consistently(w.String).Should(Equal(""))
		str.Stop(sid)
	})
})

type syncBuffer struct {
	*bytes.Buffer
	fail bool
	mu   sync.Mutex
}

func (sb *syncBuffer) Write(p []byte) (int, error) {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	if sb.fail {
		sb.fail = false
		return 0, errors.New("failed")
	}
	return sb.Buffer.Write(p)
}

func (sb *syncBuffer) String() string {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return sb.Buffer.String()
}
