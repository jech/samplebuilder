package samplebuilder

import (
	"reflect"
	"testing"
	"time"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3/pkg/media"
)

type fakeDepacketizer struct {
	headBytes   []byte
	tailChecker func([]byte, bool) bool
}

func (f *fakeDepacketizer) Unmarshal(r []byte) ([]byte, error) {
	return r, nil
}

func (f *fakeDepacketizer) IsPartitionHead(payload []byte) bool {
	for _, b := range f.headBytes {
		if payload[0] == b {
			return true
		}
	}
	return false
}

func (f *fakeDepacketizer) IsPartitionTail(marker bool, payload []byte) bool {
	if f.tailChecker != nil {
		return f.tailChecker(payload, marker)
	}
	return false
}

// for compatibility with Pion brain-damage
func (f *fakeDepacketizer) IsDetectedFinalPacketInSequence(rtpPacketMarketBit bool) bool {
	return rtpPacketMarketBit
}

type test struct {
	name        string
	maxLate     uint16
	headBytes   []byte
	tailChecker func([]byte, bool) bool
	packets     []*rtp.Packet
	samples     []*media.Sample
	timestamps  []uint32
}

// some tests stolen from Pion's samplebuilder
var tests = []test{
	{
		name: "One",
		packets: []*rtp.Packet{
			{Header: rtp.Header{SequenceNumber: 5000, Timestamp: 5}, Payload: []byte{0x01}},
		},
		samples:    []*media.Sample{},
		timestamps: []uint32{},
		maxLate:    50,
	},
	{
		name: "Sequential",
		packets: []*rtp.Packet{
			{Header: rtp.Header{SequenceNumber: 5000, Timestamp: 5}, Payload: []byte{0x01}},
			{Header: rtp.Header{SequenceNumber: 5001, Timestamp: 6}, Payload: []byte{0x02}},
			{Header: rtp.Header{SequenceNumber: 5002, Timestamp: 7}, Payload: []byte{0x03}},
		},
		samples: []*media.Sample{
			{Data: []byte{0x02}, Duration: time.Second},
		},
		timestamps: []uint32{
			6,
		},
		maxLate: 50,
	},
	{
		name: "Duplicate",
		packets: []*rtp.Packet{
			{Header: rtp.Header{SequenceNumber: 5000, Timestamp: 5}, Payload: []byte{0x01}},
			{Header: rtp.Header{SequenceNumber: 5001, Timestamp: 6}, Payload: []byte{0x02}},
			{Header: rtp.Header{SequenceNumber: 5002, Timestamp: 6}, Payload: []byte{0x03}},
			{Header: rtp.Header{SequenceNumber: 5003, Timestamp: 7}, Payload: []byte{0x04}},
		},
		samples: []*media.Sample{
			{Data: []byte{0x02, 0x03}, Duration: time.Second},
		},
		timestamps: []uint32{
			6,
		},
		maxLate: 50,
	},
	{
		name: "Gap",
		packets: []*rtp.Packet{
			{Header: rtp.Header{SequenceNumber: 5000, Timestamp: 5}, Payload: []byte{0x01}},
			{Header: rtp.Header{SequenceNumber: 5007, Timestamp: 6}, Payload: []byte{0x02}},
			{Header: rtp.Header{SequenceNumber: 5008, Timestamp: 7}, Payload: []byte{0x03}},
		},
		samples:    []*media.Sample{},
		timestamps: []uint32{},
		maxLate:    50,
	},
	{
		name: "GapPartitionHeadCheckerTrue",
		packets: []*rtp.Packet{
			{Header: rtp.Header{SequenceNumber: 5000, Timestamp: 5}, Payload: []byte{0x01}},
			{Header: rtp.Header{SequenceNumber: 5007, Timestamp: 6}, Payload: []byte{0x02}},
			{Header: rtp.Header{SequenceNumber: 5008, Timestamp: 7}, Payload: []byte{0x03}},
		},
		headBytes: []byte{0x02},
		samples: []*media.Sample{
			{Data: []byte{0x02}, Duration: 0},
		},
		timestamps: []uint32{
			6,
		},
		maxLate: 5,
	},
	{
		name: "GapPartitionHeadCheckerFalse",
		packets: []*rtp.Packet{
			{Header: rtp.Header{SequenceNumber: 5000, Timestamp: 5}, Payload: []byte{0x01}},
			{Header: rtp.Header{SequenceNumber: 5007, Timestamp: 6}, Payload: []byte{0x02}},
			{Header: rtp.Header{SequenceNumber: 5008, Timestamp: 7}, Payload: []byte{0x03}},
		},
		headBytes:  []byte{},
		samples:    []*media.Sample{},
		timestamps: []uint32{},
		maxLate:    5,
	},
	{
		name: "Multiple",
		packets: []*rtp.Packet{
			{Header: rtp.Header{SequenceNumber: 5000, Timestamp: 1}, Payload: []byte{0x01}},
			{Header: rtp.Header{SequenceNumber: 5001, Timestamp: 2}, Payload: []byte{0x02}},
			{Header: rtp.Header{SequenceNumber: 5002, Timestamp: 3}, Payload: []byte{0x03}},
			{Header: rtp.Header{SequenceNumber: 5003, Timestamp: 4}, Payload: []byte{0x04}},
			{Header: rtp.Header{SequenceNumber: 5004, Timestamp: 5}, Payload: []byte{0x05}},
			{Header: rtp.Header{SequenceNumber: 5005, Timestamp: 6}, Payload: []byte{0x06}},
		},
		samples: []*media.Sample{
			{Data: []byte{0x02}, Duration: time.Second},
			{Data: []byte{0x03}, Duration: time.Second},
			{Data: []byte{0x04}, Duration: time.Second},
			{Data: []byte{0x05}, Duration: time.Second},
		},
		timestamps: []uint32{
			2,
			3,
			4,
			5,
		},
		maxLate: 5,
	},
	{
		name: "PartitionTailChecker",
		packets: []*rtp.Packet{
			{Header: rtp.Header{SequenceNumber: 5000, Timestamp: 5}, Payload: []byte{0x01}},
			{Header: rtp.Header{SequenceNumber: 5001, Timestamp: 6}, Payload: []byte{0x02}},
			{Header: rtp.Header{SequenceNumber: 5002, Timestamp: 7}, Payload: []byte{0x03}},
		},
		samples: []*media.Sample{
			{Data: []byte{0x02}, Duration: time.Second},
			{Data: []byte{0x03}, Duration: time.Second},
		},
		timestamps: []uint32{
			6, 7,
		},
		tailChecker: func(payload []byte, marker bool) bool {
			return true
		},
		maxLate: 50,
	},
	{
		name: "Checkers",
		packets: []*rtp.Packet{
			{Header: rtp.Header{SequenceNumber: 5000, Timestamp: 5}, Payload: []byte{0x01}},
			{Header: rtp.Header{SequenceNumber: 5001, Timestamp: 6}, Payload: []byte{0x02}},
			{Header: rtp.Header{SequenceNumber: 5002, Timestamp: 7}, Payload: []byte{0x03}},
		},
		samples: []*media.Sample{
			{Data: []byte{0x01}, Duration: 0},
			{Data: []byte{0x02}, Duration: time.Second},
			{Data: []byte{0x03}, Duration: time.Second},
		},
		timestamps: []uint32{
			5, 6, 7,
		},
		headBytes: []byte{1},
		tailChecker: func(payload []byte, marker bool) bool {
			return true
		},
		maxLate: 50,
	},
}

