package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
	db "github.com/neouuuu/TrabajoPracticoEspecial/db/sqlc"
	"github.com/neouuuu/TrabajoPracticoEspecial/handlers"
)

const rutaMedia = "./static"

func main() {
	// 1. Conexión a la Base de Datos
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://root:secretpassword@localhost:5432/db?sslmode=disable"
	}

	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error al abrir la BD: %v", err)
	}
	defer conn.Close()

	if err := conn.Ping(); err != nil {
		log.Fatalf("No se pudo conectar a la BD: %v", err)
	}

	// 2. Inicialización de sqlc
	queries := db.New(conn)

	// 3. Crear el Router (http.ServeMux)
	mux := http.NewServeMux()

	// 4. Servidor de archivos estáticos (Frontend)
	fs := http.FileServer(http.Dir(rutaMedia))
	mux.Handle("/", fs)

	// 5. RUTAS REST (Inyectando queries directamente en cada handler)

	// USUARIOS
	mux.HandleFunc("POST /usuarios", handlers.CrearUsuario(queries))
	mux.HandleFunc("GET /usuarios/{id}/tickets", handlers.ListarTicketsPorUsuario(queries))

	// EVENTOS (ejemplos)
	mux.HandleFunc("POST /eventos", handlers.CrearEvento(queries))
	mux.HandleFunc("GET /eventos", handlers.ListarEventos(queries))
	mux.HandleFunc("GET /eventos/{id}", handlers.ObtenerEventoPorID(queries))
	mux.HandleFunc("PUT /eventos/{id}", handlers.EditarEvento(queries))
	mux.HandleFunc("GET /eventos/{id}/reservas", handlers.ObtenerCapacidadReservada(queries))

	// TICKETS (ejemplos)
	mux.HandleFunc("POST /tickets", handlers.ReservarTicket(queries))
	mux.HandleFunc("DELETE /tickets/{id}", handlers.EliminarTicket(queries))

	// 6. Iniciar Servidor
	fmt.Println("Servidor corriendo en http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}