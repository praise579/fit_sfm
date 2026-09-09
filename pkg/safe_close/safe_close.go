// 包 safe_close 提供优雅关闭协调器：WaitClosed 会等待所有受监管的子
// goroutine 退出后才返回，保证服务退出时清理逻辑全部完成。
//
// 典型用法：
//  1. 主服务启动后阻塞在 WaitClosed() 上。
//  2. 需要受生命周期监管的子 goroutine 一律通过 Attach 启动，
//     内部监听关闭信号并做清理后调用 done()。
//  3. 任一环节发生致命错误时，调用 SendCloseSignal(err) 触发整体关闭。
//  4. 外部调用方也可以用 SendCloseSignal 主动请求停止服务。
package safe_close

import "sync"

// SafeClose 协调「发出关闭信号 -> 等待子任务退出」的关闭流程。
type SafeClose struct {
	mu     sync.Mutex
	wg     sync.WaitGroup
	signal chan struct{} // 关闭信号：一旦 close，所有监听方立即感知
	err    error         // 触发关闭时携带的错误
	done   bool          // 是否已触发关闭（保证只生效一次）
}

// NewSafeClose 构造一个新的关闭协调器。
func NewSafeClose() *SafeClose {
	return &SafeClose{signal: make(chan struct{})}
}

// WaitClosed 阻塞直到收到关闭信号，并等待所有 Attach 的子任务结束，随后返回关闭原因。
func (s *SafeClose) WaitClosed() error {
	<-s.signal
	s.wg.Wait()
	return s.err
}

// SendCloseSignal 触发一次关闭，解除 WaitClosed 的阻塞；重复调用无副作用。
func (s *SafeClose) SendCloseSignal(err error) {
	s.mu.Lock()
	if !s.done {
		s.done = true
		s.err = err
		close(s.signal)
	}
	s.mu.Unlock()
}

// ReceiveCloseSignal 返回只读的关闭信号通道。
func (s *SafeClose) ReceiveCloseSignal() <-chan struct{} {
	return s.signal
}

// Attach 启动一个受监管的子任务：f 收到关闭信号后应执行清理并调用 done 结束。
// 若协调器已被关闭则不再启动新任务。
func (s *SafeClose) Attach(f func(done func(), closeSignal <-chan struct{})) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.done {
		return
	}
	s.wg.Add(1)
	go f(s.wg.Done, s.signal)
}
