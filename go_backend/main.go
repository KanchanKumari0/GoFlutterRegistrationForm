//Created By Kanchan 
//First Version: 18 APRIL
//Second Version : 19 APRIL : Added Email Functionality 
//Third Version : 19 APRIL : Fixed Error Handling of Database Connection

package main

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"log"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"database/sql"
	"gopkg.in/gomail.v2"
    "fmt"
    _ "github.com/denisenkom/go-mssqldb"
)

//
type User struct {
	FirstName string `json:"first_name"`
    LastName  string `json:"last_name"`
    DOB       string `json:"dob"`
    Email     string `json:"email"`
    PhoneNumber string `json:"phone_number"`
    CV        string `json:"cv"`
    FileName  string `json:"filename"`
}

func main() {
	fmt.Println("Starting Server")
	router := mux.NewRouter()
	router.HandleFunc("/register", registerHandler).Methods("POST")

	// Apply CORS middleware
	corsHandler := cors.Default().Handler(router)

	log.Fatal(http.ListenAndServe(":8080", corsHandler))
}

func sendEmail(to, subject, body string) error {
    // Set up the email message
    m := gomail.NewMessage()
    m.SetHeader("From", "learning.williamcareyuniversity@gmail.com")
    m.SetHeader("To", to)
    m.SetHeader("Subject", subject)
    m.SetBody("text/html", body)
    fmt.Println(to)
    fmt.Println(subject)
    fmt.Println(body)
    // Set up the email server configuration
    d := gomail.NewDialer("smtp.gmail.com", 587, "learning.williamcareyuniversity@gmail.com", "uhgdguieswyfgeia")
    fmt.Println("Sending Email")
    // Send the email
    if err := d.DialAndSend(m); err != nil {
    	fmt.Println(err)
        return err
    }
    fmt.Println("Email Sent")
    return nil
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Request Came")
	decoder := json.NewDecoder(r.Body)
	var user User
	err := decoder.Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	decoded, err := base64.StdEncoding.DecodeString(user.CV)
if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
}

// Set the upload directory and filename
uploadDir := "Resumes"
uploadPath := filepath.Join(uploadDir, user.FileName)

// Create the upload directory if it doesn't exist
err = os.MkdirAll(uploadDir, 0755)
if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
}

// Create a new file with the given path
file, err := os.Create(uploadPath)
if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
}
defer file.Close()

// Write the decoded data to the file
_, err = file.Write(decoded)
if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
}
	fmt.Println("File Created")
	filePath := file.Name()
    fmt.Println("Path" + filePath)
    server := "localhost"
    port := 1433
    database := "goConsole"
    connectionString := fmt.Sprintf("server=%s;port=%d;database=%s",
        server, port, database)
    
    fmt.Println("Establishing Connection to Database")
    db, err := sql.Open("mssql", connectionString)
    pingErr := db.Ping()
    if pingErr != nil {
     log.Fatal(pingErr)
     return
    }
    fmt.Println("Connected!")
    defer db.Close()

    // Prepare the SQL statement
	stmt, err := db.Prepare("INSERT INTO formdata (FirstName, LastName, Email, Phone, DOB,CVName,CVPath) VALUES (@FirstName, @LastName, @Email, @PhoneNumber, @DOB, @CVName, @FilePath)")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	// Execute the SQL statement with user data
	_, err = stmt.Exec(sql.Named("FirstName", user.FirstName), sql.Named("LastName", user.LastName), sql.Named("Email", user.Email), sql.Named("PhoneNumber", user.PhoneNumber), sql.Named("DOB", user.DOB), sql.Named("CVName", user.FileName), sql.Named("FilePath", filePath))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = sendEmail(user.Email, "Registration Confirmation", "Thank you for registering!")
	if err != nil {
	    http.Error(w, err.Error(), http.StatusInternalServerError)
	    return
	}


	response := map[string]string{"message": "Registration successful"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
