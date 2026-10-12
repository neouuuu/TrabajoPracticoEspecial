package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	db "github.com/neouuuu/TrabajoPracticoEspecial/db/sqlc"
)

func CrearEvento(queries *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Nombre    string    `json:"nombre"`
			Capacidad int32     `json:"capacidad"`
			Fecha     time.Time `json:"fecha"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "JSON inválido", http.StatusBadRequest)
			return
		}

		evento, err := queries.CrearEvento(r.Context(), db.CrearEventoParams{
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
}

func ListarEventos(queries *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		eventos, err := queries.ListarEventos(r.Context())
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
}

func ObtenerEventoPorID(queries *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		eventoID, err := strconv.ParseInt(idStr, 10, 32)
		if err != nil {
			http.Error(w, "ID de evento inválido", http.StatusBadRequest)
			return
		}

		evento, err := queries.ObtenerEventoPorID(r.Context(), int32(eventoID))
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
}

func EditarEvento(queries *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		eventoActualizado, err := queries.EditarEvento(r.Context(), db.EditarEventoParams{
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
}

func ObtenerCapacidadReservada(queries *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		eventoID, err := strconv.ParseInt(idStr, 10, 32)
		if err != nil {
			http.Error(w, "ID de evento inválido", http.StatusBadRequest)
			return
		}

		cantidadReservas, err := queries.ObtenerCapacidadReservada(r.Context(), int32(eventoID))
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
}