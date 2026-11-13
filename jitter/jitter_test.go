package jitter

import "testing"

func TestAssembleOrderedFrame(t *testing.T) {
	got := Assemble([]Packet{{Seq: 10, Timestamp: 90, Payload: "a", ArrivalMs: 1}, {Seq: 11, Timestamp: 90, Marker: true, Payload: "b", ArrivalMs: 2}})
	if len(got.Frames) != 1 || got.Frames[0].Payload != "ab" || got.Frames[0].Packets != 2 {
		t.Fatalf("unexpected result: %+v", got)
	}
}
