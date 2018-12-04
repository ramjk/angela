package merkle 

import (
	"fmt"
	"database/sql"
    // "github.com/aws/aws-sdk-go/aws"
    // "github.com/aws/aws-sdk-go/aws/session"
    // "github.com/aws/aws-sdk-go/service/sts"
    // "github.com/aws/aws-sdk-go/service/s3"
    "strings"
	// "github.com/aws/aws-sdk-go-v2/aws/external"
    // "github.com/aws/aws-sdk-go/aws/credentials"
	// "github.com/aws/aws-sdk-go/service/rds/rdsutils"
	_ "github.com/go-sql-driver/mysql"
	// "github.com/aws/aws-sdk-go-v2/aws/stscreds"
	// "os"
	"bytes"
)

const dropTableStmt = `DROP TABLE IF EXISTS nodes`
const createTableStmt = `
	CREATE TABLE IF NOT EXISTS nodes (
		nodeId VARBINARY(256) NOT NULL,
		nodeDigest VARBINARY(256) NOT NULL,
		epochNumber BIGINT NOT NULL,
		PRIMARY KEY (epochNumber, nodeId)
	)`
const showTablesStmt = "SHOW TABLES"
const getLastThousandNodesStmt = `
	SELECT nodeId, nodeDigest
	FROM nodes
	ORDER BY id DESC
	LIMIT 1000
	`
const createNodeIDIndex = `CREATE INDEX nodeID ON nodes (nodeId)`
const showMaxPacketSizeStmt = `SHOW VARIABLES LIKE 'version'`

type angelaDB struct {
	conn *sql.DB
}

func getAngelaWriteConnectionString() string {
	return writeConnectionString
}

func getAngelaReadConnectionString() string {
	return readConnectionString
}

func GetReadAngelaDB() (*angelaDB, error) {
	// Use db to perform SQL operations on database
	conn, err := sql.Open("mysql", getAngelaReadConnectionString())

	if err != nil {
		return nil, fmt.Errorf("[aurora]: could not open a connection: %v", err)
	}

	err = conn.Ping()

	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("[aurora]: unable to ping the database: %v", err)
	}

	angela := &angelaDB{conn: conn,}
	return angela, nil
}

func GetWriteAngelaDB() (*angelaDB, error) {
	// Use db to perform SQL operations on database
	conn, err := sql.Open("mysql", getAngelaWriteConnectionString())

	if err != nil {
		return nil, fmt.Errorf("[aurora]: could not open a connection: %v", err)
	}

	err = conn.Ping()

	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("[aurora]: unable to ping the database: %v", err)
	}

	angela := &angelaDB{conn: conn,}
	return angela, nil	
}

func (db *angelaDB) CreateTable() error {
	fmt.Println(createTableStmt)
	_, err := db.conn.Exec(createTableStmt)
	if err != nil {
		return err
	}
	_, err = db.conn.Exec(createNodeIDIndex)
	if err != nil {
		return err
	}
	return nil
}

func (db *angelaDB) DropTable() error {
	fmt.Println(dropTableStmt)
	_, err := db.conn.Exec(dropTableStmt)
	if err != nil {
		return err
	}
	return nil
}
func (db *angelaDB) ShowMaxPacketSize() error {
	rows, err := db.conn.Query(showMaxPacketSizeStmt)
	if err != nil {
		return fmt.Errorf("[aurora] could not query for max packet size: %v", err)
	}
	var res string
	var res2 string
    for rows.Next() {
        err := rows.Scan(&res, &res2)
        if err != nil {
            fmt.Println(err)
            return err
        }
        fmt.Println(res)
        fmt.Println(res2)
    }

	return nil
}

func (db *angelaDB) ShowTables() error {
	rows, err := db.conn.Query(showTablesStmt)
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

func (db *angelaDB) getChangeListInsertStmt(numNodes int) (*sql.Stmt, error) {
	var buffer bytes.Buffer

	buffer.WriteString("INSERT into nodes(nodeId, nodeDigest, epochNumber) VALUES ")
	buffer.WriteString(strings.Repeat("(?, ?, ?),", numNodes - 1))
    buffer.WriteString("(?, ?, ?)")
    buffer.WriteString(" ON DUPLICATE KEY UPDATE nodeDigest=VALUES(nodeDigest)")
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
	    vals = append(vals, elem.ID, elem.Digest, currentEpoch)
	}
	//format all vals at once
	transaction, err := db.conn.Begin()

	if err != nil {
		return -1, fmt.Errorf("[aurora] error in beginning transaction: %v", err)
	}
	res, err := stmt.Exec(vals...)
	transaction.Commit()	
	if err != nil {
		return -1, fmt.Errorf("[aurora] error in inserting hash nodes: %v", err)
	}
	numRowsAffected, err := res.RowsAffected()
	if err != nil {
		return -1, fmt.Errorf("[aurora]: could not get number of rows affected: %v", err)
	}
	return numRowsAffected, nil
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
		Digest:        nodeDigest,
	}
	return coPathPair, nil
}

func (db *angelaDB) getLastThousandNodes() ([]*CoPathPair, error) {
	getLastThousand, err := db.conn.Prepare(getLastThousandNodesStmt)
	if err != nil {
		return nil, fmt.Errorf("[aurora]: error in preparing get thousand statement: %v", err)		
	}
	rows, err := getLastThousand.Query()
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
	buffer.WriteString(`SELECT n1.nodeId, n1.nodeDigest 
			  			FROM nodes AS n1
			  			INNER JOIN 
			  			(SELECT nodeId, MAX(epochNumber)
			  			FROM nodes
			  			WHERE nodeId IN (`)
	buffer.WriteString(strings.Repeat("?,", numNodes - 1))
	buffer.WriteString("?) GROUP BY nodeId) AS latest ON n1.nodeId=latest.nodeId")

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
		fmt.Println(err)
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