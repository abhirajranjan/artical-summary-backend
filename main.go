package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"sync"
	_ "unsafe"

	"github.com/gorilla/mux"
)

var (
	users         map[string]user
	userIDmapUser map[uint32]string
	history       map[uint32][]historyModel
	mutex         sync.RWMutex
	historyMu     sync.RWMutex
)

//go:linkname fastrand runtime.fastrand
func fastrand() uint32

func init() {
	users = map[string]user{}
	userIDmapUser = map[uint32]string{}
	history = map[uint32][]historyModel{}
	mutex = sync.RWMutex{}
	historyMu = sync.RWMutex{}
}

type opts map[string]http.HandlerFunc

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")
	
	next.ServeHTTP(w, r)
	})
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/v1/login", routeReq(opts{"POST": login()}))
	router.HandleFunc("/v1/register", routeReq(opts{"POST": register()}))
	router.HandleFunc("/v1/{userid}/history", routeReq(opts{"GET": listhistory(), "POST": addHistory()}))
	router.HandleFunc("/v1/{userid}/history/{historyid}", routeReq(opts{"DELETE": delHistory()}))

	router.HandleFunc("/v1/data", routeReq(opts{"GET": func(w http.ResponseWriter, r *http.Request) {
		mutex.RLock()
		fmt.Printf("%#v\n", users)
		fmt.Printf("%#v\n", userIDmapUser)
		mutex.RUnlock()
	}}))


	router.Use(enableCORS)
	fmt.Println("running server on port 80")
	if err := http.ListenAndServe(":80", router); err != http.ErrServerClosed {
		panic(err)
	}
}

func routeReq(o opts) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if fn, ok := o[r.Method]; ok {
			fn(w, r)
		}
	}
}

func login() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := loginModel{}
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			data, _ := json.Marshal(response{Error: "error parsing request"})
			w.WriteHeader(http.StatusBadRequest)
			w.Write(data)
			return
		}

		mutex.RLock()
		u, ok := users[user.Key()]
		mutex.RUnlock()

		if !ok {
			data, _ := json.Marshal(response{Error: "incorrect credentials"})
			w.WriteHeader(http.StatusBadRequest)
			w.Write(data)
			return
		}

		if !bytes.Equal(hash([]byte(user.Password)), u.HashedPassword) {
			data, _ := json.Marshal(response{Error: "incorrect credentials"})
			w.WriteHeader(http.StatusBadRequest)
			w.Write(data)
			return
		}

		data, _ := json.Marshal(response{
			Userid:   u.Id,
			Username: u.Username,
			Email:    u.Email,
		})

		w.Write(data)
	}
}

func register() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var userRq registerModel
		if err := json.NewDecoder(r.Body).Decode(&userRq); err != nil {
			data, _ := json.Marshal(response{Error: "error parsing request"})
			w.WriteHeader(http.StatusBadRequest)
			w.Write(data)
			return
		}

		mutex.RLock()
		_, ok := users[userRq.Key()]
		mutex.RUnlock()

		if ok {
			data, _ := json.Marshal(response{Error: "user already exists"})
			w.WriteHeader(http.StatusBadRequest)
			w.Write(data)
			return
		}

		user := createUser(userRq)

		mutex.Lock()
		userIDmapUser[user.Id] = userRq.Key()
		users[userRq.Key()] = user
		mutex.Unlock()

		historyMu.Lock()
		history[user.Id] = make([]historyModel, 0)
		historyMu.Unlock()

		data, _ := json.Marshal(response{
			Userid:   user.Id,
			Username: user.Username,
			Email:    user.Email,
		})
		w.Write(data)
	}
}

func listhistory() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userid := getUserID(r)
		if userid == 0 {
			data, _ := json.Marshal(response{Error: "incorrect userid. need int32"})
			w.WriteHeader(http.StatusBadRequest)
			w.Write(data)
			return
		}

		historyMu.RLock()
		h, ok := history[userid]
		historyMu.RUnlock()

		if !ok {
			data, _ := json.Marshal(response{Error: "user does not exists"})
			w.WriteHeader(http.StatusBadRequest)
			w.Write(data)
			return
		}

		var reverse = make([]historyModel, 0, len(h))
		for i := len(h) - 1; i >= 0; i-- {
			reverse = append(reverse, h[i])
		}

		data, _ := json.Marshal(struct {
			History []historyModel `json:"history"`
		}{History: reverse})
		w.Write(data)
	}
}

func delHistory() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userid := getUserID(r)
		if userid == 0 {
			data, _ := json.Marshal(response{Error: "incorrect userid. need int32"})
			w.WriteHeader(http.StatusBadRequest)
			w.Write(data)
			return
		}

		historyMu.RLock()
		_, ok := history[userid]
		historyMu.RUnlock()

		if !ok {
			data, _ := json.Marshal(response{Error: "user does not exists"})
			w.WriteHeader(http.StatusBadRequest)
			w.Write(data)
			return
		}

		historyid := getHistoryID(r)
		if historyid == 0 {
			data, _ := json.Marshal(response{Error: "incorrect historyid. need int32"})
			w.WriteHeader(http.StatusBadRequest)
			w.Write(data)
			return
		}

		historyMu.Lock()
		history[userid] = slices.DeleteFunc[[]historyModel, historyModel](history[userid], func(hm historyModel) bool {
			return hm.Id == historyid
		})
		historyMu.Unlock()
	}
}
func addHistory() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var model addHistoryModel

		if err := json.NewDecoder(r.Body).Decode(&model); err != nil {
			data, _ := json.Marshal(response{Error: "error parsing request"})
			w.WriteHeader(http.StatusBadRequest)
			w.Write(data)
			return
		}

		userid := getUserID(r)
		if userid == 0 {
			data, _ := json.Marshal(response{Error: "incorrect userid. need int32"})
			w.WriteHeader(http.StatusBadRequest)
			w.Write(data)
			return
		}

		historyMu.RLock()
		_, ok := history[userid]
		historyMu.RUnlock()

		if !ok {
			data, _ := json.Marshal(response{Error: "user does not exists"})
			w.WriteHeader(http.StatusBadRequest)
			w.Write(data)
			return
		}

		hmodel := historyModel{Id: fastrand(), Url: model.Url}

		historyMu.Lock()
		history[userid] = append(history[userid], hmodel)
		historyMu.Unlock()

		data, _ := json.Marshal(response{Historyid: hmodel.Id})
		w.Write(data)
	}
}
