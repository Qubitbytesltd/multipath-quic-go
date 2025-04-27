package wire

<<<<<<< HEAD
import (
	"fmt"
	"strings"

	"github.com/quic-go/quic-go/internal/protocol"
	"github.com/quic-go/quic-go/internal/utils"
)

// LogFrame logs a frame, either sent or received
func LogFrame(logger utils.Logger, frame Frame, sent bool) {
	if !logger.Debug() {
=======
import "github.com/project-faster/mp-quic-go/internal/utils"

// LogFrame logs a frame, either sent or received
func LogFrame(frame Frame, sent bool) {
	if !utils.Debug() {
>>>>>>> project-faster/main
		return
	}
	dir := "<-"
	if sent {
		dir = "->"
	}
	switch f := frame.(type) {
<<<<<<< HEAD
	case *CryptoFrame:
		dataLen := protocol.ByteCount(len(f.Data))
		logger.Debugf("\t%s &wire.CryptoFrame{Offset: %d, Data length: %d, Offset + Data length: %d}", dir, f.Offset, dataLen, f.Offset+dataLen)
	case *StreamFrame:
		logger.Debugf("\t%s &wire.StreamFrame{StreamID: %d, Fin: %t, Offset: %d, Data length: %d, Offset + Data length: %d}", dir, f.StreamID, f.Fin, f.Offset, f.DataLen(), f.Offset+f.DataLen())
	case *ResetStreamFrame:
		logger.Debugf("\t%s &wire.ResetStreamFrame{StreamID: %d, ErrorCode: %#x, FinalSize: %d}", dir, f.StreamID, f.ErrorCode, f.FinalSize)
	case *AckFrame:
		hasECN := f.ECT0 > 0 || f.ECT1 > 0 || f.ECNCE > 0
		var ecn string
		if hasECN {
			ecn = fmt.Sprintf(", ECT0: %d, ECT1: %d, CE: %d", f.ECT0, f.ECT1, f.ECNCE)
		}
		if len(f.AckRanges) > 1 {
			ackRanges := make([]string, len(f.AckRanges))
			for i, r := range f.AckRanges {
				ackRanges[i] = fmt.Sprintf("{Largest: %d, Smallest: %d}", r.Largest, r.Smallest)
			}
			logger.Debugf("\t%s &wire.AckFrame{LargestAcked: %d, LowestAcked: %d, AckRanges: {%s}, DelayTime: %s%s}", dir, f.LargestAcked(), f.LowestAcked(), strings.Join(ackRanges, ", "), f.DelayTime.String(), ecn)
		} else {
			logger.Debugf("\t%s &wire.AckFrame{LargestAcked: %d, LowestAcked: %d, DelayTime: %s%s}", dir, f.LargestAcked(), f.LowestAcked(), f.DelayTime.String(), ecn)
		}
	case *MaxDataFrame:
		logger.Debugf("\t%s &wire.MaxDataFrame{MaximumData: %d}", dir, f.MaximumData)
	case *MaxStreamDataFrame:
		logger.Debugf("\t%s &wire.MaxStreamDataFrame{StreamID: %d, MaximumStreamData: %d}", dir, f.StreamID, f.MaximumStreamData)
	case *DataBlockedFrame:
		logger.Debugf("\t%s &wire.DataBlockedFrame{MaximumData: %d}", dir, f.MaximumData)
	case *StreamDataBlockedFrame:
		logger.Debugf("\t%s &wire.StreamDataBlockedFrame{StreamID: %d, MaximumStreamData: %d}", dir, f.StreamID, f.MaximumStreamData)
	case *MaxStreamsFrame:
		switch f.Type {
		case protocol.StreamTypeUni:
			logger.Debugf("\t%s &wire.MaxStreamsFrame{Type: uni, MaxStreamNum: %d}", dir, f.MaxStreamNum)
		case protocol.StreamTypeBidi:
			logger.Debugf("\t%s &wire.MaxStreamsFrame{Type: bidi, MaxStreamNum: %d}", dir, f.MaxStreamNum)
		}
	case *StreamsBlockedFrame:
		switch f.Type {
		case protocol.StreamTypeUni:
			logger.Debugf("\t%s &wire.StreamsBlockedFrame{Type: uni, MaxStreams: %d}", dir, f.StreamLimit)
		case protocol.StreamTypeBidi:
			logger.Debugf("\t%s &wire.StreamsBlockedFrame{Type: bidi, MaxStreams: %d}", dir, f.StreamLimit)
		}
	case *NewConnectionIDFrame:
		logger.Debugf("\t%s &wire.NewConnectionIDFrame{SequenceNumber: %d, RetirePriorTo: %d, ConnectionID: %s, StatelessResetToken: %#x}", dir, f.SequenceNumber, f.RetirePriorTo, f.ConnectionID, f.StatelessResetToken)
	case *RetireConnectionIDFrame:
		logger.Debugf("\t%s &wire.RetireConnectionIDFrame{SequenceNumber: %d}", dir, f.SequenceNumber)
	case *NewTokenFrame:
		logger.Debugf("\t%s &wire.NewTokenFrame{Token: %#x}", dir, f.Token)
	default:
		logger.Debugf("\t%s %#v", dir, frame)
=======
	case *StreamFrame:
		utils.Debugf("\t%s &wire.StreamFrame{StreamID: %d, FinBit: %t, Offset: 0x%x, Data length: 0x%x, Offset + Data length: 0x%x}", dir, f.StreamID, f.FinBit, f.Offset, f.DataLen(), f.Offset+f.DataLen())
	case *StopWaitingFrame:
		if sent {
			utils.Debugf("\t%s &wire.StopWaitingFrame{LeastUnacked: 0x%x, PacketNumberLen: 0x%x}", dir, f.LeastUnacked, f.PacketNumberLen)
		} else {
			utils.Debugf("\t%s &wire.StopWaitingFrame{LeastUnacked: 0x%x}", dir, f.LeastUnacked)
		}
	case *AckFrame:
		utils.Debugf("\t%s &wire.AckFrame{PathID: 0x%x, LargestAcked: 0x%x, LowestAcked: 0x%x, AckRanges: %#v, DelayTime: %s}", dir, f.PathID, f.LargestAcked, f.LowestAcked, f.AckRanges, f.DelayTime.String())
	case *AddAddressFrame:
		utils.Debugf("\t%s &wire.AddAddressFrame{IPVersion: %d, Addr: %s}", dir, f.IPVersion, f.Addr.String())
	case *ClosePathFrame:
		utils.Debugf("\t%s &wire.ClosePathFrame{PathID: 0x%x, LargestAcked: 0x%x, LowestAcked: 0x%x, AckRanges: %#v}", dir, f.PathID, f.LargestAcked, f.LowestAcked, f.AckRanges)
	default:
		utils.Debugf("\t%s %#v", dir, frame)
>>>>>>> project-faster/main
	}
}
