package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"kslasbackend/internal/config"
	"kslasbackend/internal/database"
	"kslasbackend/internal/handlers"
	"kslasbackend/internal/rbac"
	"kslasbackend/internal/repository"
	appserver "kslasbackend/internal/server"
	"kslasbackend/internal/services"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := database.Open(cfg)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("access sql db: %v", err)
	}
	defer sqlDB.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := database.Bootstrap(ctx, db, cfg); err != nil {
		log.Fatalf("bootstrap database: %v", err)
	}

	jwtService := services.NewJWTService(cfg)
	authRepository := repository.NewAuthRepository(db)
	academicRepository := repository.NewAcademicRepository(db)
	teachingRepository := repository.NewTeachingRepository(db)
	passwordService := services.NewPasswordService()
	authService := services.NewAuthService(authRepository, passwordService, jwtService)
	permissionService := services.NewPermissionService(rbac.NewAuthorizer(db))
	authHandler := handlers.NewAuthHandler(authService)
	academicService := services.NewAcademicService(academicRepository, permissionService)
	academicHandler := handlers.NewAcademicHandler(academicService)
	administrationService := services.NewAdministrationService(teachingRepository, passwordService, permissionService)
	administrationHandler := handlers.NewAdministrationHandler(administrationService)
	materialService := services.NewMaterialService(teachingRepository, permissionService)
	materialHandler := handlers.NewMaterialHandler(materialService)
	assignmentService := services.NewAssignmentService(teachingRepository, permissionService)
	assignmentHandler := handlers.NewAssignmentHandler(assignmentService)
	forumService := services.NewForumService(teachingRepository, permissionService)
	forumHandler := handlers.NewForumHandler(forumService)
	messageService := services.NewDirectMessageService(teachingRepository, permissionService)
	messageHandler := handlers.NewDirectMessageHandler(messageService)
	contentService := services.NewTeachingContentService(teachingRepository, permissionService)
	contentHandler := handlers.NewTeachingContentHandler(contentService)
	examService := services.NewExamService(teachingRepository, permissionService)
	examHandler := handlers.NewExamHandler(examService)
	invigilatorEvidenceHandler := handlers.NewInvigilatorEvidenceHandler()
	openAIReviewService := services.NewOpenAIPreExamReviewService(
		cfg.OpenAIAPIKey,
		cfg.OpenAIBaseURL,
		cfg.OpenAIReviewModel,
	)
	proctoringReviewHandler := handlers.NewProctoringReviewHandler(openAIReviewService)
	resultService := services.NewResultService(teachingRepository, permissionService)
	resultHandler := handlers.NewResultHandler(resultService)
	reportService := services.NewReportService(teachingRepository, permissionService)
	reportHandler := handlers.NewReportHandler(reportService)

	server := &http.Server{
		Addr: ":" + cfg.HTTPPort,
		Handler: appserver.NewRouter(&appserver.Dependencies{
			AuthHandler:                authHandler,
			AcademicHandler:            academicHandler,
			AdminHandler:               administrationHandler,
			MaterialHandler:            materialHandler,
			AssignmentHandler:          assignmentHandler,
			ForumHandler:               forumHandler,
			MessageHandler:             messageHandler,
			ContentHandler:             contentHandler,
			ExamHandler:                examHandler,
			InvigilatorEvidenceHandler: invigilatorEvidenceHandler,
			ProctoringReviewHandler:    proctoringReviewHandler,
			ResultHandler:              resultHandler,
			ReportHandler:              reportHandler,
			JWTService:                 jwtService,
			PermissionService:          permissionService,
		}),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("shutdown error: %v", err)
		}
	}()

	log.Printf("%s listening on http://localhost:%s", cfg.AppName, cfg.HTTPPort)

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server error: %v", err)
	}
}
