package main

import (
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/joho/godotenv"
	"log"
	"matchwork/mailgun-log-fetcher/fetcher"
	pusherPack "matchwork/mailgun-log-fetcher/pusher"
	"os"
	"strconv"
	"time"
)

var fetchAction = fetcher.Fetch
var pusherCreator = pusherPack.New

var clock = &RealClock{}
var now = clock.Now().Unix()

const mailgunEuDomain = "https://api.eu.mailgun.net/v3/"
const mailgunUsDomain = "https://api.mailgun.net/v3/"

type RealClock struct {
}

func (c *RealClock) Now() fetcher.TimeInterface {
	return time.Now()
}

func (c *RealClock) Sleep(d time.Duration) {
	time.Sleep(d)
}

func init() {
	dotenvErr := godotenv.Load()
	if dotenvErr != nil {
		log.Println("Error loading env")
	}

}

func getMailgunDomain() string {
	switch os.Getenv("MAILGUN_REGION") {
		case "eu": return mailgunEuDomain
		case "us": return mailgunUsDomain
	}
	panic(fmt.Sprintf("no url for current region setting: %s", os.Getenv("MAILGUN_REGION")))
}

func main() {
	var response fetcher.Response
	url := fmt.Sprintf("%s%s/events?begin=%s&ascending=yes", getMailgunDomain(), os.Getenv("MAIL_DOMAIN"), strconv.FormatInt(now, 10))
	var client = retryablehttp.NewClient()

	for true {
		response = fetchAction(url, client, clock)
		pusher := pusherCreator()
		pusher.Push(response.Items)
		url = response.Paging.Next
	}
}
