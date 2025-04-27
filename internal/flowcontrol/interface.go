package flowcontrol

<<<<<<< HEAD
import (
	"time"

	"github.com/quic-go/quic-go/internal/protocol"
)

type flowController interface {
	// for sending
	SendWindowSize() protocol.ByteCount
	UpdateSendWindow(protocol.ByteCount) (updated bool)
	AddBytesSent(protocol.ByteCount)
	// for receiving
	GetWindowUpdate(time.Time) protocol.ByteCount // returns 0 if no update is necessary
}

// A StreamFlowController is a flow controller for a QUIC stream.
type StreamFlowController interface {
	flowController
	AddBytesRead(protocol.ByteCount) (hasStreamWindowUpdate, hasConnWindowUpdate bool)
	// UpdateHighestReceived is called when a new highest offset is received
	// final has to be to true if this is the final offset of the stream,
	// as contained in a STREAM frame with FIN bit, and the RESET_STREAM frame
	UpdateHighestReceived(offset protocol.ByteCount, final bool, now time.Time) error
	// Abandon is called when reading from the stream is aborted early,
	// and there won't be any further calls to AddBytesRead.
	Abandon()
	IsNewlyBlocked() bool
}

// The ConnectionFlowController is the flow controller for the connection.
type ConnectionFlowController interface {
	flowController
	AddBytesRead(protocol.ByteCount) (hasWindowUpdate bool)
	Reset() error
	IsNewlyBlocked() (bool, protocol.ByteCount)
}

type connectionFlowControllerI interface {
	ConnectionFlowController
	// The following two methods are not supposed to be called from outside this packet, but are needed internally
	// for sending
	EnsureMinimumWindowSize(protocol.ByteCount, time.Time)
	// for receiving
	IncrementHighestReceived(protocol.ByteCount, time.Time) error
=======
import "github.com/project-faster/mp-quic-go/internal/protocol"

// WindowUpdate provides the data for WindowUpdateFrames.
type WindowUpdate struct {
	StreamID protocol.StreamID
	Offset   protocol.ByteCount
}

// A FlowControlManager manages the flow control
type FlowControlManager interface {
	NewStream(streamID protocol.StreamID, contributesToConnectionFlow bool)
	RemoveStream(streamID protocol.StreamID)
	// methods needed for receiving data
	ResetStream(streamID protocol.StreamID, byteOffset protocol.ByteCount) error
	UpdateHighestReceived(streamID protocol.StreamID, byteOffset protocol.ByteCount) error
	AddBytesRead(streamID protocol.StreamID, n protocol.ByteCount) error
	GetWindowUpdates(force bool) []WindowUpdate
	GetReceiveWindow(streamID protocol.StreamID) (protocol.ByteCount, error)
	// methods needed for sending data
	AddBytesSent(streamID protocol.StreamID, n protocol.ByteCount) error
	SendWindowSize(streamID protocol.StreamID) (protocol.ByteCount, error)
	RemainingConnectionWindowSize() protocol.ByteCount
	UpdateWindow(streamID protocol.StreamID, offset protocol.ByteCount) (bool, error)
	// methods useful to collect statistics
	GetBytesSent(streamID protocol.StreamID) (protocol.ByteCount, error)
	AddBytesRetrans(streamID protocol.StreamID, n protocol.ByteCount) error
	GetBytesRetrans(streamID protocol.StreamID) (protocol.ByteCount, error)
>>>>>>> project-faster/main
}