func TestSamplebuilder(t *testing.T) {
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := New(test.maxLate, &fakeDepacketizer{
				headBytes:   test.headBytes,
				tailChecker: test.tailChecker,
			}, 1)
			samples := []*media.Sample{}
			timestamps := []uint32{}

			for _, p := range test.packets {
				s.Push(p)
			}
			for {
				sample, timestamp := s.ForcePopWithTimestamp()
				if sample == nil {
					break
				}
				samples = append(samples, sample)
				timestamps = append(timestamps, timestamp)
			}
			if !reflect.DeepEqual(samples, test.samples) {
				t.Errorf("got %#v, expected %#v",
					samples, test.samples,
				)
			}
			if !reflect.DeepEqual(timestamps, test.timestamps) {
				t.Errorf("got %v, expected %v",
					timestamps, test.timestamps,
				)
			}
			for i := s.tail; i != s.head; i = s.inc(i) {
				println(i, s.packets[i].start, s.packets[i].end, s.packets[i].packet)
			}

		})
	}
}

type truePartitionHeadChecker struct{}

func (f *truePartitionHeadChecker) IsPartitionHead(payload []byte) bool {
	return true
}

func TestSampleBuilderSequential(t *testing.T) {
	s := New(10, &fakeDepacketizer{}, 1)
	j := 0
	for i := 0; i < 0x20000; i++ {
		p := rtp.Packet{
			Header: rtp.Header{
				SequenceNumber: uint16(i),
				Timestamp:      uint32(i + 42),
			},
			Payload: []byte{byte(i)},
		}
		s.Push(&p)
		for {
			sample, ts := s.PopWithTimestamp()
			if sample == nil {
				break
			}
			if ts != uint32(j+43) {
				t.Errorf(
					"wrong timestamp (got %v, expected %v)",
					ts, uint32(j+43),
				)
			}
			if len(sample.Data) != 1 {
				t.Errorf(
					"bad data length (got %v, expected 1)",
					len(sample.Data),
				)
			}
			if sample.Data[0] != byte(j+1) {
				t.Errorf(
					"bad data (got %v, expected %v)",
					sample.Data[0], byte(j+1))
			}
			j++
		}
	}
	// only the first and last packet should be dropped
	if j != 0x1FFFE {
		t.Errorf("Got %v, expected %v", j, 0x1FFFE)
	}
}

