package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

// Project represents the project data structure.
type Project struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	IsPublic    bool   `json:"is_public"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// ProjectService handles project-related operations.
type ProjectService struct {
	db *sql.DB
}

// NewProjectService creates a new instance of ProjectService.
func NewProjectService(db *sql.DB) *ProjectService {
	return &ProjectService{db: db}
}

// CreateProject handles project creation with file storage.
func (ps *ProjectService) CreateProject(w http.ResponseWriter, r *http.Request) {
	var newProject Project
	err := json.NewDecoder(r.Body).Decode(&newProject)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Generate a unique ID for the project (you may want to use a library for this)
	newProject.ID = generateUniqueID()

	// Insert the project into the database, including file content
	_, err = ps.db.Exec("INSERT INTO projects (id, user_id, title, description, is_public) VALUES (?, ?, ?, ?, ?)",
		newProject.ID, newProject.UserID, newProject.Title, newProject.Description, newProject.IsPublic)
	if err != nil {
		http.Error(w, "Error creating project", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newProject)
}

// GetProjectByID retrieves a project by ID.
func (ps *ProjectService) GetProjectByID(w http.ResponseWriter, r *http.Request) {
	// Extract project ID from the request parameters
	projectID := r.URL.Query().Get("id")

	var project Project

	err := ps.db.QueryRow("SELECT id, user_id, title, description, is_public, created_at, updated_at FROM projects WHERE id = ?", projectID).
		Scan(&project.ID, &project.UserID, &project.Title, &project.Description, &project.IsPublic, &project.CreatedAt, &project.UpdatedAt)
	if err != nil {
		// Print the error details
		fmt.Println("Error fetching project:", err)
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	//json.NewEncoder(w).Encode(project)
	//fmt.Println("success")
}

// generateUniqueID generates a unique ID (you may want to use a library for this).
func generateUniqueID() string {
	// Implement your logic to generate a unique ID
	// For simplicity, you can use a random string or a timestamp
	return uuid.New().String()
}

func main() {

	// Connection parameters
	username := "root"
	password := "rohan123"
	host := "localhost"
	port := "3306"
	dbName := "File_Sharing_System"

	// Create a DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, host, port, dbName)

	// Open a connection to the database
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Ping the database to check if the connection is successful
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MySQL!")

	projectService := NewProjectService(db)

	http.HandleFunc("/createProject", projectService.CreateProject)
	http.HandleFunc("/getProject", projectService.GetProjectByID)

	// Start the server on port 8081
	err = http.ListenAndServe(":8081", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}

	// Now you can perform database operations using 'db'
}
