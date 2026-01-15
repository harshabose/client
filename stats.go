package client

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type CanGetStats interface {
	Generate(*PeerConnection) Stat
}

type StatsGeneratorFunc func(*PeerConnection) Stat

func (f StatsGeneratorFunc) Generate(pc *PeerConnection) Stat {
	return f(pc)
}

type StatsGetter struct {
	// getter stats.Getter
	c *Client

	mux    sync.Mutex
	once   sync.Once
	wg     sync.WaitGroup
	cancel context.CancelFunc
	ctx    context.Context
}

func NewStatsGetter(ctx context.Context, c *Client, interval time.Duration) *StatsGetter {
	ctx2, cancel2 := context.WithCancel(ctx)

	s := &StatsGetter{
		// getter: getter,
		c:      c,
		ctx:    ctx2,
		cancel: cancel2,
	}

	go s.loop1(interval)
	return s
}

func (g *StatsGetter) loop1(interval time.Duration) {
	g.wg.Add(1)
	defer g.wg.Done()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-g.ctx.Done():
			return
		case <-ticker.C:
			for _, pc := range g.c.PeerConnections() {
				stats := pc.GetPeerConnection().GetStats()

				for _, s := range stats {
					if err := pc.stat.Consume(s); err != nil {
						fmt.Printf("error while gathering stats; (err: %v)\n", err)
						continue
					}
				}
			}
		}
	}
}

func (g *StatsGetter) Close() error {
	g.once.Do(func() {
		if g.cancel != nil {
			g.cancel()
		}

		g.wg.Wait()
	})

	return nil
}

func (g *StatsGetter) Generate(pc *PeerConnection) Stat {
	return pc.stat.Generate()
}
