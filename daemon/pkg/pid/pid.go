package pid

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"syscall"
)

type ProcessLockFile interface {
	Unlock() error
}

type processLockFile struct {
	f *os.File
}

func (p processLockFile) Unlock() error {
	path := p.f.Name()
	if err := p.f.Close(); err != nil {
		return err
	}

	return os.Remove(path)
}

func AcquireProcessIDLock(pidFilePath string) (ProcessLockFile, error) {
	if _, err := os.Stat(pidFilePath); !os.IsNotExist(err) {
		raw, err := os.ReadFile(pidFilePath)
		if err != nil {
			return nil, err
		}

		pid, err := strconv.Atoi(string(raw))
		if err != nil {
			return nil, err
		}

		if proc, err := os.FindProcess(int(pid)); err == nil && !errors.Is(proc.Signal(syscall.Signal(0)), os.ErrProcessDone) {
			return nil, fmt.Errorf("process %d is already running", proc.Pid)
		} else if err = os.Remove(pidFilePath); err != nil {
			return nil, err

		}
	}

	f, err := os.OpenFile(pidFilePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	if _, err := f.Write([]byte(fmt.Sprint(os.Getpid()))); err != nil {
		return nil, err
	}

	return processLockFile{f: f}, nil
}
