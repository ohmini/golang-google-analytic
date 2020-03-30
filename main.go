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
	"strings"

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

	updateEvery10Minutes(query, db)
}

func updateEvery10Minutes(query *analytics.DataRealtimeGetCall, db *sql.DB) {
	for t := range time.Tick(time.Minute) {
		_ = t
		updatePageviews(query, db)
	}
}

func updatePageviews(query *analytics.DataRealtimeGetCall, db *sql.DB) {
	rt, err := query.Do()
	p(err)

	views := calPageviews(rt.Rows)

	for k, v:= range views {
		updateOrInsertViews(k, v, db)
	}
}

func updateOrInsertViews(id string, views int, db *sql.DB) error {
	// use function for insert or update rows
	query := `
		SELECT update_content_analytics($1, $2);
	`
	_, err := db.Query(query, id, views)

	if err != nil {
		return fmt.Errorf("Error database query execution> %v", err)
	}

	fmt.Println("update success.")
	return nil

}  

func calPageviews(data [][]string ) map[string]int {
	views := make(map[string]int)
	for i:= range data {
		title := data[i][1]
		paths := strings.Split(strings.TrimSpace(title), "/")
		if paths[1] != "uncategorized" {
			continue
		}
		
		pageviews := data[i][2]
		v, err := strconv.Atoi(pageviews)
		if err == nil {
			total, ok := views[paths[2]]
			if ok {
				views[paths[2]] = total + v
			} else {
				views[paths[2]] = v
			}
		}
	}
	return views
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
			os.ExpandEnv("host=localhost user=postgres password=1234 dbname=true4u sslmode=disable"),
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