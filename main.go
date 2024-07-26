package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	_ "github.com/lib/pq"
)

func Getenv(key string, d string) string {
	s, ok := os.LookupEnv(key)
	if ok {
		return s
	}
	return d
}

const charSet = "UTF-8"

type htmlbuilder struct {
	strings.Builder
}

type attribute struct {
	key   string
	value string
}

var alignRight = attribute{key: "align", value: "right"}
var alignLeft = attribute{key: "align", value: "left"}
var alignCenter = attribute{key: "align", value: "left"}

func (h *htmlbuilder) WriteOpenTag(tag string, aa ...attribute) {
	h.WriteString("<")
	h.WriteString(tag)
	for _, a := range aa {
		h.WriteString(" ")
		h.WriteString(a.key)
		h.WriteString("=")
		h.WriteString(a.value)
	}
	h.WriteString(">")
}

func (h *htmlbuilder) WriteCloseTag(tag string) {
	h.WriteString("</")
	h.WriteString(tag)
	h.WriteString(">\n")
}

func main() {

	// Database connection settings
	dbHost := Getenv("DB_HOST", "localhost")
	dbPort := Getenv("DB_PORT", "5432")
	dbUser := Getenv("DB_USER", "postgres")
	dbPassword := Getenv("DB_PASSWORD", "password")
	dbName := Getenv("DB_NAME", "postgres")
	query := Getenv("QUERY", "select count(*) count, date(create_date) create_date from members group by 2 order by 2 desc limit 10")
	email := os.Getenv("EMAIL")
	region := Getenv("REGION", "us-west-2")
	sender := Getenv("SENDER", "constantinou@gmail.com")

	// Construct the PostgreSQL connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPassword, dbName)

	// Connect to the PostgreSQL database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer db.Close()

	// Execute the query
	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("Failed to execute the query: %v", err)
	}
	defer rows.Close()

	// Fetch column names
	columns, err := rows.Columns()
	if err != nil {
		log.Fatalf("Failed to fetch column names: %v", err)
	}

	var htmlResult htmlbuilder

	htmlResult.WriteOpenTag("table")

	const stringFormat = "%25s"
	const variableFormat = "%25v"
	// Prepare result string
	var result strings.Builder
	htmlResult.WriteOpenTag("tr")

	for _, column := range columns {
		result.WriteString(fmt.Sprintf(stringFormat, column))
		htmlResult.WriteOpenTag("th", alignCenter)
		htmlResult.WriteString(column)
		htmlResult.WriteCloseTag("th")
	}
	htmlResult.WriteCloseTag("tr")
	result.WriteString("\n")

	// Process rows
	for rows.Next() {
		columnPointers := make([]interface{}, len(columns))
		columnValues := make([]interface{}, len(columns))
		for i := range columnValues {
			columnPointers[i] = &columnValues[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			log.Fatalf("Failed to scan row: %v", err)
		}
		htmlResult.WriteOpenTag("tr")

		for _, value := range columnValues {
			if value == nil {
				result.WriteString(fmt.Sprintf(stringFormat, "-"))
				htmlResult.WriteOpenTag("td", alignCenter)
				htmlResult.WriteString("-")

			} else if t, ok := value.(time.Time); ok && t.Hour() == 0 && t.Minute() == 0 && t.Second() == 0 {
				result.WriteString(fmt.Sprintf(stringFormat, t.Format(time.DateOnly)))
				htmlResult.WriteOpenTag("td", alignRight)
				htmlResult.WriteString(t.Format(time.DateOnly))
			} else {
				result.WriteString(fmt.Sprintf(variableFormat, value))
				htmlResult.WriteOpenTag("td", alignRight)
				htmlResult.WriteString(fmt.Sprintf("%v", value))
			}
		}
		result.WriteString("\n")
		htmlResult.WriteCloseTag("tr")
	}

	htmlResult.WriteCloseTag("table")

	if err := rows.Err(); err != nil {
		log.Fatalf("Error iterating over rows: %v", err)
	}

	println(result.String())
	// Create a new AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		log.Fatalf("Failed to create AWS session: %v", err)
	}

	// Create SNS service client
	svc := ses.New(sess)

	// Publish result to SNS
	message := result.String()
	subject := "6th Sense Daily Stats"
	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			ToAddresses: []*string{
				aws.String(email),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String(charSet),
					Data:    aws.String(htmlResult.String()),
				},
				Text: &ses.Content{
					Charset: aws.String(charSet),
					Data:    aws.String(message),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String(charSet),
				Data:    aws.String(subject),
			},
		},
		Source: aws.String(sender),
	}

	fmt.Println(htmlResult.String())

	// Attempt to send the email
	res, err := svc.SendEmail(input)
	if err != nil {
		log.Fatalf("Failed to send email: %v", err)
	}

	fmt.Println("Query result emailed successfully!")
	fmt.Println(res.GoString())

}
