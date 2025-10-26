package progress

import (
	"fmt"
	"sync"
	"time"
)

type ProgressTracker struct {
	total     int
	current   int
	message   string
	mu        sync.Mutex
	startTime time.Time
	done      chan bool
}

func NewProgress(total int, message string) *ProgressTracker {
	p := &ProgressTracker{
		total:     total,
		message:   message,
		startTime: time.Now(),
		done:      make(chan bool),
	}
	go p.render()
	return p
}

func (p *ProgressTracker) render() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	frame := 0

	for {
		select {
		case <-p.done:
			p.mu.Lock()
			elapsed := time.Since(p.startTime)
			fmt.Printf("\r✓ %s (%d files, %s)          \n",
				p.message, p.total, elapsed.Round(time.Millisecond))
			p.mu.Unlock()
			return

		case <-ticker.C:
			p.mu.Lock()
			if p.total > 0 {
				percent := float64(p.current) / float64(p.total) * 100
				fmt.Printf("\r%s %s [%d/%d] %.0f%%  ",
					spinner[frame%len(spinner)],
					p.message,
					p.current,
					p.total,
					percent)
			} else {
				fmt.Printf("\r%s %s [%d files]  ",
					spinner[frame%len(spinner)],
					p.message,
					p.current)
			}
			p.mu.Unlock()
			frame++
		}
	}
}

func (p *ProgressTracker) Increment() {
	p.mu.Lock()
	p.current++
	p.mu.Unlock()
}

func (p *ProgressTracker) SetCurrent(n int) {
	p.mu.Lock()
	p.current = n
	p.mu.Unlock()
}

func (p *ProgressTracker) Finish() {
	close(p.done)
	time.Sleep(1 * time.Millisecond)
}
