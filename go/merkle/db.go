package merkle 

import (
	"fmt"
	"database/sql"
    // "github.com/aws/aws-sdk-go/aws"
    // "github.com/aws/aws-sdk-go/aws/session"
    // "github.com/aws/aws-sdk-go/service/sts"
    // "github.com/aws/aws-sdk-go/service/s3"
    // "strings"
	// "github.com/aws/aws-sdk-go-v2/aws/external"
    // "github.com/aws/aws-sdk-go/aws/credentials"
	// "github.com/aws/aws-sdk-go/service/rds/rdsutils"
	_ "github.com/go-sql-driver/mysql"
	// "github.com/aws/aws-sdk-go-v2/aws/stscreds"
	// "os"
	"bytes"
)

const createTableStmt = `
	CREATE TABLE IF NOT EXISTS nodes (
		id INT UNSIGNED NOT NULL AUTO_INCREMENT,
		nodeId VARBINARY(256) NOT NULL,
		nodeDigest VARBINARY(256) NOT NULL,
		epochNumber BIGINT NOT NULL,
		PRIMARY KEY (id)
	)`
const showTablesStmt = "SHOW TABLES"
const insertNodeStmt = 
	`INSERT INTO nodes (nodeId, nodeDigest, epochNumber) VALUES (?, ?, ?)`
const getLatestNodeStmt = `
	SELECT nodeDigest
	FROM nodes 
	WHERE nodeId = ? 
	AND epochNumber < ? 
	ORDER BY epochNumber DESC
	LIMIT 1
	`
const getLastThousandNodesStmt = `
	SELECT nodeId, nodeDigest
	FROM nodes
	ORDER BY id DESC
	LIMIT 1000
	`

type angelaDB struct {
	conn *sql.DB
	insert *sql.Stmt
	getLatest *sql.Stmt
	getLastThousand *sql.Stmt
}

func getAngelaDBConnectionString() string {
	return connectionString
}

func getAngelaDB() (*angelaDB, error) {
	// Use db to perform SQL operations on database
	conn, err := sql.Open("mysql", getAngelaDBConnectionString())

	if err != nil {
		return nil, fmt.Errorf("[aurora]: could not open a connection: %v", err)
	}

	err = conn.Ping()

	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("[aurora]: unable to ping the database: %v", err)
	}

	angela := &angelaDB{conn: conn,}

	angela.insert, err = conn.Prepare(insertNodeStmt)
	if err != nil {
		return nil, fmt.Errorf("[aurora]: error in preparing insert statement: %v", err)
	}

	angela.getLatest, err = conn.Prepare(getLatestNodeStmt)
	if err != nil {
		return nil, fmt.Errorf("[aurora]: error in preparing get statement: %v", err)		
	}

	angela.getLastThousand, err = conn.Prepare(getLastThousandNodesStmt)
	if err != nil {
		return nil, fmt.Errorf("[aurora]: error in preparing get thousand statement: %v", err)		
	}

	return angela, nil
} 

func createTable(conn *sql.DB) error {
	fmt.Println(createTableStmt)
	_, err := conn.Exec(createTableStmt)
	if err != nil {
		return err
	}
	return nil
}

func showTables(conn *sql.DB) error {
	rows, err := conn.Query(showTablesStmt)
	if err != nil {
		return fmt.Errorf("[aurora] could not query for tables: %v", err)
	}
	var res string
    for rows.Next() {
        err := rows.Scan(&res)
        if err != nil {
            fmt.Println(err)
            return err
        }
        fmt.Println(res)
    }

	return nil
}

func (db *angelaDB) getLatestNodeDigest(nodeId string, currentEpoch int64) (string, error) {
	node := db.getLatest.QueryRow(nodeId, currentEpoch)	
	var nodeDigest string
	err := node.Scan(&nodeDigest)

	if err != nil {
		return "", fmt.Errorf("[aurora] error when scanning row results: %v", err)
	}
	return nodeDigest, nil
}

func (db *angelaDB) insertNode(nodeId string, nodeDigest string, epochNumber int64) (int64, error) {
	res, err := db.insert.Exec(nodeId, nodeDigest, epochNumber)
	if err != nil {
		return 0, fmt.Errorf("[aurora] error in inserting hash node: %v", err)
	}

	lastInsertID, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("[aurora]: could not get last insert ID: %v", err)
	}
	return lastInsertID, nil
}

