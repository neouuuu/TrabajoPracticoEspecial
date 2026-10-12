package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	db "github.com/neouuuu/TrabajoPracticoEspecial/db/sqlc"
)

func CrearUsuario(queries *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Nombre string `json:"nombre"`
			Email  string `json:"email"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "JSON inválido", http.StatusBadRequest)
			return
		}

		usuario, err := queries.CrearUsuario(r.Context(), db.CrearUsuarioParams{
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
}

func ListarTicketsPorUsuario(queries *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		usuarioID, err := strconv.ParseInt(idStr, 10, 32)
		if err != nil {
			http.Error(w, "ID de usuario inválido", http.StatusBadRequest)
			return
		}

		tickets, err := queries.ListarTicketsPorUsuario(r.Context(), int32(usuarioID))
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
}