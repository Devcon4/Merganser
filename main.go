package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type chat struct {
	ID      int    `json:"id" db:"id"`
	Message string `json:"message" db:"message"`
	IsRead  bool   `db:"is_read"`
}

var dbContext *sqlx.DB

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/chat/{id}", getChatRequest).Methods("GET")
	server := &http.Server{
		Addr:    ":3000",
		Handler: router,
	}

	setupDB()

	log.Fatal(server.ListenAndServe())
}

func setupDB() {
	schema := `CREATE TABLE chat (id integer, message text NULL, is_read boolean);`
	insertChat := `INSERT INTO chat (id, mesasge, is_read) VALUES (:id, :message, :is_read)`

	chats := []chat{
		{ID: 0, Message: "Chat-1!", IsRead: false},
		{ID: 1, Message: "Chat-2!", IsRead: false},
		{ID: 2, Message: "Chat-3!", IsRead: true},
		{ID: 3, Message: "Chat-4!", IsRead: true},
	}

	db := sqlx.MustConnect("sqlite3", ":memory:")

	db.MustExec(schema)

	tx := db.MustBegin()
	tx.NamedExec(insertChat, chats)
	tx.Commit()

	dbContext = db
}

func writeResponse(w http.ResponseWriter, o interface{}) {
	data, err := json.Marshal(o)

	if err != nil {
		http.Error(w, "Json parse error.", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func getChatRequest(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])
	var res *chat

	dbContext.Get(&res, "SELECT * FROM chat WHERE id=$1", id)

	if res == nil {
		http.Error(w, "No chat found.", http.StatusInternalServerError)
		return
	}

	writeResponse(w, res)
}
