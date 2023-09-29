package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const workerCount = 100

type UserRequest struct {
	User   string      `json:"User"`
	Secret interface{} `json:"Secret,omitempty"` // Allow for different data types for the Secret
}

func checkPort(portChan chan int, wg *sync.WaitGroup, client *http.Client, user *string, secret *interface{}) {
	defer wg.Done()

	for port := range portChan {
		address := "10.49.122.144:" + strconv.Itoa(port)
		conn, err := net.DialTimeout("tcp", address, time.Millisecond*100)

		if err != nil {
			continue
		}
		conn.Close()

		baseURL := fmt.Sprintf("http://10.49.122.144:%d", port)

		signUpURL := baseURL + "/signup"
		signUpBody, _ := json.Marshal(UserRequest{User: *user})
		doPost(client, signUpURL, signUpBody)

		checkURL := baseURL + "/check"
		checkBody, _ := json.Marshal(UserRequest{User: *user})
		doPost(client, checkURL, checkBody)

		for i := 0; i < 15; i++ {
			getUserSecretURL := baseURL + "/getUserSecret"
			getUserSecretBody, _ := json.Marshal(UserRequest{User: *user})
			resp, err := client.Post(getUserSecretURL, "application/json", bytes.NewBuffer(getUserSecretBody))
			if err != nil {
				fmt.Println("Post error:", err)
				continue
			}
			defer resp.Body.Close()
			respBody, _ := io.ReadAll(resp.Body)
			respString := string(respBody)
			fmt.Printf("Iteration %d, Response from %s: %s\n", i, getUserSecretURL, respString) // Displaying the response
			if len(respString) > 45 {
				*secret = strings.TrimSpace(strings.Split(respString, ": ")[1]) // Extracting the secret
				break                                                           // Exiting the loop once the secret is obtained
			}
		}

		getUserLevelURL := baseURL + "/getUserLevel"
		getUserLevelBody, _ := json.Marshal(UserRequest{User: *user, Secret: *secret})
		doPost(client, getUserLevelURL, getUserLevelBody)

		getUserPointsURL := baseURL + "/getUserPoints"
		getUserPointsBody, _ := json.Marshal(UserRequest{User: *user, Secret: *secret})
		doPost(client, getUserPointsURL, getUserPointsBody)

		getHintURL := baseURL + "/iNeedAHint"
		getHintBody, _ := json.Marshal(UserRequest{User: *user, Secret: *secret})
		doPost(client, getHintURL, getHintBody)

		enterChallengeURL := baseURL + "/enterChallenge"
		enterChallengeBody, _ := json.Marshal(UserRequest{User: *user, Secret: *secret})
		doPost(client, enterChallengeURL, enterChallengeBody)
	}
}

func doPost(client *http.Client, url string, body []byte) {
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("Post error:", err)
		return
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	fmt.Printf("Response from %s: %s\n", url, string(respBody))
}

func main() {
	client := &http.Client{
		Timeout: time.Second * 2,
	}

	user := "724490"
	var secret interface{} = "someSecret"

	for {
		var wg sync.WaitGroup
		portChan := make(chan int, workerCount)

		for i := 0; i < workerCount; i++ {
			wg.Add(1)
			go checkPort(portChan, &wg, client, &user, &secret)
		}

		for port := 1; port <= 65535; port++ {
			portChan <- port
		}

		close(portChan)
		wg.Wait()
	}
}
