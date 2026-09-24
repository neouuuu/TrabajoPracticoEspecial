package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq"
	db "github.com/neouuuu/TrabajoPracticoEspecial/db/sqlc"
)

type Server struct {
	queries *db.Queries
}

const rutaMedia = "./static"

func main() {
	// 1. Conexión a la Base de Datos 
	// Se obtiene la URL de conexión desde la variable de entorno DATABASE_URL (definido en el compose) porque ahora GO está corriendo dentro de un contenedor Docker, ya no es localhost.
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

	// 2. Inicialización de sqlc y estructura Server
	queries := db.New(conn)
	srv := &Server{queries: queries}

	// 3. Crear el Router (http.ServeMux), más fácil manejar las URL REST.
	mux := http.NewServeMux()

	// 4. Servidor de archivos estáticos (Frontend)
	fs := http.FileServer(http.Dir(rutaMedia))
	mux.Handle("/", fs)


	// 5. RUTAS REST


	// USUARIOS
	mux.HandleFunc("POST /usuarios", srv.crearUsuarioHandler)
	mux.HandleFunc("GET /usuarios/{id}/tickets", srv.listarTicketsPorUsuarioHandler)

	// EVENTOS
	mux.HandleFunc("POST /eventos", srv.crearEventoHandler)
	mux.HandleFunc("GET /eventos", srv.listarEventosHandler)
	mux.HandleFunc("GET /eventos/{id}", srv.obtenerEventoPorIDHandler)
	mux.HandleFunc("PUT /eventos/{id}", srv.editarEventoHandler)
	mux.HandleFunc("GET /eventos/{id}/reservas", srv.obtenerCapacidadReservadaHandler)

	// TICKETS
	mux.HandleFunc("POST /tickets", srv.reservarTicketHandler)
	mux.HandleFunc("DELETE /tickets/{id}", srv.eliminarTicketHandler)

	// 6. Iniciar Servidor
	fmt.Println("Servidor corriendo en http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}


// HANDLERS DE USUARIOS


func (s *Server) crearUsuarioHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Nombre string `json:"nombre"`
		Email  string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	usuario, err := s.queries.CrearUsuario(r.Context(), db.CrearUsuarioParams{
		Nombre: req.Nombre,
		Email:  req.Email,
	})
	if err != nil {
		http.Error(w, "Error al crear usuario", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(usuario)
}

func (s *Server) listarTicketsPorUsuarioHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	usuarioID, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		http.Error(w, "ID de usuario inválido", http.StatusBadRequest)
		return
	}

	tickets, err := s.queries.ListarTicketsPorUsuario(r.Context(), int32(usuarioID))
	if err != nil {
		http.Error(w, "Error al obtener los tickets del usuario", http.StatusInternalServerError)
		return
	}

	if tickets == nil {
		tickets = []db.ListarTicketsPorUsuarioRow{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tickets)
}


// HANDLERS DE EVENTOS


func (s *Server) crearEventoHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Nombre    string    `json:"nombre"`
		Capacidad int32     `json:"capacidad"`
		Fecha     time.Time `json:"fecha"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	evento, err := s.queries.CrearEvento(r.Context(), db.CrearEventoParams{
		Nombre:    req.Nombre,
		Capacidad: req.Capacidad,
		Fecha:     req.Fecha,
	})
	if err != nil {
		http.Error(w, "Error al crear evento", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(evento)
}

func (s *Server) listarEventosHandler(w http.ResponseWriter, r *http.Request) {
	eventos, err := s.queries.ListarEventos(r.Context())
	if err != nil {
		http.Error(w, "Error al obtener eventos", http.StatusInternalServerError)
		return
	}

	if eventos == nil {
		eventos = []db.Evento{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(eventos)
}

func (s *Server) obtenerEventoPorIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	eventoID, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		http.Error(w, "ID de evento inválido", http.StatusBadRequest)
		return
	}

	evento, err := s.queries.ObtenerEventoPorID(r.Context(), int32(eventoID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Evento no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "Error al consultar evento", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(evento)
}

func (s *Server) editarEventoHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	eventoID, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		http.Error(w, "ID de evento inválido", http.StatusBadRequest)
		return
	}

	var req struct {
		Nombre    string    `json:"nombre"`
		Capacidad int32     `json:"capacidad"`
		Fecha     time.Time `json:"fecha"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	eventoActualizado, err := s.queries.EditarEvento(r.Context(), db.EditarEventoParams{
		ID:        int32(eventoID),
		Nombre:    req.Nombre,
		Capacidad: req.Capacidad,
		Fecha:     req.Fecha,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Evento no encontrado para actualizar", http.StatusNotFound)
			return
		}
		http.Error(w, "Error al actualizar evento", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(eventoActualizado)
}

func (s *Server) obtenerCapacidadReservadaHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	eventoID, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		http.Error(w, "ID de evento inválido", http.StatusBadRequest)
		return
	}

	cantidadReservas, err := s.queries.ObtenerCapacidadReservada(r.Context(), int32(eventoID))
	if err != nil {
		http.Error(w, "Error al obtener la cantidad de reservas", http.StatusInternalServerError)
		return
	}

	respuesta := map[string]interface{}{
		"evento_id":      eventoID,
		"total_reservas": cantidadReservas,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(respuesta)
}

// =========================================================
// HANDLERS DE TICKETS
// =========================================================

func (s *Server) reservarTicketHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDUsuario int32 `json:"id_usuario"`
		IDEvento  int32 `json:"id_evento"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	ticket, err := s.queries.ReservarTicket(r.Context(), db.ReservarTicketParams{
		IDUsuario: req.IDUsuario,
		IDEvento:  req.IDEvento,
	})
	if err != nil {
		http.Error(w, "Error al reservar el ticket", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ticket)
}

func (s *Server) eliminarTicketHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	ticketID, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		http.Error(w, "ID de ticket inválido", http.StatusBadRequest)
		return
	}

	err = s.queries.EliminarTicket(r.Context(), int32(ticketID))
	if err != nil {
		http.Error(w, "Error al eliminar el ticket", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}