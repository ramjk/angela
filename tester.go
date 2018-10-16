package main

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
	_ "github.com/go-sql-driver/mysql"
	// "github.com/aws/aws-sdk-go-v2/aws/stscreds"
	"os"
)

func main() {
	HOME := os.Getenv("HOME")

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2c"),
		Credentials: credentials.NewSharedCredentials(HOME + "/.aws/credentials", ""),
		})

	creds := sess.Config.Credentials

	//////

	region := "us-west-2c"
	dbUser := "kubidoo"
	dbName := "angela"
	endpoint := "angela.cluster-c7vkkm31zszq.us-west-2.rds.amazonaws.com:3306"

	connectionString := rdsutils.NewConnectionStringBuilder(endpoint, region, dbUser, dbName, creds)

	dnsStr, err := connectionString.WithTCPFormat().Build()

	if err != nil {
		panic(err)
	}


	// Use db to perform SQL operations on database
	db, err := sql.Open("mysql", dnsStr)

	defer db.Close()

	if err != nil {
		panic(err)
	}

	err = db.Ping()

	if err != nil {
		panic(err)
	}

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
