-- name: CrearUsuario :one
INSERT INTO usuarios (nombre, email)
VALUES ($1, $2)
RETURNING *;

-- name: CrearEvento :one
INSERT INTO eventos (nombre, capacidad, fecha)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ObtenerEventoPorID :one
SELECT * FROM eventos
WHERE id = $1 LIMIT 1;

-- name: ListarEventos :many
SELECT * FROM eventos
ORDER BY fecha ASC;

-- name: ReservarTicket :one
INSERT INTO tickets (id_usuario, id_evento)
VALUES ($1, $2)
RETURNING *;

-- name: ListarTicketsPorUsuario :many
SELECT 
    t.id AS ticket_id,
    e.id AS evento_id,
    e.nombre AS evento_nombre,
    e.fecha AS evento_fecha
FROM tickets t
JOIN eventos e ON t.id_evento = e.id
WHERE t.id_usuario = $1;

-- name: ObtenerCapacidadReservada :one
SELECT COUNT(*) FROM tickets
WHERE id_evento = $1;

-- name: EliminarTicket :exec
DELETE FROM tickets
WHERE id = $1;