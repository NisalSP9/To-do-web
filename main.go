package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/NisalSP9/To-Do-Web/middleware"
	"github.com/joho/godotenv"
)

func main() {

	godotenv.Load()

	portString := os.Getenv("PORT")
	if portString == "" {
		log.Fatal("PORT is not found in the environment")
	}

	router := http.NewServeMux()

	loadRoutes(router)

	stack := middleware.CreateStack(
		middleware.Logging,
		middleware.AllowCors,
		middleware.IsAuthenticated,
	)

	srv := &http.Server{
		Addr:    ":" + portString,
		Handler: stack(router),
	}

	fmt.Println("Server starting on port:", portString)

	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}

}
