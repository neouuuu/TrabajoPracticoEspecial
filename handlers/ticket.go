package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	db "github.com/neouuuu/TrabajoPracticoEspecial/db/sqlc"
)

func ReservarTicket(queries *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			IDUsuario int32 `json:"id_usuario"`
			IDEvento  int32 `json:"id_evento"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "JSON inválido", http.StatusBadRequest)
			return
		}

		ticket, err := queries.ReservarTicket(r.Context(), db.ReservarTicketParams{
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
}

func EliminarTicket(queries *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		ticketID, err := strconv.ParseInt(idStr, 10, 32)
		if err != nil {
			http.Error(w, "ID de ticket inválido", http.StatusBadRequest)
			return
		}

		err = queries.EliminarTicket(r.Context(), int32(ticketID))
		if err != nil {
			http.Error(w, "Error al eliminar el ticket", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}