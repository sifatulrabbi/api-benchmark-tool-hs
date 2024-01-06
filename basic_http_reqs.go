package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// test any given endpoint with 10-100 concurrent HTTP calls.
func basicHTTPReqs(lim int, method, url string) {
	startTime := time.Now()
	wg := sync.WaitGroup{}

	for i := 0; i < lim; i++ {
		go func() {
			wg.Add(1)
			req, err := http.NewRequest(method, url, http.NoBody)
			if err != nil {
				log.Fatalln("Unable to make the http call", err)
			}

			res, err := http.DefaultClient.Do(req)
			fmt.Println(res.Status)
			wg.Done()
		}()
	}

	wg.Wait()

	endTime := time.Now()
	diff := endTime.UnixMilli() - startTime.UnixMilli()
	diff = diff / 1000 // converting in seconds

	fmt.Println("total time required:", diff)
}
