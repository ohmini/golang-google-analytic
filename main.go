package main

import (
	"fmt"
	"io/ioutil"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/analytics/v3"
)

func main() {
	// key.json = google service accont
	key, _ := ioutil.ReadFile("key.json")

	jwtConf, err := google.JWTConfigFromJSON(
		key,
		analytics.AnalyticsReadonlyScope,
	)
	p(err)

	httpClient := jwtConf.Client(oauth2.NoContext)
	svc, err := analytics.New(httpClient)
	p(err)

	fmt.Println("retrieve the real time users of this profile")

	metrics := "rt:pageviews"
	dimensions := "rt:minutesAgo,rt:pagePath"
	query := svc.Data.Realtime.Get("ga:214630239", metrics)
	query.Dimensions(dimensions)
	query.Sort("rt:minutesAgo")

	rt, err := query.Do()
	p(err)

	fmt.Println(rt.Rows)
}

func p(err error) {
	if err != nil {
		panic(err)
	}
}