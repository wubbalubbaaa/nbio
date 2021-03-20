// Copyright 2020 lesismal. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package taskpool

import (
	"runtime"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/lesismal/nbio/loging"
)

// fixedRunner .
type fixedRunner struct {
	wg *sync.WaitGroup

	chTask   chan func()
	chTaskBy chan func()
	chClose  chan struct{}
}

func (r *fixedRunner) call(f func()) {
	defer func() {
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			loging.Error("taskpool fixedRunner call failed: %v\n%v\n", err, *(*string)(unsafe.Pointer(&buf)))
		}
	}()
	f()
}

func (r *fixedRunner) taskLoop() {
	defer r.wg.Done()

	// run all tasks
	defer func() {
		for {
			select {
			case f := <-r.chTaskBy:
				r.call(f)
			case f := <-r.chTask:
				r.call(f)
			default:
				return
			}
		}
	}()

	for {
		select {
		case f := <-r.chTaskBy:
			r.call(f)
		case f := <-r.chTask:
			r.call(f)
		case <-r.chClose:
			return
		}
	}
}

// FixedPool .
type FixedPool struct {
	wg      *sync.WaitGroup
	mux     sync.Mutex
	stopped int32

	chTask   chan func()
	chRunner chan struct{}
	chClose  chan struct{}

	runners []*fixedRunner
}

func (tp *FixedPool) push(f func()) error {
	select {
	case tp.chTask <- f:
	case <-tp.chClose:
		return ErrStopped
	}
	return nil
}

func (tp *FixedPool) pushByIndex(index int, f func()) error {
	r := tp.runners[uint32(index)%uint32(len(tp.runners))]
	select {
	case r.chTaskBy <- f:
	case <-tp.chClose:
		return ErrStopped
	}
	return nil
}

// Go .
func (tp *FixedPool) Go(f func()) error {
	if atomic.LoadInt32(&tp.stopped) == 1 {
		return ErrStopped
	}
	return tp.push(f)
}

// GoByIndex .
func (tp *FixedPool) GoByIndex(index int, f func()) error {
	if atomic.LoadInt32(&tp.stopped) == 1 {
		return ErrStopped
	}
	return tp.pushByIndex(index, f)
}

// Stop .
func (tp *FixedPool) Stop() {
	if atomic.CompareAndSwapInt32(&tp.stopped, 0, 1) {
		close(tp.chClose)
		tp.wg.Done()
		tp.wg.Wait()
	}
}

// NewFixedPool .
func NewFixedPool(size int, bufferSize int) *FixedPool {
	tp := &FixedPool{
		wg:      &sync.WaitGroup{},
		chTask:  make(chan func(), bufferSize),
		chClose: make(chan struct{}),
		runners: make([]*fixedRunner, size),
	}
	tp.wg.Add(1)

	for i := 0; i < size; i++ {
		r := &fixedRunner{
			wg:       tp.wg,
			chTask:   tp.chTask,
			chTaskBy: make(chan func(), bufferSize),
			chClose:  tp.chClose,
		}
		tp.runners[i] = r
		tp.wg.Add(1)
		go r.taskLoop()
	}

	return tp
}
