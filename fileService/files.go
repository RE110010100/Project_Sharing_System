package main

import (
	"archive/zip"
	"bytes"
	"compress/flate"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	//"github.com/aws/aws-sdk-go/aws/credentials"

	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Global variables for MinIO access key and secret key
var (
	minIOaccessKey = "minio_user"
	minIOsecretKey = "minio_password"
	bucketName     = "projectfiles1"
)

type FilesService struct {
	minIO *minio.Client
	db    *sql.DB
	rdb   *redis.Client
}

// File struct represents the structure of the files table
type File struct {
	ID              string // varchar(36) corresponds to UUID or similar
	ProjectID       string // varchar(36) corresponds to UUID or similar
	FileName        string // varchar(255)
	FileSize        int    // int
	FileType        string // varchar(50)
	UploadTimestamp string // timestamp
}

// Publish a message with additional data
type Message struct {
	Text   string `json:"text"`
	UserID string `json:"text"`
}

// NewProjectService creates a new instance of ProjectService.
func NewFilesService(minIO *minio.Client, db *sql.DB, rdb *redis.Client) *FilesService {
	return &FilesService{minIO: minIO, db: db, rdb: rdb}

}

func uploadDirectoryToS3(db *sql.DB, minIO *minio.Client, localDirectoryPath, s3BaseKey string) error {
	// List files and subdirectories in the local directory
	files, err := os.ReadDir(localDirectoryPath)
	if err != nil {
		return err
	}

	// Iterate through files and subdirectories
	for _, file := range files {
		filePath := filepath.Join(localDirectoryPath, file.Name())
		relativePath, _ := filepath.Rel(localDirectoryPath, filePath)
		s3ObjectKey := filepath.Join(s3BaseKey, strings.ReplaceAll(relativePath, string(filepath.Separator), "/"))

		fmt.Println(s3ObjectKey)

		if file.IsDir() {
			// If it's a subdirectory, recursively upload its contents
			err := uploadDirectoryToS3(db, minIO, filePath, s3ObjectKey)
			if err != nil {
				return err
			}
		} else {
			// If it's a file, upload it to S3
			err := uploadFileToMinio_tmp(minIO, db, s3ObjectKey, filePath, "id")
			if err != nil {
				return err
			}

		}
	}

	return nil
}

func createDirectoriesIfNotExist(path string) error {
	// Get the absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	// Split the path into individual directories
	directories := strings.Split(absPath, `/`)

	fmt.Println(" dir [0] : " + directories[2])

	//fmt.Println(directories)
	dirPath := "/"

	// Iterate through the directories
	for index, dir := range directories {

		if index == 0 {
			continue
		}

		dirPath += dir + "/"
		fmt.Println("directory path : " + dirPath)

		_, err := os.Stat(dirPath)
		if os.IsNotExist(err) {
			// If the directory doesn't exist, create it
			fmt.Printf("Creating directory: %s\n", dir)
			err := os.Mkdir(dirPath, os.ModePerm)
			if err != nil {
				return err
			}
		} else if err != nil {
			// If there is an error other than "not exists," return it
			return err
		} else if err == nil {
			fmt.Println("Path exists for : " + dirPath)
		}
	}

	return nil
}

func downloadFolderFromS3(minIO *minio.Client, localPath string, folderKey string) error {

	// Create a downloader with the session and default options
	//downloader := s3manager.NewDownloader(sess)

	//Form the local path of the project directory
	localFolderPath := localPath + folderKey

	counter := 1

	for {
		// Check if the directory exists
		_, err := os.Stat(localFolderPath)
		if os.IsNotExist(err) {
			// Directory does not exist, break the loop
			break
		}

		// If the directory exists, try a new name
		localFolderPath = fmt.Sprintf("%s(%d)", localFolderPath, counter)
		folderKey = fmt.Sprintf("%s(%d)", folderKey, counter)
		counter++
	}

	// Create the directory (including parent directories) with 0755 permission
	err := os.MkdirAll(localFolderPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("Error creating directory:", err)

	}

	fmt.Println("successfully created a folder titled ", localFolderPath)

	//List the files in the directory
	fileKeys, err := ListFilesInMinioDirectory(minIO, bucketName, folderKey)
	if err != nil {
		return fmt.Errorf("Error listing files:", err)
	}

	fmt.Println("Files in the MinIO directory:")
	for _, fileKey := range fileKeys {
		fmt.Println(fileKey)
	}

	// Specify the parameters for the download
	for _, filekey := range fileKeys {

		//create the local file before populating it's contents
		localFilePath := localPath + filekey

		localDirPath := filepath.Dir(localFilePath)

		error1 := createDirectoriesIfNotExist(localDirPath)
		if error1 != nil {
			fmt.Printf("Error: %v\n", error1)
		} else {
			fmt.Println("Directories created or already exist.")
		}

		fmt.Print(localDirPath)

		// Create a file to write the downloaded content to
		file, err := os.Create(localFilePath)
		if err != nil {
			return fmt.Errorf("Error creating file: %v", err)
		}
		defer file.Close()

		err = downloadFileFromMinio_tmp(minIO, bucketName, filekey, localFilePath)
		if err != nil {
			return fmt.Errorf("error downloading folder: %v", err)

		}

	}

	fmt.Println("Folder downloaded successfully!")
	return nil
}

func (fs *FilesService) downloadFolderHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	localPath := r.URL.Query().Get("localPath")
	folderKey := r.URL.Query().Get("folderKey")

	if localPath == "" || folderKey == "" {
		http.Error(w, "localPath and folderKey are required parameters", http.StatusBadRequest)
		return
	}

	// Call the downloadFolderFromS3 function
	err := downloadFolderFromS3(fs.minIO, localPath, folderKey)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error downloading folder: %v", err), http.StatusInternalServerError)
		return
	}

	// Respond with a success message
	fmt.Fprint(w, "Folder downloaded successfully!")
}

