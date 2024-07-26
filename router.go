package main

import (
	"net/http"

	"github.com/NisalSP9/To-Do-Web/health"
	"github.com/NisalSP9/To-Do-Web/task"
	"github.com/NisalSP9/To-Do-Web/user"
)

func loadRoutes(router *http.ServeMux) {

	healthHandler := &health.Handler{}
	userHandler := &user.Handler{}
	taskHandler := &task.Handler{}

	//Health APIs
	router.HandleFunc("GET /", healthHandler.GetHealth)

	//User APIs
	router.HandleFunc("POST /signup", userHandler.Create)
	router.HandleFunc("POST /login", userHandler.Login)

	//Task APIs
	router.HandleFunc("POST /task", taskHandler.Create)
	router.HandleFunc("GET /tasks", taskHandler.GetTasksByUserID)
	router.HandleFunc("PUT /task", taskHandler.EditTaskDetails)
	router.HandleFunc("DELETE /task/{taskID}", taskHandler.DeleteTask)
	// router.HandleFunc("PUT /monster/{id}", handler.UpdateByID)
	// router.HandleFunc("GET /monster/{id}", handler.FindByID)
	// router.HandleFunc("DELETE /monster/{id}", handler.DeleteByID)

}
