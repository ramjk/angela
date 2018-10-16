package main

import "fmt"

import (
	"database/sql"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    // "github.com/aws/aws-sdk-go/service/sts"
    // "github.com/aws/aws-sdk-go/service/s3"
    // "strings"
	// "github.com/aws/aws-sdk-go-v2/aws/external"
    "github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/rds/rdsutils"
	// "github.com/go-sql-driver/mysql"
	// "github.com/aws/aws-sdk-go-v2/aws/stscreds"
	"os"
)

func main() {
	HOME := os.Getenv("HOME")

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2c"),
		Credentials: credentials.NewSharedCredentials(HOME + "/.aws/credentials", ""),
		})

	credentialsObj := sess.Config.Credentials

	//////

	awsRegion := "us-west-2c"
	dbUser := "kubidoo"
	dbName := "angela"
	dbEndpoint := "http://angela.cluster-c7vkkm31zszq.us-west-2.rds.amazonaws.com:3306"

	authToken, err := rdsutils.BuildAuthToken(dbEndpoint, awsRegion, dbUser, credentialsObj)

	if err != nil {
		fmt.Println("Could not build auth token.")
	}

	// Create the MySQL DNS string for the DB connection
	// user:password@protocol(endpoint)/dbname?<params>
	dnsStr := fmt.Sprintf("%s:%s@tcp(%s)/%s?tls=true",
	   dbUser, authToken, dbEndpoint, dbName,
	)

	// Use db to perform SQL operations on database
	db, err := sql.Open("mysql", dnsStr)

	if err != nil {
		fmt.Println("Cannot open database.")
	}

	fmt.Println(db)

	// cfg, err := external.LoadDefaultAWSConfig()
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "failed to load configuration, %v", err)
	// 	os.Exit(1)
	// }
	// cfg.Region = awsRegion

	// credProvider := stscreds.NewAssumeRoleProvider(sts.New(cfg), os.Args[5])

	// awsCreds := aws.NewCredentials(credProvider)
	// authToken, err := rdsutils.BuildAuthToken(dbEndpoint, awsRegion, dbUser, awsCreds)

	// // Create the MySQL DNS string for the DB connection
	// // user:password@protocol(endpoint)/dbname?<params>
	// dnsStr := fmt.Sprintf("%s:%s@tcp(%s)/%s?tls=true",
	// 	dbUser, authToken, dbEndpoint, dbName,
	// )

	// driver := mysql.MySQLDriver{}
	// _ = driver
	// // Use db to perform SQL operations on database
	// if db, err = sql.Open("mysql", dnsStr); err != nil {
	// 	panic(err)
	// }

	// fmt.Println("Successfully opened connection to database")

}
