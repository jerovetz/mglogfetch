package fetcher

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Paging struct {
	Previous string
	First    string
	Last     string
	Next     string
}

type ResponseChecker struct {
	Items []Item
}

type Response struct {
	Items  []json.RawMessage
	Paging Paging
}

type Item struct {
	Timestamp float32
}

type HttpClientInterface interface {
	Do(req *retryablehttp.Request) (*http.Response, error)
}

type ClockInterface interface {
	Now() TimeInterface
	Sleep(d time.Duration)
}

type TimeInterface interface {
	Unix() int64
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func tryToFetch(url string, client HttpClientInterface) []byte {
	request, _ := retryablehttp.NewRequest("GET", url, nil)
	request.Header.Add("Authorization", "Basic "+basicAuth(os.Getenv("MAILGUN_API_USERNAME"), os.Getenv("MAILGUN_API_SECRET")))

	response, err := client.Do(request)

	if err != nil {
		panic(fmt.Sprintf( "Client failed completely, stop now. Url was %s. %s", url, err))
	}

	if response.StatusCode != 200 {
		panic(fmt.Sprintf("Statuscode was %d, url is %s", response.StatusCode, url))
	}

	bodyBytes, _ := ioutil.ReadAll(response.Body)
	var formatBuffer bytes.Buffer
	json.Compact(&formatBuffer, bodyBytes)
	formatted, _ := ioutil.ReadAll(&formatBuffer)
	return formatted
}

func Fetch(url string, client HttpClientInterface, clock ClockInterface) Response {
	var response Response
	var checker ResponseChecker
	var body []byte
	var checkTime = int64(0)
	var necessaryWaitTime = int64(0)
	threshold, _ := strconv.Atoi(os.Getenv("OLD_THRESHOLD_SECONDS"))

	for retryNeeded(checker, checkTime, necessaryWaitTime) {
		body = tryToFetch(url, client)
		json.Unmarshal(body, &checker)

		checkTime = getCheckTime(checker, checkTime)
		necessaryWaitTime = clock.Now().Unix() - int64(threshold)

		if retryNeeded(checker, checkTime, necessaryWaitTime) {
			clock.Sleep(10 * time.Second)
		}
	}

	json.Unmarshal(body, &response)
	return response
}

func retryNeeded(checker ResponseChecker, checkTime int64, necessaryWaitTime int64) bool {
	return len(checker.Items) == 0 || checkTime > necessaryWaitTime
}

func getCheckTime(checker ResponseChecker, checkTime int64) int64 {
	if len(checker.Items) - 1 >= 0 {
		checkTime = int64(checker.Items[len(checker.Items)-1].Timestamp)
	} else {
		checkTime = 0
	}
	return checkTime
}
