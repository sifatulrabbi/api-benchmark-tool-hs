package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

const ENV_FILE_PATH = ".env"
const TEST_DURATION = time.Second * 30
const MAX_SLEEP_THRESHOLD = 1000
const ACTIVE_USERS = 50

type TestContext struct {
	UserID      string
	AccessToken string
	UserEmail   string
	AccountID   string
}
type TestResult struct {
	TotalUsersHandled   int
	TotalRequestHandled int
	UsersActive         int

	UsersChan    chan int
	RequestsChan chan int
}

var results = TestResult{
	TotalUsersHandled:   0,
	TotalRequestHandled: 0,
	UsersActive:         0,

	UsersChan:    make(chan int),
	RequestsChan: make(chan int),
}

func main() {
	err := godotenv.Load(ENV_FILE_PATH)
	if err != nil {
		log.Fatalln("unable to load the .env file")
	}

	stopSig := make(chan bool, 1)
	wg := sync.WaitGroup{}

	go func() {
		time.Sleep(TEST_DURATION)
		stopSig <- true
	}()

	go func() {
		for range results.UsersChan {
			results.TotalUsersHandled += 1
		}
	}()

	go func() {
		for range results.RequestsChan {
			results.TotalRequestHandled += 1
		}
	}()

	log.Println("Test started")
	go createActiveUsers(&wg, stopSig, ACTIVE_USERS)

	<-stopSig
	wg.Wait()
	log.Println("Test completed")
	fmt.Printf(
		"Total users: %d\nTotal requests: %d\n",
		results.TotalUsersHandled,
		results.TotalRequestHandled,
	)
}

func createActiveUsers(mainWg *sync.WaitGroup, stopSig chan bool, n int) {
	mainWg.Add(1)

	wg := sync.WaitGroup{}
	done := false

	go func() {
		<-stopSig
		done = true
	}()

	for i := range make([]int, n) {
		go func(i int) {
			wg.Add(1)
			results.UsersChan <- 1
			for !done {
				waitForRandomDuration()
				makeHttpRequests()
			}
			wg.Done()
		}(i)
	}

	wg.Wait()
	mainWg.Done()
}

func waitForRandomDuration() {
	n := rand.Intn(int(MAX_SLEEP_THRESHOLD))
	time.Sleep(time.Duration(n) * time.Millisecond)
}

func makeHttpRequests() {
	testCtx := TestContext{
		UserID:      os.Getenv("TEST_USER_ID"),
		AccountID:   os.Getenv("TEST_USER_ACCOUNT_ID"),
		UserEmail:   os.Getenv("TEST_USER_EMAIL"),
		AccessToken: os.Getenv("TEST_ACCESS_TOKEN"),
	}
	if testCtx.AccessToken == "" {
		log.Fatalln("no access token found to use with the http requests.")
	}

	updateNoteTitle(testCtx)
	results.RequestsChan <- 1
	updateProjectName(testCtx)
	results.RequestsChan <- 1
	updateTemplateName(testCtx)
	results.RequestsChan <- 1
}

func updateProjectName(ctx TestContext) {
	payload := map[string]string{
		"name": "Test project updated",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		log.Fatalln(err)
	}
	url := os.Getenv("API_BASE_URL")
	url += "/api/project/64ddf79ab40bc5668f46b46d" // test project on localhost:8080
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Set("Authorization", "Bearer "+ctx.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	c := http.Client{}
	res, err := c.Do(req)
	if err != nil {
		log.Fatalln("can't update project", err)
	}
	defer res.Body.Close()
}

func updateNoteTitle(ctx TestContext) {
	url := os.Getenv("API_BASE_URL")
	url += "/api/notes/652f9ad5d769854a8c558d97"
	payload := map[string]string{
		"slug": "Test note for helloscribe. Test ID - 1990 | updated",
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Set("Authorization", "Bearer "+ctx.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	c := http.Client{}
	res, err := c.Do(req)
	if err != nil {
		log.Fatalln("can't update notes", err)
	}
	defer res.Body.Close()
}

func updateTemplateName(ctx TestContext) {
	url := os.Getenv("API_BASE_URL")
	url += "/api/templates/652fe87c1bcf5f9e5674d7e0"
	payload := map[string]string{
		"name": "Test template name | updated",
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Set("Authorization", "Bearer "+ctx.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	c := http.Client{}
	res, err := c.Do(req)
	if err != nil {
		log.Fatalln("can't update templates", err)
	}
	defer res.Body.Close()
}
