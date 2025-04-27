package wire

import (
<<<<<<< HEAD
	"testing"

	"github.com/quic-go/quic-go/internal/protocol"

	"github.com/stretchr/testify/require"
)

func TestWritePingFrame(t *testing.T) {
	frame := PingFrame{}
	b, err := frame.Append(nil, protocol.Version1)
	require.NoError(t, err)
	require.Equal(t, []byte{0x1}, b)
	require.Len(t, b, int(frame.Length(protocol.Version1)))
}
=======
	"bytes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/project-faster/mp-quic-go/internal/protocol"
)

var _ = Describe("PingFrame", func() {
	Context("when parsing", func() {
		It("accepts sample frame", func() {
			b := bytes.NewReader([]byte{0x07})
			_, err := ParsePingFrame(b, protocol.VersionWhatever)
			Expect(err).ToNot(HaveOccurred())
			Expect(b.Len()).To(BeZero())
		})

		It("errors on EOFs", func() {
			_, err := ParsePingFrame(bytes.NewReader(nil), protocol.VersionWhatever)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when writing", func() {
		It("writes a sample frame", func() {
			b := &bytes.Buffer{}
			frame := PingFrame{}
			frame.Write(b, protocol.VersionWhatever)
			Expect(b.Bytes()).To(Equal([]byte{0x07}))
		})

		It("has the correct min length", func() {
			frame := PingFrame{}
			Expect(frame.MinLength(0)).To(Equal(protocol.ByteCount(1)))
		})
	})
})
>>>>>>> project-faster/main