func (fs *FilesService) uploadHandler(w http.ResponseWriter, r *http.Request) {

	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*") // Allow requests from any origin
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Respond to preflight request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

}

func (fs *FilesService) uploadDirectoryHandler(w http.ResponseWriter, r *http.Request) {

	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*") // Allow requests from any origin
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Respond to preflight request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Parse query parameters
	localPath := r.URL.Query().Get("localPath")
	s3BaseKey := r.URL.Query().Get("s3BaseKey")

	if localPath == "" || s3BaseKey == "" {
		http.Error(w, "localPath and s3BaseKey are required parameters", http.StatusBadRequest)
		return
	}

	// Call the uploadDirectoryToS3 function
	err := uploadDirectoryToS3(fs.db, fs.minIO, localPath, s3BaseKey)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error uploading directory: %v", err), http.StatusInternalServerError)
		return
	}

	//message := fmt.Sprintf("%s %s %s!", "Directory : ", s3BaseKey, "uploaded successfully")

	// Respond with a success message
	fmt.Fprint(w, "Directory uploaded successfully!")
}

// ListFilesInMinioDirectory lists files in a directory in a MinIO bucket
func ListFilesInMinioDirectory(minioClient *minio.Client, bucketName, directoryPath string) ([]string, error) {
	// Set context
	ctx := context.Background()

	// List objects in the specified directory
	objectCh := minioClient.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
		Prefix:    directoryPath,
		Recursive: true, // Set to true if you want to list objects recursively within the directory
	})

	var fileKeys []string

	// Iterate over the objects
	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("error listing objects: %w", object.Err)
		}
		// Append object key (file path) to the slice
		fileKeys = append(fileKeys, object.Key)
	}

	return fileKeys, nil
}

func DeleteFileFromMinIO(minioClient *minio.Client, db *sql.DB, bucketName, objectName string) error {
	// Set context to cancel operation after a certain timeout if needed
	ctx := context.Background()

	// Delete the file from the bucket
	err := minioClient.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return err
	}

	id, err := getFileIDByObjectKey(db, objectName)
	if err != nil {
		return err
	}

	err = deleteFileByID(db, id)
	if err != nil {
		return err
	}

	log.Println("File", objectName, "deleted successfully from bucket", bucketName)
	return nil
}

// Function to delete a record from the 'files' table by ID
func deleteFileByID(db *sql.DB, id string) error {
	// Prepare the SQL query
	query := "DELETE FROM files WHERE id = ?"

	// Execute the SQL query
	_, err := db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting file from 'files' table: %w", err)
	}

	fmt.Printf("File with ID %d deleted successfully.\n", id)
	return nil
}

