// Package shutdown allows specifying functions to run before application exits.
package shutdown

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
)

var h = newHandler()

type handler struct {
	active atomic.Value
	mtx    sync.Mutex
	stack  []func()
}

func newHandler() *handler {
	h := &handler{}
	h.active.Store(false)
	go h.wait()
	return h
}

// IsActive returns whether the exit handler is currently executing.
func IsActive() bool {
	return h.active.Load().(bool)
}

// BeforeExit specifies a function to run before exit.
func BeforeExit(f func()) {
	h.mtx.Lock()
	h.stack = append(h.stack, f)
	h.mtx.Unlock()
}

// Exit runs shutdown functions and exits.
func Exit() {
	h.exit(nil, 0, recover())
}

// ExitWithCode runs shutdown functions and exits with specified code.
func ExitWithCode(code int) {
	h.exit(nil, code, recover())
}

// Fatal is the equivalent of fmt.Fatal but runs shutdown functions first.
func Fatal(v ...interface{}) {
	h.exit(errors.New(fmt.Sprint(v...)), 1, recover())
}

// Fatalf is the equivalent of fmt.Fatalf but runs shutdown functions first.
func Fatalf(format string, v ...interface{}) {
	h.exit(fmt.Errorf(format, v...), 1, recover())
}

func (h *handler) wait() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, os.Signal(syscall.SIGTERM))
	<-ch
	h.exit(nil, 0, nil)
}

func (h *handler) exit(err error, code int, serious interface{}) {
	h.mtx.Lock()
	h.active.Store(true)
	for i := len(h.stack) - 1; i >= 0; i-- {
		h.stack[i]()
	}
	if serious != nil {
		panic(serious)
	}
	if err != nil {
		log.New(os.Stderr, "", log.Lshortfile|log.Lmicroseconds).Output(3, err.Error())
	}
	os.Exit(code)
}
