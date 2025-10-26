package shell

import (
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

type Shell struct {
	currentProcesses []*exec.Cmd
	mu               sync.Mutex
}

func (s *Shell) AddProcess(cmd *exec.Cmd) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentProcesses = append(s.currentProcesses, cmd)
}

func (s *Shell) RemoveProcess(cmd *exec.Cmd) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, p := range s.currentProcesses {
		if p == cmd {
			s.currentProcesses = append(s.currentProcesses[:i], s.currentProcesses[i+1:]...)
			break
		}
	}
}

func (s *Shell) KillAllProcesses() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, p := range s.currentProcesses {
		if p.Process != nil {
			// Мягкое завершение
			p.Process.Signal(syscall.SIGTERM)

			// Жесткое завершение через 2 секунды если не ответил
			go func(proc *os.Process) {
				time.Sleep(2 * time.Second)
				proc.Kill()
			}(p.Process)
		}
	}
	s.currentProcesses = nil
}