func (fs *FilesService) DeleteFileHandler(w http.ResponseWriter, r *http.Request) {

	// Parse query parameters
	fileKey := r.URL.Query().Get("fileKey")

	if fileKey == "" {
		http.Error(w, "fileKey is a required parameters", http.StatusBadRequest)
		return
	}

	err := DeleteFileFromMinIO(fs.minIO, fs.db, bucketName, fileKey)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error deleting file or directory: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Println("Deleted succesfully")
}

// DeleteFilesInDirectory deletes all files in the specified directory of the bucket.
func DeleteFilesInDirectory(minioClient *minio.Client, db *sql.DB, directoryKey string) error {
	// Set context to cancel operation after a certain timeout if needed
	ctx := context.Background()

	// List all objects in the specified directory
	objectCh := minioClient.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
		Prefix:    directoryKey,
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return object.Err
		}
		// Skip directories
		if strings.HasSuffix(object.Key, "/") {
			continue
		}

		// Delete each file
		/*err := minioClient.RemoveObject(ctx, bucketName, object.Key, minio.RemoveObjectOptions{})
		if err != nil {
			return err
		}*/

		err := DeleteFileFromMinIO(minioClient, db, bucketName, object.Key)
		if err != nil {
			return object.Err
		}
		log.Println("File", object.Key, "deleted successfully from bucket", bucketName)
	}

	return nil
}

func (fs *FilesService) DeleteDirectoryHandler(w http.ResponseWriter, r *http.Request) {

	// Parse query parameters
	folderKey := r.URL.Query().Get("folderKey")

	if folderKey == "" {
		http.Error(w, "folderKey is a required parameters", http.StatusBadRequest)
		return
	}

	err := DeleteFilesInDirectory(fs.minIO, fs.db, folderKey)
	if err != nil {
		log.Fatal(err)
	}

}

// Function to create a new MinIO session
func newMinioSession_TMP() (*minio.Client, error) {
	// Initialize MinIO client object
	endpoint := "minio:9000" // MinIO endpoint

	// Initialize minio client object
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(minIOaccessKey, minIOsecretKey, ""),
		Secure: false, // Enable secure (HTTPS) connection
	})
	if err != nil {
		return nil, fmt.Errorf("error creating MinIO client: %w", err)
	}

	fmt.Printf("Bucket '%s' created successfully\n", bucketName)

	fmt.Println("Created MinIO client")

	return minioClient, nil
}

// Function to insert file information into the 'files' table
func insertFileInfo(db *sql.DB, fileName, fileType string, fileSize int64, projectID string) error {
	// Prepare the SQL query
	query := "INSERT INTO files (id, project_id, file_name, file_size, file_type) VALUES (?, ?, ?, ?, ?)"

	//
	id := uuid.New().String()

	// Execute the SQL query
	_, err := db.Exec(query, id, projectID, fileName, fileSize, fileType)
	if err != nil {
		return fmt.Errorf("error inserting file info into 'files' table: %w", err)
	}

	fmt.Println("File information inserted into 'files' table successfully.")
	return nil
}

// Function to update file information in the 'files' table based on ID
func updateFileInfoByID(db *sql.DB, id string, fileName, fileType string, fileSize int64) error {
	// Prepare the SQL query
	query := "UPDATE files SET file_name=?, file_type=?, file_size=? WHERE id=?"

	// Execute the SQL query
	_, err := db.Exec(query, fileName, fileType, fileSize, id)
	if err != nil {
		return fmt.Errorf("error updating file info in 'files' table: %w", err)
	}

	fmt.Printf("File information with ID %d updated successfully.\n", id)
	return nil
}

// Function to upload a file to a MinIO bucket
func uploadFileToMinio_tmp(minioClient *minio.Client, db *sql.DB, objectKey, filePath, projectID string) error {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	fmt.Println(objectKey)

	// Get file stats to determine file size
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("error getting file info: %w", err)
	}

	// Get file size
	fileSize := fileInfo.Size()

	// Get file name
	//fileName := filepath.Base(filePath)

	// Get file type (MIME type)
	fileType := mime.TypeByExtension(filepath.Ext(filePath))

	// Create a context
	ctx := context.Background()

	// Upload the file to the bucket
	_, err = minioClient.PutObject(ctx, bucketName, objectKey, file, fileInfo.Size(), minio.PutObjectOptions{})
	if err != nil {
		fmt.Print("error")
		return fmt.Errorf("error uploading file to bucket: %w", err)
	}

	err = insertFileInfo(db, objectKey, fileType, fileSize, projectID)
	if err != nil {
		fmt.Print("error")
		return fmt.Errorf("error uploading record to DB: %w", err)
	}

	fmt.Println("File uploaded successfully.")
	return nil
}

