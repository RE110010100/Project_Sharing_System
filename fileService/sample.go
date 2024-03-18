package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	//"github.com/aws/aws-sdk-go/aws/credentials"

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
}

// NewProjectService creates a new instance of ProjectService.
func NewFilesService(minIO *minio.Client, db *sql.DB) *FilesService {
	return &FilesService{minIO: minIO, db: db}

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
			err := uploadFileToMinio_tmp(minIO, db, s3ObjectKey, filePath)
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

func (fs *FilesService) uploadDirectoryHandler(w http.ResponseWriter, r *http.Request) {
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
	endpoint := "localhost:9000" // MinIO endpoint

	// Initialize minio client object
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(minIOaccessKey, minIOsecretKey, ""),
		Secure: false, // Enable secure (HTTPS) connection
	})
	if err != nil {
		return nil, fmt.Errorf("error creating MinIO client: %w", err)
	}

	fmt.Println("Created MinIO client")

	return minioClient, nil
}

// Function to insert file information into the 'files' table
func insertFileInfo(db *sql.DB, fileName, fileType string, fileSize int64) error {
	// Prepare the SQL query
	query := "INSERT INTO files (id, project_id, file_name, file_size, file_type) VALUES (?, ?, ?, ?, ?)"

	//
	id := uuid.New().String()

	// Execute the SQL query
	_, err := db.Exec(query, id, "3fd0f0d4-ee52-4ce4-9e6a-7fe3e7db2e5c", fileName, fileSize, fileType)
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
func uploadFileToMinio_tmp(minioClient *minio.Client, db *sql.DB, objectKey, filePath string) error {
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
		return fmt.Errorf("error uploading file to bucket: %w", err)
	}

	err = insertFileInfo(db, objectKey, fileType, fileSize)
	if err != nil {
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

func connectToDB() (*sql.DB, error) {

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

	FilesService := NewFilesService(minioClient, db)

	http.HandleFunc("/uploadProjectFiles", FilesService.uploadDirectoryHandler)
	http.HandleFunc("/downloadProjectFiles", FilesService.downloadFolderHandler)
	http.HandleFunc("/deleteProjectFiles", FilesService.DeleteFileHandler)
	http.HandleFunc("/deleteDirectory", FilesService.DeleteDirectoryHandler)
	http.HandleFunc("/updateProjectFiles", FilesService.updateProjectFilesHandler)

	// Start the server on port 8081
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
