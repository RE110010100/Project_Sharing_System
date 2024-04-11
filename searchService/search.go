package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

// Project represents a project record from the database
type Project struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	IsPublic    bool   `json:"is_public"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type SearchService struct {
	db *sql.DB
}

// NewProjectService creates a new instance of ProjectService.
func NewSearchService(db *sql.DB) *SearchService {
	return &SearchService{db: db}
}

// Function to connect to the mysql DB
func connectToDB() (*sql.DB, error) {

	// Connection parameters
	username := "root"
	password := "root"
	host := "mysql"
	port := "3306"
	dbName := "File_Sharing_System"

	// Create a DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, host, port, dbName)

	// Open a connection to the database
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("error opening DB: %w", err)
	}

	// Ping the database to check if the connection is successful
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("error pinging to DB: %w", err)
	}

	fmt.Println("Connected to MySQL!")
	return db, nil

}

func (ss *SearchService) searchAllProjects(w http.ResponseWriter, r *http.Request) {
	// Parse query parameter "q" for the search string
	searchString := r.URL.Query().Get("q")

	// Prepare the query statement
	query := "SELECT id, user_id, title, description, is_public, created_at, updated_at FROM projects WHERE title LIKE ? OR description LIKE ? LIMIT 10"
	rows, err := ss.db.Query(query, "%"+searchString+"%", "%"+searchString+"%")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Create a slice to hold the results
	var projects []Project

	// Iterate through the result set
	for rows.Next() {
		var project Project
		err := rows.Scan(&project.ID, &project.UserID, &project.Title, &project.Description, &project.IsPublic, &project.CreatedAt, &project.UpdatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		projects = append(projects, project)
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert the projects slice to JSON
	jsonResponse, err := json.Marshal(projects)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("Top 10 Projects:")
	for i, project := range projects {
		fmt.Printf("%d. ID: %s, UserID: %s, Title: %s, Description: %s, Public: %t, CreatedAt: %s, UpdatedAt: %s\n",
			i+1, project.ID, project.UserID, project.Title, project.Description, project.IsPublic, project.CreatedAt, project.UpdatedAt)
	}

	// Set the Content-Type header and write the JSON response
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func (ss *SearchService) searchUserAllProjects(w http.ResponseWriter, r *http.Request) {
	// Parse query parameter "userId" for the user ID
	userID := r.URL.Query().Get("userId")

	// Prepare the query statement
	query := "SELECT id, user_id, title, description, is_public, created_at, updated_at FROM projects WHERE user_id = ? LIMIT 10"
	rows, err := ss.db.Query(query, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Create a slice to hold the results
	var projects []Project

	// Iterate through the result set
	for rows.Next() {
		var project Project
		err := rows.Scan(&project.ID, &project.UserID, &project.Title, &project.Description, &project.IsPublic, &project.CreatedAt, &project.UpdatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		projects = append(projects, project)
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert the projects slice to JSON
	jsonResponse, err := json.Marshal(projects)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("Top 10 Projects:")
	for i, project := range projects {
		fmt.Printf("%d. ID: %s, UserID: %s, Title: %s, Description: %s, Public: %t, CreatedAt: %s, UpdatedAt: %s\n",
			i+1, project.ID, project.UserID, project.Title, project.Description, project.IsPublic, project.CreatedAt, project.UpdatedAt)
	}

	// Set the Content-Type header and write the JSON response
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func (ss *SearchService) listUserAllProjects(w http.ResponseWriter, r *http.Request) {

	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*") // Allow requests from any origin
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Respond to preflight request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Parse query parameter "userId" for the user ID
	userID := r.URL.Query().Get("userId")

	// Prepare the query statement
	query := "SELECT id, user_id, title, description, is_public, created_at, updated_at FROM projects WHERE user_id = ?"
	rows, err := ss.db.Query(query, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Create a slice to hold the results
	var projects []Project

	// Iterate through the result set
	for rows.Next() {
		var project Project
		err := rows.Scan(&project.ID, &project.UserID, &project.Title, &project.Description, &project.IsPublic, &project.CreatedAt, &project.UpdatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		projects = append(projects, project)
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert the projects slice to JSON
	jsonResponse, err := json.Marshal(projects)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("Top 10 Projects:")
	for i, project := range projects {
		fmt.Printf("%d. ID: %s, UserID: %s, Title: %s, Description: %s, Public: %t, CreatedAt: %s, UpdatedAt: %s\n",
			i+1, project.ID, project.UserID, project.Title, project.Description, project.IsPublic, project.CreatedAt, project.UpdatedAt)
	}

	// Set the Content-Type header and write the JSON response
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func (ss *SearchService) searchProjects(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters for userID and search string
	userID := r.URL.Query().Get("userId")
	searchString := r.URL.Query().Get("q")

	// Prepare the query statement
	query := "SELECT id, user_id, title, description, is_public, created_at, updated_at FROM projects WHERE user_id = ? AND (title LIKE ? OR description LIKE ?) LIMIT 10"
	rows, err := ss.db.Query(query, userID, "%"+searchString+"%", "%"+searchString+"%")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Create a slice to hold the results
	var projects []Project

	// Iterate through the result set
	for rows.Next() {
		var project Project
		err := rows.Scan(&project.ID, &project.UserID, &project.Title, &project.Description, &project.IsPublic, &project.CreatedAt, &project.UpdatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		projects = append(projects, project)
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert the projects slice to JSON
	jsonResponse, err := json.Marshal(projects)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set the Content-Type header and write the JSON response
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func main() {

	db, err := connectToDB()
	if err != nil {
		log.Fatalln("Error connecting to DB:", err)
	}
	defer db.Close()

	searchService := NewSearchService(db)

	http.HandleFunc("/searchAllProjectWithString", searchService.searchAllProjects)
	http.HandleFunc("/searchUserAllProjects", searchService.searchUserAllProjects)
	http.HandleFunc("/searchUserProjectWithString", searchService.searchProjects)
	http.HandleFunc("/listUserProjectWithString", searchService.listUserAllProjects)

	// Start the server on port 8081
	err = http.ListenAndServe(":8083", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}

}
