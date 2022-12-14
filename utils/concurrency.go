package utils

import (
	"context"
	"sync"

	"golang.org/x/sync/semaphore"
)

func workerPool[W any](ctx context.Context, goroutines int, f func(W) W, chans ...<-chan W) <-chan W {

	out := make(chan W)
	wg := &sync.WaitGroup{}
	sem := semaphore.NewWeighted(int64(goroutines))

	for _, c := range chans {

		for i := range c {

			wg.Add(1)
			go func(i W) {
				sem.Acquire(ctx, 1)
				{
					out <- f(i)
					wg.Done()
				}
				sem.Release(1)
			}(i)

		}

	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out

}

func fanIn[W any](chans ...<-chan W) <-chan W {

	out := make(chan W)
	wg := &sync.WaitGroup{}
	wg.Add(len(chans))

	for _, c := range chans {

		go func(c <-chan W) {
			for r := range c {
				out <- r
			}
			wg.Done()
		}(c)

	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out

}
