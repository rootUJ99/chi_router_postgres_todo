package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load(".env")

	fmt.Println("hello from go project")
	database_url := os.Getenv("DATABASE_URL")
	fmt.Println(database_url, "url here")
	conn, err := pgx.Connect(context.Background(), database_url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to connect to database %v\n", err)
		os.Exit(1)
	}

	defer conn.Close(context.Background())

	var greeting string
	err = conn.QueryRow(context.Background(), "select current_database()").Scan(&greeting)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(greeting)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(cors.Handler(
		cors.Options{
			AllowedOrigins:   []string{"http://*", "https://*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "OPTIONS"},
			AllowedHeaders:   []string{"Content-Type", "Accept", "Autorization"},
			ExposedHeaders:   []string{"link"},
			AllowCredentials: false,
			MaxAge:           300,
		},
	))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("helloz"))
	})

	r.Post("/healthy-path", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("this is a healthy path"))
	})

	r.Get("/create-todo/{todo}", func(w http.ResponseWriter, r *http.Request) {
		todo := chi.URLParam(r, "todo")
		query := fmt.Sprintf("insert into todos (name) values ('%v') returning name;", todo)
		fmt.Println(query)
		var finalString string
		err := conn.QueryRow(context.Background(), query).Scan(&finalString)
		if err != nil {
			msg := fmt.Sprintf("QueryRow failed: %v\n", err)
			w.WriteHeader(400)
			w.Write([]byte(msg))
			return
		}
		fmt.Println(finalString)
		w.Write([]byte(finalString))

	})

	http.ListenAndServe(":9090", r)

	fmt.Println("exiting server thankyou for time")
}
