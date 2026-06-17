package runtime

import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"
)

type Stats struct {
	parsed  atomic.Uint64
	mapped  atomic.Uint64
	dropped atomic.Uint64

	mu          sync.Mutex
	dropReasons map[string]uint64 // drop reason
}

func NewStats() *Stats {
	return &Stats{
		dropReasons: make(map[string]uint64),
	}
}

func (s *Stats) Print(w io.Writer) {
	fmt.Fprintf(w, "parsed=%d mapped=%d dropped=%d\n",
		s.parsed.Load(),
		s.mapped.Load(),
		s.dropped.Load())

	s.mu.Lock()
	defer s.mu.Unlock()
	for reason, count := range s.dropReasons {
		fmt.Fprintf(w, "  dropped[%s]=%d\n", reason, count)
	}
}

func (s *Stats) IncParsed() {
	s.parsed.Add(1)
}

func (s *Stats) IncMapped() {
	s.mapped.Add(1)
}

func (s *Stats) IncDropped(reason string) {
	s.dropped.Add(1)
	s.mu.Lock()
	s.dropReasons[reason]++
	s.mu.Unlock()
}
