package safe_close

import (
	"errors"
	"testing"
	"time"
)

func TestSendCloseSignalUnblocksWaitClosed(t *testing.T) {
	sc := NewSafeClose()
	closed := make(chan struct{})
	go func() {
		_ = sc.WaitClosed()
		close(closed)
	}()

	sc.SendCloseSignal(nil)

	select {
	case <-closed:
	case <-time.After(time.Second):
		t.Fatal("WaitClosed 未在收到关闭信号后返回")
	}
}

func TestSendCloseSignalOnceAndKeepsErr(t *testing.T) {
	sc := NewSafeClose()
	want := errors.New("boom")

	sc.SendCloseSignal(want)
	sc.SendCloseSignal(errors.New("ignored")) // 第二次调用应为 noop

	if got := sc.WaitClosed(); !errors.Is(got, want) {
		t.Fatalf("WaitClosed 返回 %v，期望 %v", got, want)
	}
}

func TestAttachWaitsSubTaskExit(t *testing.T) {
	sc := NewSafeClose()
	entered := make(chan struct{})

	sc.Attach(func(done func(), signal <-chan struct{}) {
		close(entered)
		<-signal // 收到关闭信号后做清理
		done()
	})
	<-entered // 确保子任务已启动

	exited := make(chan struct{})
	go func() {
		_ = sc.WaitClosed()
		close(exited)
	}()
	sc.SendCloseSignal(nil)

	select {
	case <-exited:
	case <-time.After(time.Second):
		t.Fatal("WaitClosed 未等待 Attach 的子任务退出")
	}
}
