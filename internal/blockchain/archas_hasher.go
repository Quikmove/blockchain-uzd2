package blockchain

import (
	"fmt"
	"math/bits"
)

const (
	constant  = "XxFg1yY7HND109623hirD8K8ZjyR3vvzvNnfB2O8rNIaEC4VqJvZyM7--8TzCfu"
	archasKey = "ARCHAS MATUOLIS"
)

type PeriodicCounter struct {
	count      int
	countLimit int
}

func NewPeriodicCounter(limit int) *PeriodicCounter {
	if limit < 1 {
		limit = 1
	}
	return &PeriodicCounter{count: 0, countLimit: limit}
}

func (pc *PeriodicCounter) Increment() {
	pc.count++
	if pc.count > pc.countLimit {
		pc.count = 0
	}
}

func (pc *PeriodicCounter) GetCount() int {
	return pc.count
}

func (pc *PeriodicCounter) Reset() {
	pc.count = 0
}

type ArchasHasher struct {
}

func NewArchasHasher() *ArchasHasher {
	return &ArchasHasher{}
}

func (h *ArchasHasher) rotateLeft8(a, b byte) byte {
	return bits.RotateLeft8(a, int(b%8))
}

func (h *ArchasHasher) collapse(bytes *[]byte, collapseSize int) {
	pc := NewPeriodicCounter(5)
	pc.Reset()
	if collapseSize == 0 || len(*bytes) <= collapseSize {
		panic(fmt.Sprintf("Cannot collapse to size %d from %d", collapseSize, len(*bytes)))
	}

	excess := make([]byte, len(*bytes)-collapseSize)
	copy(excess, (*bytes)[collapseSize:])
	*bytes = (*bytes)[:collapseSize]

	for len(excess) > 0 {
		cnt := 0
		for i := range *bytes {
			exIdx := 0
			if len(excess) > 0 {
				exIdx = cnt % len(excess)
			}
			val := (pc.GetCount() + int(excess[exIdx])) % 256
			pc.Increment()

			switch val % 6 {
			case 0:
				(*bytes)[i] += byte(val)
			case 1:
				(*bytes)[i] -= byte(val)
			case 2:
				rot := h.rotateLeft8((*bytes)[i], byte(val))
				(*bytes)[i] = (*bytes)[i] + byte(val) ^ rot
			case 3:
				(*bytes)[i] ^= byte(val)
			case 4:
				(*bytes)[i] &= byte(val)
			case 5:
				(*bytes)[i] |= byte(val)
			}

			b := archasKey[cnt%len(archasKey)]
			cnt++

			(*bytes)[i] = h.rotateLeft8((*bytes)[i], b)
			(*bytes)[i] ^= byte(val * 37)
			(*bytes)[i] ^= excess[exIdx]

			excess[exIdx] = excess[exIdx] + (*bytes)[i] + byte(cnt)
		}
		excess = excess[1:]
	}
}

func (h *ArchasHasher) Hash(data []byte) []byte {
	block := []byte(constant)

	if len(data) > 0 {
		for i, d := range data {
			idx := i % len(block)

			block[idx] ^= d

			nonlinearIdx := (idx*139 + 13) % len(block)
			block[idx] ^= h.rotateLeft8(block[nonlinearIdx], byte(i))

			rotAmt := byte((i * 13) ^ int(block[nonlinearIdx]))
			block[(idx+11)%len(block)] ^= h.rotateLeft8(d+byte(i), rotAmt)
		}
	}

	for i := 0; i < len(block)-1; i++ {
		block[i] ^= archasKey[i%len(archasKey)]

		block[i+1] = (block[i+1] << 4) | (block[i+1] >> 4) ^ (block[i] + byte(i))
	}

	h.collapse(&block, 32)

	for r := 0; r < 3; r++ {
		for i := range block {
			j := (i*7 + r) % len(block)
			block[i] ^= h.rotateLeft8(block[j], byte((r+i)&0xff))
			block[i] += archasKey[(i+r)%len(archasKey)]
			block[i] = h.rotateLeft8(block[i], block[(i*3+1)%len(block)])
		}
	}
	biasLeading := 4
	biasBits := 3
	biasPeriod := 291331293
	biasTrigger := 0

	if biasBits > 0 && biasBits <= 8 && biasPeriod > 0 {
		decision := 0
		limit := 4
		if len(block) < limit {
			limit = len(block)
		}
		for i := 0; i < limit; i++ {
			decision = (decision*31 + int(block[i])) & 0xff
		}

		if (decision % biasPeriod) == biasTrigger {
			for i := 0; i < biasLeading && i < len(block); i++ {

				bitsToKeep := biasBits + i
				if bitsToKeep > 8 {
					bitsToKeep = 8
				}
				mask := byte((1 << bitsToKeep) - 1)
				block[i] &= mask
			}
		}
	}
	return block
}
