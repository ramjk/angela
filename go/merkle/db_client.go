package merkle

import (
	_ "fmt"
)

func getLastThousandNodes() ([]*CoPathPair, error) {
	db, err := GetReadAngelaDB()
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
	db, err := GetReadAngelaDB()
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
	db, err := GetWriteAngelaDB()
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