package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"errors"
	"log"
	"os"
	"time"
	"strconv"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/analytics/v3"
	_ "github.com/lib/pq"
)

func main() {
	// key.json = google service accont
	db, err := connectDb()
	if err != nil {
		log.Println("Cannot connect to Postgres. Abort.")
		os.Exit(1)
	}

	defer db.Close()

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
	query := svc.Data.Realtime.Get("ga:214613684", metrics)
	query = query.Dimensions(dimensions)
	query = query.Sort("rt:minutesAgo")

	rt, err := query.Do()
	p(err)

	fmt.Println(rt.Rows)
	calPageviews(rt.Rows)
}

func calPageviews(data [][]string ) {
	views := make(map[string]int)
	for i:= range data {
		title := data[i][1]
		

		pageviews := data[i][2]
		v, err := strconv.Atoi(pageviews)
		if err == nil {
			total, ok := views[title]
			if ok {
				views[title] = total + v
			} else {
				views[title] = v
			}
		}
	}
	fmt.Println(views)
}

func p(err error) {
	if err != nil {
		panic(err)
	}
}

func connectDb() (*sql.DB, error) {
	failureCount := 0
	for {
		db, err := sql.Open(
			"postgres",
			os.ExpandEnv("host=localhost user=postgres dbname=true4u sslmode=disable"),
		)

		if err != nil {
			failureCount++

			if failureCount > 30 {
				log.Printf(
					"Postgres is not avaiable longer than 1 minute. Give up.\n",
				)
				return nil, errors.New("Cannot connect to Posgres")
			}

			log.Printf(
				"Postgres is not available now. Sleep. (Error> %v)\n",
				err,
			)
			time.Sleep(time.Second * 2)
			continue
		}

		db.SetMaxOpenConns(32)
		db.SetMaxIdleConns(8)
		db.SetConnMaxLifetime(time.Hour)

		log.Printf(
			"Postgres is ready.\n",
		)
		return db, nil
	}
}