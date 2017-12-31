package main

import (
	"errors"
	"time"

	"fmt"

	"github.com/go-playground/backoff"
)

func main() {
	bo := backoff.New(5, time.Second, time.Minute)
	defer bo.Close()

	// retry function and notification function to log failures etc...
	bo.Run(func() (bail bool, err error) {
		// do something
		return false, errors.New("ERR")
	}, func(attempt uint16, waiting time.Duration, err error) {
		fmt.Printf("Attempt:%d Waiting:%s Error:%s\n", attempt, waiting, err)

	})
}
