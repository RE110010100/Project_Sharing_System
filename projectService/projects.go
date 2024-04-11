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

	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*") // Allow requests from any origin
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Respond to preflight request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

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

func deleteProjectByID(db *sql.DB, projectID string) error {

	// Prepare SQL query to delete project by ID
	query := "DELETE FROM projects WHERE id = ?"
	result, err := db.Exec(query, projectID)
	if err != nil {
		return fmt.Errorf("failed to execute delete query: %v", err)
	}

	// Check the number of rows affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no project record found with ID %s", projectID)
	}

	fmt.Printf("Deleted project record with ID %s\n", projectID)

	return nil
}

// GetProjectByID retrieves a project by ID.
func (ps *ProjectService) deleteProjectByIDHandler(w http.ResponseWriter, r *http.Request) {

	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*") // Allow requests from any origin
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Respond to preflight request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Parse project ID from URL query parameters
	projectID := r.URL.Query().Get("projectID")
	if projectID == "" {
		http.Error(w, "Missing projectID parameter", http.StatusBadRequest)
		return
	}

	// Delete project record by ID
	err := deleteProjectByID(ps.db, projectID)
	if err != nil {
		fmt.Print(err)
		http.Error(w, fmt.Sprintf("Failed to delete project record: %v", err), http.StatusInternalServerError)
		return
	}

	// Respond with success message
	response := map[string]string{"message": "Project record deleted successfully"}
	json.NewEncoder(w).Encode(response)

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
	password := "root"
	host := "mysql"
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
	http.HandleFunc("/deleteProject", projectService.deleteProjectByIDHandler)

	// Start the server on port 8082
	err = http.ListenAndServe(":8082", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}

	// Now you can perform database operations using 'db'
}