func TestSampleBuilderFull(t *testing.T) {
	s := New(10, &fakeDepacketizer{[]byte{0}, nil}, 1)
	s.Push(&rtp.Packet{
		Header: rtp.Header{SequenceNumber: 5000, Timestamp: 5},
		Payload: []byte{0},
	})
	for i := uint16(5001); i < 5020; i++ {
		s.Push(&rtp.Packet{
			Header: rtp.Header{SequenceNumber: i, Timestamp: 5},
			Payload: []byte{1},
		})
	}
}

func BenchmarkSampleBuilderSequential(b *testing.B) {
	s := New(100, &fakeDepacketizer{}, 1)
	b.ResetTimer()
	j := 0
	for i := 0; i < b.N; i++ {
		p := rtp.Packet{
			Header: rtp.Header{
				SequenceNumber: uint16(i),
				Timestamp:      uint32(i + 42),
			},
			Payload: make([]byte, 50),
		}
		s.Push(&p)
		for {
			s := s.Pop()
			if s == nil {
				break
			}
			j++
		}
	}
	if b.N > 200 && j < b.N-100 {
		b.Errorf("Got %v (N=%v)", j, b.N)
	}
}

func BenchmarkSampleBuilderLoss(b *testing.B) {
	s := New(100, &fakeDepacketizer{}, 1)
	b.ResetTimer()
	j := 0
	for i := 0; i < b.N; i++ {
		if i%13 == 0 {
			continue
		}
		p := rtp.Packet{
			Header: rtp.Header{
				SequenceNumber: uint16(i),
				Timestamp:      uint32(i + 42),
			},
			Payload: make([]byte, 50),
		}
		s.Push(&p)
		for {
			s := s.Pop()
			if s == nil {
				break
			}
			j++
		}
	}
	if b.N > 200 && j < b.N/2-100 {
		b.Errorf("Got %v (N=%v)", j, b.N)
	}
}

func BenchmarkSampleBuilderReordered(b *testing.B) {
	s := New(100, &fakeDepacketizer{}, 1)
	b.ResetTimer()
	j := 0
	for i := 0; i < b.N; i++ {
		p := rtp.Packet{
			Header: rtp.Header{
				SequenceNumber: uint16(i ^ 3),
				Timestamp:      uint32((i ^ 3) + 42),
			},
			Payload: make([]byte, 50),
		}
		s.Push(&p)
		for {
			s := s.Pop()
			if s == nil {
				break
			}
			j++
		}
	}
	if b.N > 2 && j < b.N-5 && j > b.N {
		b.Errorf("Got %v (N=%v)", j, b.N)
	}
}

func BenchmarkSampleBuilderFragmented(b *testing.B) {
	s := New(100, &fakeDepacketizer{}, 1)
	b.ResetTimer()
	j := 0
	for i := 0; i < b.N; i++ {
		p := rtp.Packet{
			Header: rtp.Header{
				SequenceNumber: uint16(i),
				Timestamp:      uint32(i/2 + 42),
			},
			Payload: make([]byte, 50),
		}
		s.Push(&p)
		for {
			s := s.Pop()
			if s == nil {
				break
			}
			j++
		}
	}
	if b.N > 200 && j < b.N/2-100 {
		b.Errorf("Got %v (N=%v)", j, b.N)
	}
}

func BenchmarkSampleBuilderFragmentedLoss(b *testing.B) {
	s := New(100, &fakeDepacketizer{}, 1)
	b.ResetTimer()
	j := 0
	for i := 0; i < b.N; i++ {
		if i%13 == 0 {
			continue
		}
		p := rtp.Packet{
			Header: rtp.Header{
				SequenceNumber: uint16(i),
				Timestamp:      uint32(i/2 + 42),
			},
			Payload: make([]byte, 50),
		}
		s.Push(&p)
		for {
			s := s.Pop()
			if s == nil {
				break
			}
			j++
		}
	}
	if b.N > 200 && j < b.N/3-100 {
		b.Errorf("Got %v (N=%v)", j, b.N)
	}
}
