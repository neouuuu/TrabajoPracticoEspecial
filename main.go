package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
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
	connStr := "postgres://root:secretpassword@localhost:5432/db?sslmode=disable"
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

	// 3. Crear el Router (http.ServeMux)
	mux := http.NewServeMux()

	// 4. Servidor de archivos estáticos (Frontend)
	fs := http.FileServer(http.Dir(rutaMedia))
	mux.Handle("/", fs)

	// ---------------------------------------------------------
	// 5. RUTAS REST
	// ---------------------------------------------------------

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

/*
Antes que nada un poco de teoría antes de la implementación de los Handler porque se me está haciendo un re quilombo en la cabeza y andar escribiendo el proceso me ayuda a asimilarlo

Según el diseño REST para APIs, el Javascript en el front-end va a ser la interface con el usuario y por ejemplo cuando este haga click en un botoncito de reservar va a ejecutarse esto:

fetch("/ticket", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ usuario_id: 5, evento_id: 12 })
})

Es acá de donde salen las URL para que el router sepa que handler invocar

En el diseño REST, el prefijo de la URL define el recurso principal sobre el que estás consultando:
/tickets/{id} Debe devolver la información de un ticket en específico (cuyo ID es el del ticket).
/usuarios/{id}/tickets Devuelve la lista de tickets pertenecientes a un usuario.
/eventos/{id}/tickets Devuelve los tickets (o cantidad de reservas) pertenecientes a un evento.

Lo que viaja por la red es un paquete de texto HTTP:

Si es un request:

POST /ticket HTTP/1.1
Content-Type: application/json

{"usuario_id": 5, "evento_id": 12}

Si es un response:

HTTP/1.1 201 Created
Content-Type: application/json

{"id": 104, "usuario_id": 5, "evento_id": 12}

Bien ahora como entre el back y el front se comunican por mensajes json, hay que extraer ese json del request y convertirlo a tipos de datos para poder meterlo en variables de Go, ejemplo:

var req struct {
		Nombre string `json:"nombre"`
		Email  string `json:"email"`
	}

if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
	http.Error(w, "JSON inválido", http.StatusBadRequest)
	return
}

Por suerte "encoding/json" es una librería que lo hace por nosotros

Así es que conseguimos un DTO

Recién ahora podemos invocar las funciones que nos generó el sqlc

Pero antes, explicando un poco más porque leyendo me volví a perder ajajaj, no entendía la ventaja de sqlc en el sentido del mapeo porque si ya tenes un struct para decodificar el json, 
porque no lo reuso para la consulta en la BD (sqlc te lo vuelve a mapear) pero va queriendo:

Primero como funciona SQL puro en GO:

1. Conectarse a la base de datos, db es un POOL de conexiones
db, err := sql.Open("postgres", "postgres://user:pass@localhost:5432/dbname?sslmode=disable")

2. Entender los métodos del db:

db.ExecContext      Operaciones que no devuelven filas (UPDATE, DELETE o INSERT sin RETURNING)
db.QueryRowContext  Operaciones que devuelven 1 fila (SELECT LIMIT 1 o INSERT con RETURNING)
db.QueryContext     Operaciones que devuelven 0, 1 o varias filas (SELECT *)

3. Entender la diferencia entre DTO (Data Transfer Object) y Entity:

DTO: Es el objeto que se usa exclusivamente para transportar datos entre el cliente (Frontend) y tu API (Backend). Describe exactamente lo que la API espera recibir o enviar en un JSON.
Entity: Es la representación exacta de una tabla de tu base de datos en el código de Go. Refleja todas las columnas, claves primarias, claves foráneas y tipos de datos del esquema SQL.

Estos DEBIERAN ser diferentes, el usuario no manda todos los datos que aparecen en la tabla de la BD, como el ID que se genera solo. Por esto son 2 cosas DIFERENTES.

4. Entender la diferencia entre lo que devuelve la base de datos y lo que entiende GO (para entender el .Scan()):

La base de datos y el runtime de Go tienen sistemas de tipos, memoria y protocolos de red distintos. La función .Scan() actúa como el puente de conversión y asignación en memoria entre ambos.
La BD envía: Bytes crudos con metadatos de columnas.
Go tiene: Un struct Usuario alojado en la memoria RAM con tipos estáticos fuertemente tipados.

5. Entender qué es el context:

La interfaz context.Context (del paquete estándar context) es una estructura diseñada para transmitir señales de cancelación, límites de tiempo (timeouts/deadlines) y metadatos a lo largo 
de la cadena de llamadas de un programa o entre distintas goroutines.

Entonces, r.Context() es el Contexto de la petición HTTP actual.

5. Veamos como quedaría una operación de INSERT con RETURNING:

//Definimos el DTO
type CrearUsuarioDTO struct {
	Nombre string `json:"nombre"`
	Email  string `json:"email"`
}

//Creamos una variable del DTO
var dto CrearUsuarioDTO

//Supongamos que estamos dentro de un handler con un request y el response, entonces decodificamos el JSON que viene en el Body del request (r.Body) dentro del DTO
err := json.NewDecoder(r.Body).Decode(&dto)
if err != nil {
	http.Error(w, "JSON inválido", http.StatusBadRequest)
	return
}

//Definimos la ENTITY de la BD donde vamos a mapear lo que nos devuelva la BD
type Usuario struct {
    ID       int64     `json:"id"`
    Nombre   string    `json:"nombre"`
    Email    string    `json:"email"`
    CreadoEn time.Time `json:"creado_en"`
}

//Definimos la QUERY como string 
query := `
    INSERT INTO usuarios (nombre, email) 
    VALUES ($1, $2) 
    RETURNING id, nombre, email, creado_en
`

//Acá se va a volcar el resultado de la query con el .Scan()
var usuarioGuardado Usuario

// QueryRowContext usa todo lo que definimos y devuelve ese *sql.Row
row := db.QueryRowContext(r.Context(), query, dto.Nombre, dto.Email)

// .Scan() lee esa fila de bytes y mapea valor por valor dentro de los punteros, entonces es ese usuarioGuardado la estructura que contiene lo que queremos devolver
err := row.Scan(
    &usuarioGuardado.ID,
    &usuarioGuardado.Nombre,
    &usuarioGuardado.Email,
    &usuarioGuardado.CreadoEn,
)

//Seteamos el header de respuesta como JSON
w.Header().Set("Content-Type", "application/json")

//Para que la respuesta HTTP incluya el código de éxito
w.WriteHeader(http.StatusCreated)

//Codificamos la Entity mapeada (usuarioGuardado) a JSON para enviarla al cliente. El encoding/json se da cuenta que campos codificar y cuales no leyendo los tags como `json:"id"`, `json:"-"` NUNCA se serializa a JSON
json.NewEncoder(w).Encode(usuarioGuardado)

BUEEEEEEEEEEEEEEEEEEEEEEEEEEEEENO después de saber todo esto vemos todas las cosas que sqlc nos salva de hacer:

NO definimos la Entity, por ende el .Scan() TAMPOCO

Lo que sí hacemos es escribir consultas en un archivo aparte y con "sqlc generate" se arma todo por nosotros
Despues solo hay que invocar las funciones que genera, con un parametro ejemplo db.CrearUsuarioParams para que le pasemos los parametros para el QueryRow(), y te devuelven directamente la Entity para usar para el Encode.


*/




