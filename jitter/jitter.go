package jitter

import "sort"

// maxWindow 是抖动重排窗口的上限：最多缓存这么多未输出的包，
// 超过后强制按环形序号顺序输出，保证工作量随包数线性增长。
const maxWindow = 64

// seqLess 按 16 位环形序号空间比较，a 在 b 之前（含 65535 -> 0 回绕）时为真。
func seqLess(a, b uint16) bool { return int16(a-b) < 0 }

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
	if len(packets) == 0 {
		return Result{}
	}
	ordered := append([]Packet(nil), packets...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].ArrivalMs < ordered[j].ArrivalMs })

	frames := make([]Frame, 0)
	current := Frame{}
	buffer := make(map[uint16]Packet)
	seen := make(map[uint16]bool, len(ordered))
	var base uint16 // 下一个期望输出的序号
	baseSet := false
	emitted := false
	dropped := 0
	pendingLost := 0 // 帧间发现的缺口，挂到下一帧

	closeFrame := func() {
		if current.Packets > 0 {
			frames = append(frames, current)
			current = Frame{}
		}
	}
	emit := func(p Packet) {
		// 时间戳变化说明上一帧没有 marker 也结束了，帧间状态隔离。
		if current.Packets > 0 && p.Timestamp != current.Timestamp {
			closeFrame()
		}
		if current.Packets == 0 {
			current.Timestamp = p.Timestamp
			current.Lost = pendingLost
			pendingLost = 0
		}
		current.Payload += p.Payload
		current.Packets++
		if p.Marker {
			closeFrame()
		}
	}
	// drain 按环形顺序从 buffer 输出包；forceAll 为假时只在窗口溢出时输出。
	drain := func(forceAll bool) {
		for len(buffer) > 0 && (forceAll || len(buffer) > maxWindow) {
			if p, ok := buffer[base]; ok {
				delete(buffer, base)
				emit(p)
				base++
				emitted = true
				continue
			}
			// base 缺失：在窗口内找环形意义下最靠前的包，缺口记为丢包。
			var next uint16
			first := true
			for s := range buffer {
				if first || seqLess(s, next) {
					next = s
					first = false
				}
			}
			gap := int(uint16(next - base))
			if current.Packets > 0 {
				current.Lost += gap
			} else {
				pendingLost += gap
			}
			base = next
		}
	}

	for _, p := range ordered {
		if seen[p.Seq] {
			dropped++ // 重复到达，只消费一次
			continue
		}
		seen[p.Seq] = true
		if !baseSet {
			base = p.Seq
			baseSet = true
		} else if seqLess(p.Seq, base) {
			if emitted {
				dropped++ // 太晚到达，序号已被越过
				continue
			}
			base = p.Seq // 尚未输出，允许更靠前的晚到包回退基准
		}
		buffer[p.Seq] = p
		drain(false)
	}
	drain(true)
	closeFrame()
	return Result{Frames: frames, Dropped: dropped, Scanned: len(packets)}
}
