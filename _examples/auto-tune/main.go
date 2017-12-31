package main

import (
	"errors"
	"fmt"
	"time"

	"os"

	"os/signal"
	"syscall"

	"math/rand"

	"github.com/go-playground/backoff"
)

func main() {
	bo := backoff.New(5, time.Second, 10*time.Second)
	bo.AutoTune(30*time.Second, 2*time.Minute)
	defer bo.Close()

	go func() {
		for {
			time.Sleep(time.Millisecond * 500)
			go func() {
				bo.Run(func() (bool, error) {
					// simulating random success/failure
					count := rand.Intn(5)
					switch count {
					case 1:
						return false, errors.New("ERR")
					case 2:
						return false, errors.New("ERR")
					case 3:
						count = 0
						return false, nil
					case 4:
						return false, errors.New("ERR")
					default: //case 5:
						count = 0
						return false, nil
					}
				}, func(attempt uint16, waiting time.Duration, err error) {
					fmt.Printf("Attempt:%d Waiting: %s Error:%s\n", attempt, waiting, err)
				})
			}()
		}
	}()
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)
	<-s
}
