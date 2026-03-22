package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// default types

type RegisterRequest struct {
	InitData string `json:"init_data"`
	FullName string `json:"full_name"`
	Group    string `json:"group"`
}

type AddSubjectRequest struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Node struct {
	ID             int      `json:"id"`
	Type           string   `json:"type,omitempty"`
	Name           string   `json:"name"`
	Value          string   `json:"value,omitempty"`
	Weight         float64  `json:"weight,omitempty"`
	DisplayFormula string   `json:"display_formula,omitempty"`
	Children[]Node   `json:"children,omitempty"`
}

type Subject struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	Grade     string `json:"grade,omitempty"`
	Structure[]Node `json:"structure,omitempty"`
}

// mock db

var (
	mu             sync.Mutex
	subjectCounter = 3
	subjects       =[]Subject{
		{
			ID:     1,
			Name:   "Алгоритмы и структуры данных",
			Status: "ready",
			Grade:  "8,4",
			Structure:[]Node{
				{
					ID:             100,
					Name:           "Итог",
					Value:          "8,4",
					DisplayFormula: "0.4 * ДЗ + 0.6 * Экзамен",
					Children:[]Node{
						{
							ID:     101,
							Type:   "folder",
							Name:   "Домашние задания",
							Value:  "9,0",
							Weight: 0.4,
							Children:[]Node{
								{ID: 102, Name: "ДЗ 1", Value: "10,0"},
								{ID: 103, Name: "ДЗ 2", Value: "8,0"},
							},
						},
						{
							ID:     104,
							Type:   "folder",
							Name:   "Экзамен",
							Value:  "8,0",
							Weight: 0.6,
							Children:[]Node{
								{ID: 105, Name: "Письменный", Value: "8,0"},
							},
						},
					},
				},
			},
		},
		{
			ID:     2,
			Name:   "Линейная алгебра",
			Status: "processing",
			Grade:  "0,0",
			Structure:[]Node{
				{
					ID:             200,
					Name:           "Итог",
					Value:          "0,0",
					DisplayFormula: "0.5 * КР + 0.5 * Экзамен",
					Children:[]Node{},
				},
			},
		},
	}
)

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:[]string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:[]string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))


	r.Post("/api/register", func(w http.ResponseWriter, r *http.Request) {
		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"token": "mock-jwt-token-xyz123"})
	})

	r.Post("/api/logs/navigation", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Route("/api/subjects", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			mu.Lock()
			defer mu.Unlock()
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(subjects)
		})
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			var req AddSubjectRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			mu.Lock()
			newSubject := Subject{
				ID:     subjectCounter,
				Name:   req.Name,
				Status: "processing",
				Grade:  "0.0",
				Structure:[]Node{
					{
						ID:             subjectCounter * 100,
						Name:           "Итог",
						Value:          "0,0",
						DisplayFormula: "Формула формируется...",
						Children:[]Node{},
					},
				},
			}
			subjectCounter++
			subjects = append(subjects, newSubject)
			mu.Unlock()

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(newSubject)
		})

		r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
			idStr := chi.URLParam(r, "id")
			id, _ := strconv.Atoi(idStr)

			mu.Lock()
			defer mu.Unlock()

			for _, s := range subjects {
				if s.ID == id {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(s)
					return
				}
			}
			http.Error(w, "Subject not found", http.StatusNotFound)
		})

		r.Delete("/{id}", func(w http.ResponseWriter, r *http.Request) {
			idStr := chi.URLParam(r, "id")
			id, _ := strconv.Atoi(idStr)

			mu.Lock()
			defer mu.Unlock()

			for i, s := range subjects {
				if s.ID == id {
					subjects = append(subjects[:i], subjects[i+1:]...)
					w.WriteHeader(http.StatusOK)
					return
				}
			}
			http.Error(w, "Subject not found", http.StatusNotFound)
		})
	})

	fmt.Println("Mock Backend is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
