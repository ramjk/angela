package merkle

import (
	"fmt"
)

//export BatchUpdate
func BatchUpdate() {
	db, err := getAngelaDB()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	id, err := db.insertNode("001", "RAM", 1)
	if err != nil {
		panic(err)
	}
	fmt.Println(id)

	nodeDigest, err := db.getLatestNodeDigest("001", 2)
	if err != nil {
		panic(err)
	}
	fmt.Println(nodeDigest)
}

func getLastThousandNodes() ([]*CoPathPair, error) {
	db, err := getAngelaDB()
	if err != nil {
		return make([]*CoPathPair, 0), err
	}
	defer db.Close()

	copathPairs, err := db.getLastThousandNodes()
	if err != nil {
		return copathPairs, err
	}

	return copathPairs, nil	
} 

func retrieveCopaths(copaths []string) ([]*CoPathPair, error) {
	db, err := getAngelaDB()
	if err != nil {
		return make([]*CoPathPair, 0), err
	}
	defer db.Close()

	copathPairs, err := db.retrieveLatestCopathDigests(copaths)
	if err != nil {
		return copathPairs, err
	}

	return copathPairs, nil
}

func writeChangeList(changeList []*CoPathPair, currentEpoch int64) (int64, error) {
	db, err := getAngelaDB()
	if err != nil {
		return -1, err
	}
	defer db.Close()

	lastNodeId, err := db.insertChangeList(changeList, currentEpoch)
	if err != nil {
		return -1, err
	}
	return lastNodeId, nil
}

func main() {
	BatchUpdate()
}