package main

import (
	"context"
	"database/sql"
	"testing"
	"time"

	// esto importa el codigo Go autogenerado con sqlc
	db "github.com/neouuuu/TrabajoPracticoEspecial/db/sqlc"

	_ "github.com/lib/pq" // Driver para Postgres
)

func Test(t *testing.T) {
	// 1. Conectar a la DB en Docker
	connStr := "postgres://root:secretpassword@localhost:5432/db?sslmode=disable"
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("No se pudo conectar a la DB: %v", err)
	}
	defer conn.Close()

	// 2. Instanciar Queries de sqlc
	queries := db.New(conn)
	ctx := context.Background()

	// 3. Probamos Crear Usuario
	usuario, err := queries.CrearUsuario(ctx, db.CrearUsuarioParams{
		Nombre: "Carlos Gardel",
		Email:  "carlos@tango.com",
	})
	if err != nil {
		t.Fatalf("Error al crear usuario: %v", err)
	}

	// 4. Verificación (Testing)
	if usuario.ID == 0 {
		t.Errorf("Se esperaba un ID válido de usuario, se obtuvo 0")
	}

	// 5. Probamos Crear Evento
	evento, err := queries.CrearEvento(ctx, db.CrearEventoParams{
		Nombre:    "Recital de Rock",
		Capacidad: 100,
		Fecha:     time.Now().Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("Error al crear evento: %v", err)
	}

	// 6. Probamos Reservar Ticket
	ticket, err := queries.ReservarTicket(ctx, db.ReservarTicketParams{
		IDUsuario: usuario.ID,
		IDEvento:  evento.ID,
	})
	if err != nil {
		t.Fatalf("Error al reservar ticket: %v", err)
	}

	if ticket.IDUsuario != usuario.ID {
		t.Errorf("El ticket debía pertenecer al usuario %d", usuario.ID)
	}
}