func (db *angelaDB) getChangeListInsertStmt(numNodes int) (*sql.Stmt, error) {
	var buffer bytes.Buffer

	buffer.WriteString("INSERT into nodes(nodeId, nodeDigest, epochNumber) VALUES ")
    for i := 0; i < numNodes - 1; i++ {
        buffer.WriteString("(?, ?, ?),")
    }
    buffer.WriteString("(?, ?, ?)")

    stmt := buffer.String()

    changeListStmt, err := db.conn.Prepare(stmt)
	if err != nil {
		return nil, fmt.Errorf("[aurora]: error in preparing write statement: %v", err)		
	}
	return changeListStmt, nil
}

func (db *angelaDB) insertChangeList(changeList []*CoPathPair, currentEpoch int64) (int64, error) {
	stmt, err := db.getChangeListInsertStmt(len(changeList))

	if err != nil {
		return -1, fmt.Errorf("[aurora]: error in making changeList statement: %v", err)
	}

	vals := []interface{}{}
	for _, elem := range changeList {
	    vals = append(vals, elem.ID, elem.digest, currentEpoch)
	}
	//format all vals at once
	res, err := stmt.Exec(vals...)	
	if err != nil {
		return -1, fmt.Errorf("[aurora] error in inserting hash nodes: %v", err)
	}

	lastInsertID, err := res.LastInsertId()
	if err != nil {
		return -1, fmt.Errorf("[aurora]: could not get last insert ID: %v", err)
	}
	return lastInsertID, nil
}

type rowScanner interface {
	Scan(dest ...interface{}) error
}

// scan in the information for a copath node
func readRows(scanner rowScanner) (*CoPathPair, error) {
	var nodeId string
	var nodeDigest []byte
	err := scanner.Scan(&nodeId, &nodeDigest)
	if err != nil {
		return nil, fmt.Errorf("[aurora]: error in scanning copath nodes: %v", err)
	}

	coPathPair := &CoPathPair{
		ID:            nodeId,
		digest:        nodeDigest,
	}
	return coPathPair, nil
}

func (db *angelaDB) getLastThousandNodes() ([]*CoPathPair, error) {
	rows, err := db.getLastThousand.Query()
	if err != nil {
		return make([]*CoPathPair, 0), fmt.Errorf("[aurora]: error in querying copath statement: %v", err)
	}
	defer rows.Close()

	var results []*CoPathPair
	for rows.Next() {
		copathPair, err := readRows(rows)
		if err != nil {
			return make([]*CoPathPair, 0), fmt.Errorf("[aurora]: could not read row: %v", err)
		}
		results = append(results, copathPair)
	}

	return results, nil
}

func (db *angelaDB) getCopathQueryStmt(numNodes int) (*sql.Stmt, error) {

	var buffer bytes.Buffer

	// TODO: Get only the latest node id, node digest pair 
	buffer.WriteString("SELECT nodeId, nodeDigest FROM nodes WHERE nodeId IN (")
    for i := 0; i < numNodes - 1; i++ {
        buffer.WriteString("?,")
    }
    buffer.WriteString("?)")

    stmt := buffer.String()

    copathStmt, err := db.conn.Prepare(stmt)
	if err != nil {
		return nil, fmt.Errorf("[aurora]: error in preparing get statement: %v", err)		
	}
	return copathStmt, nil
}

func (db *angelaDB) retrieveLatestCopathDigests(copaths []string) ([]*CoPathPair, error) {
	stmt, err := db.getCopathQueryStmt(len(copaths))
	if err != nil {
		return make([]*CoPathPair, 0), fmt.Errorf("[aurora]: error in making copath statement: %v", err)
	}
	vals := []interface{}{}
	for _, id := range copaths {
	    vals = append(vals, id)
	}
	rows, err := stmt.Query(vals...)
	if err != nil {
		return make([]*CoPathPair, 0), fmt.Errorf("[aurora]: error in querying copath statement: %v", err)
	}
	defer rows.Close()

	var results []*CoPathPair
	for rows.Next() {
		copathPair, err := readRows(rows)
		if err != nil {
			return make([]*CoPathPair, 0), fmt.Errorf("[aurora]: could not read row: %v", err)
		}
		results = append(results, copathPair)
	}

	return results, nil
}

func (db *angelaDB) Close() {
	db.conn.Close()
}