// =========================================================
// HANDLERS DE USUARIOS
// =========================================================

// POST /usuarios
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

// GET /usuarios/{id}/tickets
func (s *Server) listarTicketsPorUsuarioHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	usuarioID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "ID de usuario inválido", http.StatusBadRequest)
		return
	}

	tickets, err := s.queries.ListarTicketsPorUsuario(r.Context(), usuarioID)
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

// =========================================================
// HANDLERS DE EVENTOS
// =========================================================

// POST /eventos
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

// GET /eventos
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

// GET /eventos/{id}
func (s *Server) obtenerEventoPorIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	eventoID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "ID de evento inválido", http.StatusBadRequest)
		return
	}

	evento, err := s.queries.ObtenerEventoPorID(r.Context(), eventoID)
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

// PUT /eventos/{id}
func (s *Server) editarEventoHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	eventoID, err := strconv.ParseInt(idStr, 10, 64)
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
		ID:        eventoID,
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

// GET /eventos/{id}/reservas
func (s *Server) obtenerCapacidadReservadaHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	eventoID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "ID de evento inválido", http.StatusBadRequest)
		return
	}

	cantidadReservas, err := s.queries.ObtenerCapacidadReservada(r.Context(), eventoID)
	if err != nil {
		http.Error(w, "Error al obtener la cantidad de reservas", http.StatusInternalServerError)
		return
	}

	respuesta := map[string]interface{}{
		"evento_id":         eventoID,
		"total_reservas": cantidadReservas,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(respuesta)
}

// =========================================================
// HANDLERS DE TICKETS
// =========================================================

// POST /tickets
func (s *Server) reservarTicketHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDUsuario int64 `json:"id_usuario"`
		IDEvento  int64 `json:"id_evento"`
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

// DELETE /tickets/{id}
func (s *Server) eliminarTicketHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	ticketID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "ID de ticket inválido", http.StatusBadRequest)
		return
	}

	err = s.queries.EliminarTicket(r.Context(), ticketID)
	if err != nil {
		http.Error(w, "Error al eliminar el ticket", http.StatusInternalServerError)
		return
	}

	// 204 No Content para indicar eliminación exitosa sin cuerpo de respuesta
	w.WriteHeader(http.StatusNoContent)
}