// handler to update a file
func (fs *FilesService) updateProjectFilesHandler(w http.ResponseWriter, r *http.Request) {

	// Parse query parameters
	fileKey := r.URL.Query().Get("fileKey")
	filePath := r.URL.Query().Get("filePath")

	if fileKey == "" || filePath == "" {
		http.Error(w, "Both fileKey or filePath are required parameters", http.StatusBadRequest)
		return
	}

	err := updateFileInMinio_tmp(fs.minIO, fs.db, fileKey, filePath)
	if err != nil {
		http.Error(w, "error updating files", http.StatusBadRequest)
		return
	}
}

func (fs *FilesService) updateFileHandler(w http.ResponseWriter, r *http.Request) {

	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*") // Allow requests from any origin
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Respond to preflight request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Parse multipart form
	/*err := r.ParseMultipartForm(10 << 20) // 10 MB max file size
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}*/

	// Get the file from the form
	file, Header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "No file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read the contents of the file
	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}

	fmt.Println(data)

	fileReader := bytes.NewReader(data)

	// Get file information
	/*fileInfo, err := file.Stat()
	if err != nil {
		http.Error(w, "Error getting file information", http.StatusInternalServerError)
		return
	}*/

	// Get the file size
	//fileSize := r.ContentLength

	// Get the user ID from the form data
	fileKey := r.FormValue("file_key")
	fileType := r.FormValue("file_type")
	fileid := r.FormValue("file_id")
	project := r.FormValue("project_title")
	UserID := r.FormValue("user_id")

	fmt.Println(fileid)

	// Upload file to MinIO
	_, err = fs.minIO.PutObject(context.Background(), bucketName, fileKey, fileReader, Header.Size, minio.PutObjectOptions{})
	if err != nil {
		http.Error(w, "Error uploading file to MinIO", http.StatusInternalServerError)
		fmt.Print(err)
		return
	}

	err = updateFileInfoByID(fs.db, fileid, fileKey, fileType, Header.Size)
	if err != nil {
		http.Error(w, "error getting file info:", http.StatusInternalServerError)
		return
	}

	message := fmt.Sprintf(`{"Text":"Updated file [%s] successfully in project [%s]", "UserID":"%s"}`, fileKey, project, UserID)

	fmt.Print(message)

	// Publish message to Redis after successful upload
	err = publishMessageToRedis(fs.rdb, message, "channel.upload")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error publishing message to Redis: %v", err), http.StatusInternalServerError)
		return
	}

	// Update file record in database (placeholder)
	// Replace this with your database update logic

	fmt.Fprintf(w, "File uploaded successfully")
}

// Function to update a file in a MinIO bucket
func updateFileInMinio_tmp(minioClient *minio.Client, db *sql.DB, objectKey, filePath string) error {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	// Get file stats to determine file size
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("error getting file info: %w", err)
	}

	// Get file size
	fileSize := fileInfo.Size()

	// Get file name
	//fileName := filepath.Base(filePath)

	// Get file type (MIME type)
	fileType := mime.TypeByExtension(filepath.Ext(filePath))

	id, err := getFileIDByObjectKey(db, objectKey)
	if err != nil {
		return fmt.Errorf("error getting file ID: %w", err)
	}

	err = updateFileInfoByID(db, id, objectKey, fileType, fileSize)
	if err != nil {
		return fmt.Errorf("error getting file info: %w", err)
	}

	// Create a context
	ctx := context.Background()

	// Upload the file to the bucket
	_, err = minioClient.PutObject(ctx, bucketName, objectKey, file, fileInfo.Size(), minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("error uploading file to bucket: %w", err)
	}

	fmt.Println("File uploaded successfully.")
	return nil
}

