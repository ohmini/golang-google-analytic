package main

import (
	"fmt"
	"io/ioutil"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/analytics/v3"
)

func main() {

	key, _ := ioutil.ReadFile("key.json")

	jwtConf, err := google.JWTConfigFromJSON(
		key,
		analytics.AnalyticsReadonlyScope,
	)
	p(err)

	httpClient := jwtConf.Client(oauth2.NoContext)
	svc, err := analytics.New(httpClient)
	p(err)

	accountResponse, err := svc.Management.Accounts.List().Do()
	p(err)

	var accountId string
	fmt.Println(accountId)
	fmt.Println("Found the following accounts:")
	for i, acc := range accountResponse.Items {

		if i == 0 {
			accountId = acc.Id
		}

		fmt.Println(acc.Id, acc.Name)
	}

	webProps, err := svc.Management.Webproperties.List(accountId).Do()
	p(err)

	var wpId string

	fmt.Println("\nFound the following properties:")
	for i, wp := range webProps.Items {

		if i == 0 {
			wpId = wp.Id
		}

		fmt.Println(wp.Id, wp.Name)
	}

	profiles, err := svc.Management.Profiles.List(accountId, wpId).Do()
	p(err)

	var viewId string

	fmt.Println("\nFound the following profiles:")
	for i, p := range profiles.Items {

		if i == 0 {
			viewId = "ga:" + p.Id
		}

		fmt.Println(p.Id, p.Name)
	}

	fmt.Println("\nTime to retrieve the real time users of this profile")

	metrics := "rt:activeUsers"
	rt, err := svc.Data.Realtime.Get(viewId, metrics).Do()
	p(err)

	fmt.Println(rt.Rows)
}

func p(err error) {
	if err != nil {
		panic(err)
	}
}