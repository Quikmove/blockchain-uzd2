package blockchain

import (
	"fmt"
	"math/bits"
)

const (
	constant   = "XxFg1yY7HND109623hirD8K8ZjyR3vvzvNnfB2O8rNIaEC4VqJvZyM7--8TzCfu"
	archas_key = "ARCHAS MATUOLIS"
)

// PeriodicCounter is a Go equivalent of the C++ PeriodicCounter class.
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
	pc *PeriodicCounter
}

func NewArchasHasher() *ArchasHasher {
	return &ArchasHasher{
		pc: NewPeriodicCounter(5),
	}
}

func (h *ArchasHasher) rotateLeft8(a, b byte) byte {
	return bits.RotateLeft8(a, int(b%8))
}

func (h *ArchasHasher) collapse(bytes *[]byte, collapseSize int) error {
	h.pc.Reset()
	if collapseSize == 0 || len(*bytes) <= collapseSize {
		return fmt.Errorf("invalid collapse size")
	}

	excess := make([]byte, len(*bytes)-collapseSize)
	copy(excess, (*bytes)[collapseSize:])
	*bytes = (*bytes)[:collapseSize]

	for len(excess) > 0 {
		cnt := 0
		for i := range *bytes {
			val := (h.pc.GetCount() + int(excess[0])) % 256
			h.pc.Increment()

			switch val % 6 {
			case 0:
				(*bytes)[i] += byte(val)
			case 1:
				(*bytes)[i] -= byte(val)
			case 2:
				(*bytes)[i] *= byte(val)
			case 3:
				(*bytes)[i] ^= byte(val)
			case 4:
				(*bytes)[i] &= byte(val)
			case 5:
				(*bytes)[i] |= byte(val)
			}

			b := archas_key[cnt%len(archas_key)]
			cnt++

			(*bytes)[i] = h.rotateLeft8((*bytes)[i], b)
			(*bytes)[i] ^= byte(val * 37)
			(*bytes)[i] ^= excess[0]
			excess[0] = excess[0] + (*bytes)[i] + byte(cnt)
		}
		excess = excess[1:]
	}
	return nil
}

func (h *ArchasHasher) Hash(data []byte) ([]byte, error) {
	block := []byte(constant)

	if len(data) > 0 {
		for i, d := range data {
			idx := i % len(block)

			block[idx] ^= d

			nonlinearIdx := (idx*3 + 13) % len(block)
			block[idx] ^= h.rotateLeft8(block[nonlinearIdx], byte(i))

			block[(idx+11)%len(block)] ^= h.rotateLeft8(d+byte(i), (byte(i*13))&0xc5)
		}
	}

	for i := 0; i < len(block)-1; i++ {
		block[i] ^= archas_key[i%len(archas_key)]

		block[i+1] = (block[i+1] << 4) | (block[i+1] >> 4) ^ (block[i] + byte(i))
	}

	err := h.collapse(&block, 32)
	if err != nil {
		return nil, err
	}

	return block, nil
}
