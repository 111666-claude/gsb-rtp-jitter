package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"example.com/rtpjitter/jitter"
)

func main() {
	sample := flag.String("sample", "ordered", "sample name")
	flag.Parse()
	var packets []jitter.Packet
	switch *sample {
	case "ordered":
		packets = []jitter.Packet{{Seq: 100, Timestamp: 480, Payload: "I", ArrivalMs: 1}, {Seq: 101, Timestamp: 480, Marker: true, Payload: "P", ArrivalMs: 2}}
	case "reorder":
		packets = []jitter.Packet{{Seq: 201, Timestamp: 960, Marker: true, Payload: "B", ArrivalMs: 3}, {Seq: 200, Timestamp: 960, Payload: "A", ArrivalMs: 4}}
	default:
		fmt.Fprintln(os.Stderr, "unknown sample")
		os.Exit(2)
	}
	data, _ := json.Marshal(jitter.Assemble(packets))
	fmt.Println(string(data))
}
