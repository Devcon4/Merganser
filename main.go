package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"

	_ "github.com/jackc/pgx/stdlib"
)

type chat struct {
	ID      int    `json:"id" db:"chat_id"`
	Message string `json:"message" db:"message"`
	IsRead  bool   `db:"is_read"`
}

var dbContext *sqlx.DB

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/chat/{id}", getChatRequest).Methods("GET")
	router.HandleFunc("/chat", postChatRequest).Methods("POST")

	server := &http.Server{
		Addr:    ":3000",
		Handler: router,
	}

	setupDB()

	log.Fatal(server.ListenAndServe())
}

func connectDB() *sqlx.DB {
	if dbContext == nil {
		dbContext = sqlx.MustConnect("pgx", "postgres://dev:MerganserDev@localhost:5432/Merganser?sslmode=disable")
	}
	return dbContext
}

var insertChat = `INSERT INTO chat (chat_id, message, is_read) VALUES (:chat_id, :message, :is_read)`

func setupDB() {
	clearSchema := `DROP SCHEMA public CASCADE;
	CREATE SCHEMA public;
	GRANT ALL ON SCHEMA public TO dev;
	GRANT ALL ON SCHEMA public TO public;`

	schema := `CREATE TABLE chat (chat_id serial primary key, message varchar(255), is_read boolean);`

	chats := []chat{
		{ID: 1, Message: "Chat-1!", IsRead: false},
		{ID: 2, Message: "Chat-2!", IsRead: false},
		{ID: 3, Message: "Chat-3!", IsRead: true},
		{ID: 4, Message: "Chat-4!", IsRead: true},
	}

	db := connectDB()

	db.MustExec(clearSchema)
	db.MustExec(schema)

	_, err := db.NamedExec(insertChat, chats)

	if err != nil {
		panic(err)
	}
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
	id, _ := params["id"]
	res := chat{}

	err := dbContext.Get(&res, "SELECT * FROM chat WHERE chat_id=$1", id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeResponse(w, res)
}

func postChatRequest(w http.ResponseWriter, r *http.Request) {
	var req chat

	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusInternalServerError)
	}

	var id int

	rows, err := dbContext.NamedQuery(insertChat+" RETURNING chat_id", req)
	rows.Scan(&id)
	defer rows.Close()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	res := chat{}

	dbContext.Get(&res, `SELECT * FORM chat WHERE chat_id=$1`, id)

	writeResponse(w, res)
}
