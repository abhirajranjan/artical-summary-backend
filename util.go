package main

import (
	"crypto/sha256"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func hash(key []byte) []byte {
	sha := sha256.New()
	sha.Write(key)
	return sha.Sum(nil)
}

func getUserID(r *http.Request) uint32 {
	return extractVars(r, "userid")
}

func getHistoryID(r *http.Request) uint32 {
	return extractVars(r, "historyid")
}

func extractVars(r *http.Request, param string) uint32 {
	vars := mux.Vars(r)
	i, err := strconv.Atoi(vars[param])
	if err != nil {
		i = 0
	}

	return uint32(i)
}
