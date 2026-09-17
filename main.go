package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
	db "github.com/neouuuu/TrabajoPracticoEspecial/db/sqlc"
)

type Server struct {
	queries *db.Queries
}

func (s *Server) ticketHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
		case http.MethodPost:
			// Reservar un ticket usando sqlc
			// ticket, err := s.queries.ReservarTicket(r.Context(), ...)

		case http.MethodGet:
			// Buscar un ticket usando sqlc
			// ticket, err := s.queries.GetTicketByID(r.Context(), ...)

		case http.MethodDelete:
			// Borrar un ticket usando sqlc
			// err := s.queries.DeleteTicket(r.Context(), ...)

		default:
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
	}
}

const ruta_media = "./static"

func main() {
	// 1. Conexión a la base de datos
	connStr := "postgres://root:secretpassword@localhost:5432/db?sslmode=disable"
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error al abrir la BD: %v", err)
	}
	defer conn.Close()

	if err := conn.Ping(); err != nil {
		log.Fatalf("No se pudo conectar a la BD: %v", err)
	}

	// 2. Inicialización de sqlc y Server
	queries := db.New(conn)
	srv := &Server{
		queries: queries,
	}

	// 3. Servir el frontend (index.html, CSS, JS) desde ./static en la raíz "/"
	fs := http.FileServer(http.Dir(ruta_media))
	http.Handle("/", fs)

	// 4. Endpoint de la API REST para los tickets
	http.HandleFunc("/ticket", srv.ticketHandler)

	// 5. Iniciar servidor
	fmt.Println("Servidor corriendo en http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}