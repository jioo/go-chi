package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	_ "github.com/joho/godotenv/autoload"

	"github.com/jioo/go-chi/db"
)

// album represents data about a record album.
type album struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}

func main() {
	// setup router
	r := chi.NewRouter()

	// setup middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// setup db connection
	r.Use(dbCtx)

	// setup routes
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		db, ok := ctx.Value("db").(*sql.DB)
		if !ok {
			http.Error(w, http.StatusText(422), 422)
			return
		}

		albums, err := getAlbums(db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(albums)
	})

	// start server
	http.ListenAndServe(":8080", r)
}

func dbCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db, err := db.Connect()
		if err != nil {
			log.Fatal(err)
		}

		ctx := context.WithValue(r.Context(), "db", db)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getAlbums(db *sql.DB) ([]album, error) {
	var albums []album

	rows, err := db.Query("SELECT * FROM album")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var alb album
		if err := rows.Scan(&alb.ID, &alb.Title, &alb.Artist, &alb.Price); err != nil {
			return nil, err
		}
		albums = append(albums, alb)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return albums, nil
}