// Function to download a file from a MinIO bucket
func downloadFileFromMinio_tmp(minioClient *minio.Client, bucketName, objectName, filePath string) error {
	// Create a context
	ctx := context.Background()

	// Open the file for writing
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	// Download the object from the bucket
	err = minioClient.FGetObject(ctx, bucketName, objectName, file.Name(), minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("error downloading file from bucket: %w", err)
	}

	// Copy the object data to the file
	/*_, err = io.Copy(file, object)
	if err != nil {
		return fmt.Errorf("error copying object data to file: %w", err)
	}*/

	fmt.Println("File downloaded successfully.")
	return nil
}

// Function to get file ID by matching file name
func getFileIDByObjectKey(db *sql.DB, fileKey string) (string, error) {
	// Prepare the SQL query
	query := "SELECT id FROM files WHERE file_name = ?"

	// Execute the SQL query
	var id string
	err := db.QueryRow(query, fileKey).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			// Handle the case where no record is found
			return "", fmt.Errorf("file with name %s not found", fileKey)
		}
		return "", fmt.Errorf("error getting file ID: %w", err)
	}

	// Return the file ID
	return id, nil
}

func retrieveFileNameByProjectID(db *sql.DB, projectID string) (string, error) {

	// Prepare SQL query to retrieve filename by project ID
	query := "SELECT file_name FROM files WHERE project_id = ? LIMIT 1"
	var fileName string
	err := db.QueryRow(query, projectID).Scan(&fileName)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("no files found for project ID %s", projectID)
		}
		return "", fmt.Errorf("failed to retrieve filename: %v", err)
	}

	// Return the retrieved filename
	return fileName, nil
}

// GenerateDownloadURL generates a presigned URL for downloading a directory from MinIO
func GenerateDownloadURL(minioClient *minio.Client, bucketName, directoryPath string, expiry time.Duration) (string, error) {
	// Set context
	ctx := context.Background()

	// Generate a presigned URL for the directory
	presignedURL, err := minioClient.PresignedGetObject(ctx, bucketName, directoryPath, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("error generating presigned URL: %w", err)
	}

	return presignedURL.String(), nil
}

func getLastPathComponent(input string) string {
	// Split the input string by the "/" delimiter
	parts := strings.Split(input, "/")

	// If there's only one part, return it
	if len(parts) == 1 {
		return parts[0]
	}

	// Return the last part
	return parts[len(parts)-1]
}

func extractZipAndUploadProjectToMinIO(zipFile io.ReaderAt, size int64, bucketName string, minioClient *minio.Client, db *sql.DB, projectID string) error {
	// Open the zip file
	reader, err := zip.NewReader(zipFile, size)
	if err != nil {
		return err
	}

	// Iterate over each file in the zip archive
	for _, file := range reader.File {
		// Open the file within the zip archive
		fmt.Print(file.Name)
		inFile, err := file.Open()
		if err != nil {
			fmt.Print("error4")
			return err
		}
		defer inFile.Close()

		// Create a temporary file
		tmpFile, err := os.CreateTemp(".", "zip-extract-")
		if err != nil {
			fmt.Print("error3")
			return err
		}
		defer tmpFile.Close()

		// Write the contents of the file into the temporary file
		_, err = io.Copy(tmpFile, inFile)
		if err != nil {
			fmt.Print("error2")
			return err
		}

		// Upload the temporary file to MinIO storage
		objectName := file.Name

		fmt.Print(objectName)
		_, err = minioClient.FPutObject(context.Background(), bucketName, objectName, tmpFile.Name(), minio.PutObjectOptions{})
		if err != nil {
			fmt.Print("error1")
			return err
		}

		fileType := mime.TypeByExtension(filepath.Ext(file.Name))

		err = insertFileInfo(db, objectName, fileType, file.FileInfo().Size(), projectID)
		if err != nil {
			fmt.Print("error")
			return fmt.Errorf("error uploading record to DB: %w", err)
		}

		fmt.Printf("Uploaded file '%s' to MinIO storage\n", objectName)
	}

	return nil
}

