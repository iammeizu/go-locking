package locking_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/tgulacsi/go-locking"
)

func TestFLock(t *testing.T) {
	fh, err := ioutil.TempFile("", "lock-test.")
	if err != nil {
		t.Fatal(err)
	}
	fh.Close()
	defer os.Remove(fh.Name())

	flock, err := locking.NewFLock(fh.Name())
	if err != nil {
		t.Fatal(err)
	}
	if err := testLock(flock); err != nil {
		t.Fatal(err)
	}
}

func TestPortLock(t *testing.T) {
	port := 1337
	for port < 65535 {
		lock := locking.NewPortLock(port)
		if ok, _ := lock.TryLock(); ok {
			lock.Unlock()
			break
		}
		port++
	}
	t.Logf("port=%d", port)
	lock := locking.NewPortLock(port)
	if err := testLock(lock); err != nil {
		t.Fatal(err)
	}
}

type locker interface {
	Lock() error
	Unlock() error
}

// test the lock.
//
// FIXME(tgulacsi): to test IPC locks, a separate process should be run.
func testLock(lock locker) error {
	tryLock, isTryLocker := lock.(interface {
		TryLock() (bool, error)
	})
	if isTryLocker {
		if err := func() error {
			ok, err := tryLock.TryLock()
			if err != nil {
				return err
			}
			if !ok {
				return fmt.Errorf("lock %v is already locked!", lock)
			}
			return lock.Unlock()
		}(); err != nil {
			return err
		}
	}
	if err := lock.Lock(); err != nil {
		return err
	}
	return lock.Unlock()
}
