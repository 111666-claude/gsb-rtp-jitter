package jitter

import "sort"

type Packet struct {
	Seq       uint16 `json:"seq"`
	Timestamp uint32 `json:"timestamp"`
	Marker    bool   `json:"marker"`
	Payload   string `json:"payload"`
	ArrivalMs int64  `json:"arrival_ms"`
}

type Frame struct {
	Timestamp uint32 `json:"timestamp"`
	Payload   string `json:"payload"`
	Packets   int    `json:"packets"`
	Lost      int    `json:"lost"`
}

type Result struct {
	Frames  []Frame `json:"frames"`
	Dropped int     `json:"dropped"`
	Scanned int     `json:"scanned"`
}

func Assemble(packets []Packet) Result {
	ordered := append([]Packet(nil), packets...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].ArrivalMs < ordered[j].ArrivalMs })
	if len(ordered) == 0 {
		return Result{}
	}
	expected := ordered[0].Seq
	seen := map[uint16]bool{}
	frames := make([]Frame, 0)
	current := Frame{}
	for _, packet := range ordered {
		if packet.Seq < expected {
			continue
		}
		if packet.Seq > expected {
			if current.Packets > 0 {
				current.Lost += int(packet.Seq - expected)
			}
			expected = packet.Seq
		}
		seen[packet.Seq] = true
		if current.Packets == 0 {
			current.Timestamp = packet.Timestamp
		}
		current.Payload += packet.Payload
		current.Packets++
		if packet.Marker {
			frames = append(frames, current)
			current = Frame{}
		}
		expected = packet.Seq + 1
	}
	if current.Packets > 0 {
		frames = append(frames, current)
	}
	return Result{Frames: frames, Dropped: len(packets) - len(seen), Scanned: len(packets)}
}
