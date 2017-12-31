package backoff

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestBackoff(t *testing.T) {
	bo := New(5, time.Millisecond*100, time.Second)
	defer bo.Close()
	err := bo.Run(func() (bool, error) {
		return false, errors.New("ERR")
	}, func(attempt uint16, waiting time.Duration, err error) {
		fmt.Printf("Attempt:%d Waiting:%s Error:%s\n", attempt, waiting, err)
	})
	if err == nil {
		t.Fatalf("Expected %v Got %s", nil, err)
	}
}

func TestBackoffAutoAdjust(t *testing.T) {
	bo := New(5, time.Millisecond*100, time.Second)
	bo.AutoTune(time.Millisecond*400, time.Millisecond*500)
	defer bo.Close()
	err := bo.Run(func() (bool, error) {
		return false, errors.New("ERR")
	}, func(attempt uint16, waiting time.Duration, err error) {
		fmt.Printf("Attempt:%d Waiting:%s Error:%s\n", attempt, waiting, err)
	})
	if err == nil {
		t.Fatalf("Expected %v Got %s", nil, err)
	}
}

func TestBackoffAutoAdjustParallel(t *testing.T) {
	bo := New(5, time.Millisecond*100, time.Second)
	bo.AutoTune(time.Millisecond*400, time.Millisecond*500)
	defer bo.Close()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := bo.Run(func() (bool, error) {
				return false, errors.New("ERR")
			}, func(attempt uint16, waiting time.Duration, err error) {
				fmt.Printf("Attempt:%d Waiting:%s Error:%s\n", attempt, waiting, err)
			})
			if err == nil {
				t.Fatalf("Expected %v Got %s", nil, err)
			}
		}()
	}

	wg.Wait()
}