func (fs *FilesService) uploadDirHandler(w http.ResponseWriter, r *http.Request) {

	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*") // Allow requests from any origin
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Respond to preflight request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Parse the multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	// Retrieve the file from the form data
	file, _, err := r.FormFile("zipFile")
	if err != nil {
		http.Error(w, "Unable to retrieve file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Get the user ID from the form data
	projectID := r.FormValue("project_id")
	UserID := r.FormValue("user_id")
	projectTitle := r.FormValue("project_title")

	// Read the entire content of the file into memory
	zipData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Unable to read file content", http.StatusInternalServerError)
		return
	}

	// Create a reader for the zip file content
	zipReaderAt := bytes.NewReader(zipData)

	// Extract and upload the files
	err = extractZipAndUploadProjectToMinIO(zipReaderAt, int64(len(zipData)), bucketName, fs.minIO, fs.db, projectID)
	if err != nil {
		errorMessage := "Error extracting and uploading zip file to MinIO: " + err.Error()
		http.Error(w, errorMessage, http.StatusInternalServerError)
		return
	}

	/*message := Message{
		Text:   "Hello, Redis!",
		UserID: UserID,
	}*/

	message := fmt.Sprintf(`{"Text":"Uploaded Project [%s] successfully", "UserID":"%s"}`, projectTitle, UserID)

	fmt.Print(message)

	// Publish message to Redis after successful upload
	err = publishMessageToRedis(fs.rdb, message, "channel.upload")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error publishing message to Redis: %v", err), http.StatusInternalServerError)
		return
	}

	// Respond with success message
	fmt.Fprintln(w, "File uploaded and extracted successfully")
}

func deleteFilesByProjectID(db *sql.DB, projectID string) error {

	// Prepare SQL query to delete files by project ID
	query := "DELETE FROM files WHERE project_id = ?"
	result, err := db.Exec(query, projectID)
	if err != nil {
		return fmt.Errorf("failed to execute delete query: %v", err)
	}

	// Check the number of rows affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	fmt.Printf("Deleted %d records with project ID %s\n", rowsAffected, projectID)

	return nil
}

func (fs *FilesService) updateDirHandler(w http.ResponseWriter, r *http.Request) {

	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*") // Allow requests from any origin
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Respond to preflight request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Parse the multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	// Retrieve the file from the form data
	file, _, err := r.FormFile("zipFile")
	if err != nil {
		http.Error(w, "Unable to retrieve file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Get the user ID from the form data
	projectID := r.FormValue("user_id")

	// Read the entire content of the file into memory
	zipData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Unable to read file content", http.StatusInternalServerError)
		return
	}

	// Create a reader for the zip file content
	zipReaderAt := bytes.NewReader(zipData)

	if err := deleteFilesByProjectID(fs.db, projectID); err != nil {
		fmt.Printf("Error deleting files: %v\n", err)
		return
	}

	// Extract and upload the files
	err = extractZipAndUploadProjectToMinIO(zipReaderAt, int64(len(zipData)), bucketName, fs.minIO, fs.db, projectID)
	if err != nil {
		http.Error(w, "Error extracting and uploading zip file to MinIO", http.StatusInternalServerError)
		return
	}
	// Respond with success message
	fmt.Fprintln(w, "File uploaded and extracted successfully")
}

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

func initializeRedisClient() *redis.Client {

	//var ctx = context.Background()

	// Initialize the Redis client
	client := redis.NewClient(&redis.Options{
		Addr:     "redis:6379", // Replace with your Redis server address
		Password: "",           // No password set
		DB:       0,            // Use default DB
	})

	// Ping the Redis server to check if it's reachable
	_, err := client.Ping().Result()
	if err != nil {
		panic(err)
	}

	fmt.Println("Connected to Redis server")

	return client
}

func publishMessageToRedis(redisClient *redis.Client, message, channel string) error {
	// Publish message to Redis channel
	err := redisClient.Publish(channel, message).Err()
	if err != nil {
		return err
	}
	return nil
}

// Function to get all file details by project ID
func getFilesByProjectID(db *sql.DB, projectID string) ([]File, error) {
	// Prepare the SQL query
	query := `
        SELECT id, project_id, file_name, file_size, file_type, upload_timestamp
        FROM files
        WHERE project_id = ?
    `

	// Execute the SQL query
	rows, err := db.Query(query, projectID)
	if err != nil {
		return nil, fmt.Errorf("error querying files: %w", err)
	}
	defer rows.Close()

	// Iterate over the rows and scan the file details into a slice
	var files []File
	for rows.Next() {
		var file File
		if err := rows.Scan(&file.ID, &file.ProjectID, &file.FileName, &file.FileSize, &file.FileType, &file.UploadTimestamp); err != nil {
			return nil, fmt.Errorf("error scanning file row: %w", err)
		}
		files = append(files, file)
	}

	// Check for any errors during iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating file rows: %w", err)
	}

	// Return the slice of file details
	return files, nil
}

