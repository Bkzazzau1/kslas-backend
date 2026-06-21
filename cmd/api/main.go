package main

import (
	"log"
	"net/http"
	"os"

	"kslasbackend/internal/database"
	"kslasbackend/internal/handlers"
	"kslasbackend/internal/models"
)

func main() {
	db, err := database.Connect()
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}

	if err := db.AutoMigrate(
		&models.Department{},
		&models.Course{},
		&models.Assessment{},
		&models.Question{},
		&models.QuestionOption{},
		&models.QuestionAsset{},
		&models.StudentSubmission{},
		&models.StudentAnswer{},
	); err != nil {
		log.Fatalf("database migration failed: %v", err)
	}

	mux := http.NewServeMux()
	h := handlers.NewAssessmentHandler(db)
	h.RegisterRoutes(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("K-SLAS backend running on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
