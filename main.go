package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const workerCount = 100

func checkPort(portChan chan int, wg *sync.WaitGroup, client *http.Client) {
	defer wg.Done()

	for port := range portChan {
		address := "10.49.122.144:" + strconv.Itoa(port)
		conn, err := net.DialTimeout("tcp", address, time.Millisecond*10)
		if err != nil {
			continue
		}
		conn.Close()

		url := fmt.Sprintf("http://10.49.122.144:%d/ping", port)
		resp, err := client.Get(url)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			continue
		}
		fmt.Printf("Response from port %d: %s\n", port, string(body)) // Debugging print

		if resp.StatusCode == http.StatusOK {

			fmt.Println("Success:", resp.StatusCode)
		} else {
			fmt.Println("Failed:", resp.StatusCode)
		}
	}
}

func main() {
	client := &http.Client{
		Timeout: time.Second * 2,
	}

	for {
		var wg sync.WaitGroup
		portChan := make(chan int, workerCount)

		for i := 0; i < workerCount; i++ {
			wg.Add(1)
			go checkPort(portChan, &wg, client)
		}

		for port := 1; port <= 65535; port++ {
			portChan <- port
		}

		close(portChan)
		wg.Wait()
	}
}