// HTTP handler to get files by project ID
func (fs *FilesService) listFilesHandler(w http.ResponseWriter, r *http.Request) {

	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*") // Allow requests from any origin
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Respond to preflight request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Get the project ID from the URL query parameters
	projectID := r.URL.Query().Get("project_id")
	if projectID == "" {
		http.Error(w, "missing project_id parameter", http.StatusBadRequest)
		return
	}

	// Get files by project ID
	files, err := getFilesByProjectID(fs.db, projectID)
	if err != nil {
		http.Error(w, fmt.Sprintf("error getting files: %v", err), http.StatusInternalServerError)
		return
	}

	// Marshal files slice to JSON
	jsonResponse, err := json.Marshal(files)
	if err != nil {
		http.Error(w, fmt.Sprintf("error marshaling JSON response: %v", err), http.StatusInternalServerError)
		return
	}

	// Set content-type header and write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)

}

// RetrieveFilesAndCompress retrieves files from MinIO bucket, compresses them into a zip file, and returns as a blob
func RetrieveFilesAndCompress(minioClient *minio.Client, bucketName, directoryPath string) ([]byte, error) {
	// Get list of file keys
	fileKeys, err := ListFilesInMinioDirectory(minioClient, bucketName, directoryPath)
	if err != nil {
		return nil, err
	}

	// Create a temporary directory to store downloaded files
	tempDir, err := os.MkdirTemp("./", directoryPath)
	if err != nil {
		return nil, err
	}
	//defer os.RemoveAll(tempDir)

	// Download files from MinIO and write to temporary directory
	for _, fileKey := range fileKeys {
		objectInfo, err := minioClient.GetObject(context.Background(), bucketName, fileKey, minio.GetObjectOptions{})
		if err != nil {
			return nil, fmt.Errorf("error getting object %s: %w", fileKey, err)
		}
		defer objectInfo.Close()

		filePath := filepath.Join(tempDir, fileKey)
		fmt.Println(filePath)

		// Create directories as needed
		dirPath := filepath.Dir(filePath)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return nil, fmt.Errorf("error creating directory %s: %w", dirPath, err)
		}

		// Create the file in the temporary directory
		file, err := os.Create(filePath)
		if err != nil {
			return nil, err
		}
		defer file.Close() // Ensure that the file is closed after its use

		// Copy the contents of the object (file) to the created file
		_, err = io.Copy(file, objectInfo)
		if err != nil {
			return nil, fmt.Errorf("error copying object %s to file: %w", fileKey, err)
		}
	}

	// Create a buffer to store the compressed zip file
	var buf bytes.Buffer

	// Create a new zip archive
	zipWriter := NewZipWriter(&buf)

	// Walk through the temporary directory and add files to the zip archive
	err = filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Open the file
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		// Create a new zip file header
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// Set the file name as the relative path from the temporary directory
		header.Name, err = filepath.Rel(tempDir, path)
		if err != nil {
			return err
		}

		// Add the file header to the zip archive
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		// Copy the file contents to the zip archive
		_, err = io.Copy(writer, file)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Close the zip archive
	err = zipWriter.Close()
	if err != nil {
		return nil, err
	}

	// Return the compressed zip file as a blob
	return buf.Bytes(), nil
}

// NewZipWriter creates a new zip.Writer with a specific compression level
func NewZipWriter(w io.Writer) *zip.Writer {
	zipWriter := zip.NewWriter(w)
	zipWriter.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
		return flate.NewWriter(out, flate.BestCompression)
	})
	return zipWriter
}

