Backoff library
=============
[![Go Report Card](https://goreportcard.com/badge/github.com/go-playground/backoff)](https://goreportcard.com/report/github.com/go-playground/backoff)
[![GoDoc](https://godoc.org/github.com/go-playground/backoff?status.svg)](https://godoc.org/github.com/go-playground/backoff)
![License](https://img.shields.io/dub/l/vibe-d.svg)

Backoff library uses an exponential backoff algorithm to backoff between retries.

What makes this different from other backoff libraries?
1. Simple, by automatically calculating the exponential factor between the min and max backoff times; which properly tunes to your desired values.
2. Provides an optional AutoTune function which will adjust the min backoff duration based upon successful backoff durations.

Why AutoTune?

For long running services that hit external services such as writing to a DB or that hit a 3rd party API's, where successful attempts backoff durations can vary over time as load changes. Additionally by using auto it should provide an automatic jitter when multiple copies of a service are running.

Example Basic
------------
```go
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
```

Example AutoTune
------------
```go
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
						return false, nil
					case 4:
						return false, errors.New("ERR")
					default: //case 5:
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
```

Package Versioning
---------------
I'm jumping on the vendoring bandwagon, you should vendor this package as I will not
be creating different version with gopkg.in like allot of my other libraries.

Why? because my time is spread pretty thin maintaining all of the libraries I have + LIFE,
it is so freeing not to worry about it and will help me keep pouring out bigger and better
things for you the community.

License
------
Distributed under MIT License, please see license file in code for more details.