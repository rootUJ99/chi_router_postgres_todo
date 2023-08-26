package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

type Todo struct {
	Name string `json:"name"`
	Id   int    `json:"id"`
}

type querries struct {
	create string
	update string
	get    string
	delete string
}
type Row struct {
	Name string `json:"name"`
	Id   string `json:"id"`
}

func returnError(w http.ResponseWriter, message string) {
	type ErrorMessage struct {
		ErrMessage string `json:"errMsg"`
	}
	w.WriteHeader(400)
	json.NewEncoder(w).Encode(ErrorMessage{ErrMessage: message})
}

func crudQuerry(key string, value string, key2 int) querries {
	q := querries{
		create: fmt.Sprintf("insert into todos (%v) values ('%v') returning %v;", key, value, key),
		update: fmt.Sprintf("update todos set %v = '%v' where id=%v returning %v;", key, value, key2, key),
		get:    fmt.Sprintf("select name, id from todos;"),
		delete: fmt.Sprintf("delete from todos where id=%v returning %v;", key2, key),
	}
	return q
}

func main() {
	godotenv.Load(".env")
	fmt.Println("hello from go project")

	database_url := os.Getenv("DATABASE_URL")
	conn, err := pgx.Connect(context.Background(), database_url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to connect to database %v\n", err)
		os.Exit(1)
	}

	defer conn.Close(context.Background())

	var dbname string
	err = conn.QueryRow(context.Background(), "select current_database()").Scan(&dbname)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("connected to the database %v\n", dbname)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(cors.Handler(
		cors.Options{
			AllowedOrigins:   []string{"http://*", "https://*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "OPTIONS", "DELETE"},
			AllowedHeaders:   []string{"Content-Type", "Accept", "Autorization"},
			ExposedHeaders:   []string{"link"},
			AllowCredentials: false,
			MaxAge:           300,
		},
	))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("helloz"))
	})

	r.Get("/get-todos", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		query := crudQuerry("", "", 0).get
		fmt.Println(query)
		row, err := conn.Query(context.Background(), query)
		if err != nil {
			returnError(w, "query error")
		}
		defer row.Close()
		var rowSlice []Row
		for row.Next() {
			var r Row
			if err := row.Scan(&r.Id, &r.Name); err != nil {
				returnError(w, "query error")

			}
			rowSlice = append(rowSlice, r)

		}
		fmt.Println(rowSlice)

		err = json.NewEncoder(w).Encode(&rowSlice)
		if err != nil {
			returnError(w, "server error")

		}

	})

	r.Post("/create-todo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		type todoBody struct {
			Name string `json:"name"`
		}
		var todobody todoBody
		if err := json.NewDecoder(r.Body).Decode(&todobody); err != nil {
			returnError(w, "parsing error")
		}
		query := crudQuerry("name", todobody.Name, 0).create
		fmt.Println(query)
		var name string
		err := conn.QueryRow(context.Background(), query).Scan(&name)
		if err != nil {
			returnError(w, "query error")
		}
		type resMesg struct {
			Message string `json:"message"`
		}

		err = json.NewEncoder(w).Encode(resMesg{Message: fmt.Sprintf("new %v has been created", name)})
		if err != nil {
			returnError(w, "server error")
		}

	})

	r.Put("/update-todo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var todo Todo
		if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
			returnError(w, "parsing error")
		}
		query := crudQuerry("name", todo.Name, todo.Id).update
		var name string
		if err := conn.QueryRow(context.Background(), query).Scan(&name); err != nil {
			returnError(w, "query error")
		}

		type resMesg struct {
			Message string `json:"message"`
		}

		err = json.NewEncoder(w).Encode(resMesg{Message: fmt.Sprintf("%v has been updated", name)})
		if err != nil {
			returnError(w, "server error")
		}

	})

	r.Delete("/delete-todo/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		id := chi.URLParam(r, "id")
		intId, err := strconv.Atoi(id)
		if err != nil {
			returnError(w, "wrong id format")
		}
		query := crudQuerry("name", "", intId).delete
		var name string
		if err := conn.QueryRow(context.Background(), query).Scan(&name); err != nil {
			returnError(w, "query error")
		}

		type resMesg struct {
			Message string `json:"message"`
		}

		err = json.NewEncoder(w).Encode(resMesg{Message: fmt.Sprintf("%v has been deleted", name)})
		if err != nil {
			returnError(w, "server error")
		}

	})

	http.ListenAndServe(":9090", r)

	fmt.Println("exiting server thankyou for time")
}