func (fs *FilesService) zipDownloadHandler(w http.ResponseWriter, r *http.Request) {

	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*") // Allow requests from any origin
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Respond to preflight request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Retrieve folderKey from URL query parameter
	projectID := r.URL.Query().Get("projectID")

	fmt.Print(projectID)

	fileName, err := retrieveFileNameByProjectID(fs.db, projectID)
	if err != nil {
		fmt.Printf("Error retrieving filename: %v\n", err)
		return
	}

	fmt.Print(fileName)

	dirs := filepath.SplitList(fileName)
	if len(dirs) == 0 {
		http.Error(w, fmt.Sprintf("empty filepath %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Print(fileName)

	// Split the string by "/"
	parts := strings.Split(fileName, "/")

	// Retrieve folderKey from URL query parameter
	UserID := r.URL.Query().Get("UserID")

	fmt.Print(parts[1])

	// Retrieve files and compress them into a zip file
	zipBlob, err := RetrieveFilesAndCompress(fs.minIO, bucketName, parts[1])
	if err != nil {
		fmt.Print(err)
		http.Error(w, fmt.Sprintf("Error retrieving files and compressing: %v", err), http.StatusInternalServerError)
		return
	}

	zipName := parts[1] + ".zip"
	// Set the content type header
	w.Header().Set("Content-Type", "application/zip")
	// Set the content disposition header to make the browser download the file
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", zipName))

	// Write the zip file to the response writer
	_, err = w.Write(zipBlob)
	if err != nil {
		fmt.Print(err)
		http.Error(w, "Error writing zip file to response", http.StatusInternalServerError)
		return
	}

	message := fmt.Sprintf(`{"Text":"Downloaded Project zip file [%s] successfully", "UserID":"%s"}`, zipName, UserID)

	fmt.Print(message)

	// Publish message to Redis after successful upload
	err = publishMessageToRedis(fs.rdb, message, "channel.upload")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error publishing message to Redis: %v", err), http.StatusInternalServerError)
		return
	}

}

func (fs *FilesService) deleteProjectHandler(w http.ResponseWriter, r *http.Request) {

	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*") // Allow requests from any origin
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Respond to preflight request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Retrieve folderKey from URL query parameter
	projectID := r.URL.Query().Get("projectID")

	fileName, err := retrieveFileNameByProjectID(fs.db, projectID)
	if err != nil {
		fmt.Printf("Error retrieving filename: %v\n", err)
		return
	}

	dirs := filepath.SplitList(fileName)
	if len(dirs) == 0 {
		http.Error(w, fmt.Sprintf("empty filepath %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Print(fileName)

	// Split the string by "/"
	parts := strings.Split(fileName, "/")

	// Retrieve folderKey from URL query parameter
	UserID := r.URL.Query().Get("UserID")
	projectTitle := r.URL.Query().Get("project_title")

	err = DeleteFilesInDirectory(fs.minIO, fs.db, parts[1])
	if err != nil {
		log.Fatal(err)
	}

	message := fmt.Sprintf(`{"Text":"Deleted Project [%s] successfully", "UserID":"%s"}`, projectTitle, UserID)

	fmt.Print(message)

	// Publish message to Redis after successful upload
	err = publishMessageToRedis(fs.rdb, message, "channel.upload")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error publishing message to Redis: %v", err), http.StatusInternalServerError)
		return
	}

	// Respond with success message
	response := map[string]string{"message": "Project files deleted successfully"}
	json.NewEncoder(w).Encode(response)

}

func main() {

	minioClient, err := newMinioSession_TMP()
	if err != nil {
		log.Fatalln("Error creating MinIO session:", err)
	}

	db, err := connectToDB()
	if err != nil {
		log.Fatalln("Error connecting to DB:", err)
	}
	defer db.Close()

	// Initialize the Redis client
	redisClient := initializeRedisClient()
	defer redisClient.Close()

	FilesService := NewFilesService(minioClient, db, redisClient)

	http.HandleFunc("/uploadProjectFiles", FilesService.uploadDirectoryHandler)
	http.HandleFunc("/downloadProjectFiles", FilesService.downloadFolderHandler)
	http.HandleFunc("/deleteProjectFiles", FilesService.DeleteFileHandler)
	http.HandleFunc("/deleteDirectory", FilesService.DeleteDirectoryHandler)
	http.HandleFunc("/updateProjectFiles", FilesService.updateProjectFilesHandler)
	http.HandleFunc("/upload", FilesService.uploadDirHandler)
	http.HandleFunc("/listProjectFiles", FilesService.listFilesHandler)
	http.HandleFunc("/updateFile", FilesService.updateFileHandler)
	http.HandleFunc("/downloadZip", FilesService.zipDownloadHandler)
	http.HandleFunc("/updateProjectNewFiles", FilesService.updateDirHandler)
	http.HandleFunc("/deleteProjectHandler", FilesService.deleteProjectHandler)

	// Start the server on port 8081
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
