package health

import (
	"log"
	"net/http"
)

type Handler struct{}

func (h *Handler) GetHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	log.Println("received request to get a health")
	w.Write([]byte("Up and running!"))
}
