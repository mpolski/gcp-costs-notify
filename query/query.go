package query

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"

	"net/http"
	"os"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

func init() {

	functions.HTTP("query", query)
}

//query demonstrates issuing a query to billing dataset and summarizes total cost and credits for a specific day
func query(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	projectID := os.Getenv("PROJECT_ID")
	if projectID == "" {
		fmt.Println("PROJECT_ID environment variable must be set.")
		os.Exit(1)
	}
	dataSet := os.Getenv("DATASET")
	if dataSet == "" {
		fmt.Println("DATASET environment variable must be set.")
		os.Exit(1)
	}
	bqTable := os.Getenv("BQ_TABLE_NAME")
	if bqTable == "" {
		fmt.Println("BQ_TABLE_NAME environment variable must be set.")
		os.Exit(1)
	}
	location := os.Getenv("LOCATION")
	if location == "" {
		fmt.Println("LOCATION environment variable must be set.")
		os.Exit(1)
	}
	chatURL := os.Getenv("GOOGLE_CHAT_URL")
	if chatURL == "" {
		fmt.Println("GOOGLE_CHAT_URL environment variable must be set.")
		os.Exit(1)
	}

	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return
	}
	defer client.Close()

	fmt.Println("Querying....", projectID+"."+dataSet+"."+bqTable, " in:", location)

	q := client.Query(`
	    SELECT
			billing_account_id as billing_id,
			project.name,
			currency,
	        	SUM(cost) AS total_cost,
				SUM(IFNULL((SELECT SUM(c.amount) FROM UNNEST(credits) c), 0)) as total_credits,
				(SUM(cost)) + (SUM(IFNULL((SELECT SUM(c.amount) FROM UNNEST(credits) c), 0))) as after_credits
		FROM ` + projectID + "." + dataSet + "." + bqTable + `
	    WHERE DATE(_PARTITIONTIME) = @date
		GROUP BY billing_account_id, project.name, currency
		ORDER BY total_cost DESC
		LIMIT 1000;
		`)

	q.Location = location
	q.Parameters = []bigquery.QueryParameter{
		{
			Name:  "date",
			Value: time.Now().AddDate(0, 0, -1).Format("2006-01-02"),
		},
	}

	// Run the query and print results when the query job is completed.
	job, err := q.Run(ctx)
	if err != nil {
		return
	}
	status, err := job.Wait(ctx)
	if err != nil {
		return
	}
	if err := status.Err(); err != nil {
		return
	}
	it, err := job.Read(ctx)
	for {
		var row []bigquery.Value
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return
		}

		//Building and sending Google Workspace Chat message
		bID := fmt.Sprintf("%s", row[0])
		pName := fmt.Sprintf("%s", row[1])
		curr := fmt.Sprintf("%s", row[2])
		cost := fmt.Sprintf("%.2f", row[3])
		credit := fmt.Sprintf("%.2f", row[4])
		total := fmt.Sprintf("%.2f", row[5])

		msgStr := fmt.Sprintf(`{"text":"*Project: %s* \n[Billing account: %s]\nCost: %s %s , Credits applied: %s %s \nTotal: %s %s\n"}`, pName, bID, cost, curr, credit, curr, total, curr)

		var jsonStr = []byte(msgStr)

		req, err := http.NewRequest("POST", chatURL, bytes.NewBuffer(jsonStr))
		if err != nil {
			return
		}
		req.Header.Set("Content-Type", "application/json; charset=utf-8")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		fmt.Println("response Status:", resp.Status)
		fmt.Println("response Headers:", resp.Header)
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("response Body:", string(body))
	}
	if err != nil {
		log.Fatalln("Query failed:", err)
	}
}
