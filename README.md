# IAccess — Gestión de eventos y tickets

Aplicación web para crear eventos y reservar entradas. Backend en **Go** con una API REST, base de datos **PostgreSQL** con acceso tipado mediante **sqlc**, y un frontend en **HTML + CSS + JavaScript puro** servido por el mismo servidor Go.

Trabajo Práctico Especial — rama `tp4`.

---

## Índice

1. [Dominio](#1-dominio)
2. [Stack](#2-stack)
3. [Estructura del proyecto](#3-estructura-del-proyecto)
4. [Requisitos previos](#4-requisitos-previos)
5. [Cómo ejecutarlo](#5-cómo-ejecutarlo)
6. [API REST](#6-api-rest)
7. [Frontend](#7-frontend)
8. [Tests](#8-tests)
9. [Documentación teórica](#9-documentación-teórica)

---

## 1. Dominio

El sistema modela tres entidades (definidas en [`db/schema.sql`](db/schema.sql)):

| Entidad | Tabla | Campos |
|---|---|---|
| **Usuario** | `usuarios` | `id`, `nombre`, `email` (único) |
| **Evento** | `eventos` | `id`, `nombre`, `capacidad`, `fecha` |
| **Ticket** | `tickets` | `id`, `id_usuario`, `id_evento`, `fecha` (momento de la reserva) |

Un ticket vincula a un usuario con un evento. Si se elimina un usuario o un evento, sus tickets se eliminan en cascada (`ON DELETE CASCADE`).

---

## 2. Stack

- **Go** (ver versión en `go.mod`) con `net/http` y el router estándar `http.ServeMux` (rutas con método y parámetros, ej. `GET /eventos/{id}`).
- **PostgreSQL 15**.
- **[sqlc](https://sqlc.dev/)**: genera código Go tipado a partir de las consultas SQL.
- **[lib/pq](https://github.com/lib/pq)**: driver de Postgres para `database/sql`.
- **Docker / Docker Compose** para la base de datos y para la app.
- **HTML + CSS + JS vanilla** para el frontend (renderizado del lado del cliente con `fetch`).

---

## 3. Estructura del proyecto

```
.
├── main.go                 # Punto de entrada: conexión a BD, rutas y servidor
├── handlers/               # Handlers HTTP (capa de lógica)
│   ├── usuario.go
│   ├── evento.go
│   └── ticket.go
├── db/
│   ├── schema.sql          # Esquema de la base de datos
│   ├── queries/            # Consultas SQL (entrada de sqlc)
│   │   ├── usuarios.sql
│   │   ├── eventos.sql
│   │   └── tickets.sql
│   └── sqlc/               # Código Go autogenerado por sqlc (no se versiona)
├── static/                 # Frontend
│   ├── index.html
│   ├── styles.css
│   └── app.js
├── tests/                  # Tests de integración contra Postgres
├── sqlc.yaml               # Configuración de sqlc
├── docker-compose.yml      # Servicios: db (Postgres) y app (servidor Go)
├── Dockerfile              # Build multi-stage del servidor
├── Makefile                # `make test`: ciclo completo de pruebas
├── dev.sh                  # Desarrollo: BD en Docker + servidor en tu máquina
├── run.sh                  # Levanta todo en Docker y corre requests.sh
├── requests.sh             # Recorrido de la API con curl
└── teoría_*.md             # Apuntes teóricos de cada capa
```

La aplicación está organizada en tres capas:

```
Frontend (static/)  ──HTTP/JSON──►  Handlers (handlers/)  ──►  Queries sqlc (db/sqlc)  ──►  PostgreSQL
```

---

## 4. Requisitos previos

- [Go](https://go.dev/) (versión indicada en `go.mod`)
- [Docker](https://www.docker.com/) y Docker Compose
- [sqlc](https://sqlc.dev/): `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`
- [Make](https://www.gnu.org/software/make/) (opcional, para `make test`)

El código de `db/sqlc/` **no está en el repositorio** (está en el `.gitignore`), así que hay que generarlo la primera vez después de clonar:

```bash
sqlc generate
```

---

## 5. Cómo ejecutarlo

Primero, clonar la rama:

```bash
git clone --branch tp4 https://github.com/neouuuu/TrabajoPracticoEspecial.git
cd TrabajoPracticoEspecial
```

### Opción A — Desarrollo (recomendada)

Levanta solo PostgreSQL en Docker y corre el servidor Go directamente en tu máquina. Genera el código de sqlc automáticamente si todavía no existe.

```bash
./dev.sh            # conserva los datos entre ejecuciones
./dev.sh --clean    # borra el contenedor y arranca con la BD vacía
```

Abrí <http://localhost:8080>. Con `Ctrl+C` se detiene el servidor y el contenedor de la base.

### Opción B — Todo en Docker

Compila y levanta la app y la base de datos con Docker Compose:

```bash
sqlc generate                 # necesario antes del build
docker compose up --build
```

La app (API + frontend) queda en <http://localhost:8080>. Para apagar y borrar todo:

```bash
docker compose down -v
```

### Opción C — Probar los endpoints con `curl`

`run.sh` genera el código de sqlc si hace falta, levanta los servicios con Docker Compose, espera a que la API responda, ejecuta [`requests.sh`](requests.sh) (un recorrido completo por todos los endpoints) y al terminar limpia los contenedores:

```bash
./run.sh
```

### Configuración

| Variable | Descripción | Valor por defecto |
|---|---|---|
| `DATABASE_URL` | Cadena de conexión a Postgres | `postgres://root:secretpassword@localhost:5432/db?sslmode=disable` |

El servidor escucha en el puerto **8080**. La base usa `tmpfs`, por lo que los datos viven en memoria y se pierden al bajar el contenedor de Postgres. El esquema se aplica automáticamente al crear el contenedor.

---

## 6. API REST

Todas las respuestas con cuerpo son JSON.

### Usuarios

| Método | Ruta | Descripción | Respuesta |
|---|---|---|---|
| `POST` | `/usuarios` | Crea un usuario | `201` + usuario |
| `GET` | `/usuarios/{id}/tickets` | Lista los tickets de un usuario (con datos del evento) | `200` + lista |

### Eventos

| Método | Ruta | Descripción | Respuesta |
|---|---|---|---|
| `POST` | `/eventos` | Crea un evento | `201` + evento |
| `GET` | `/eventos` | Lista todos los eventos (ordenados por fecha) | `200` + lista |
| `GET` | `/eventos/{id}` | Obtiene un evento | `200` / `404` |
| `PUT` | `/eventos/{id}` | Edita un evento | `200` / `404` |
| `GET` | `/eventos/{id}/reservas` | Cantidad de tickets reservados para el evento | `200` |

### Tickets

| Método | Ruta | Descripción | Respuesta |
|---|---|---|---|
| `POST` | `/tickets` | Reserva un ticket | `201` + ticket |
| `DELETE` | `/tickets/{id}` | Cancela un ticket | `204` |

### Ejemplos

**Crear usuario**

```bash
curl -X POST http://localhost:8080/usuarios \
  -H "Content-Type: application/json" \
  -d '{"nombre": "Neo Allende", "email": "neo@example.com"}'
```

**Crear evento**

```bash
curl -X POST http://localhost:8080/eventos \
  -H "Content-Type: application/json" \
  -d '{"nombre": "Recital de Rock", "capacidad": 100, "fecha": "2026-10-15T21:00:00Z"}'
```

**Reservar un ticket**

```bash
curl -X POST http://localhost:8080/tickets \
  -H "Content-Type: application/json" \
  -d '{"id_usuario": 1, "id_evento": 1}'
```

**Consultar reservas de un evento**

```bash
curl http://localhost:8080/eventos/1/reservas
# {"evento_id":1,"total_reservas":1}
```

### Códigos de error

| Código | Cuándo ocurre |
|---|---|
| `400` | JSON inválido o ID que no es un número |
| `404` | Evento inexistente (en `GET` y `PUT` por ID) |
| `500` | Error al consultar o escribir en la base de datos |

---

## 7. Frontend

Al abrir <http://localhost:8080> el servidor entrega la carpeta `static/`. La página permite:

- **Crear un usuario** (o ingresar el ID de uno existente).
- **Crear y editar eventos** (nombre, capacidad, fecha y hora).
- **Ver los eventos** con sus reservas actuales sobre la capacidad (`Reservas: X / Y`).
- **Reservar** una entrada para un evento.
- **Ver y cancelar** los tickets propios.

El frontend usa renderizado del lado del cliente: `app.js` consume la API con `fetch` y arma el DOM a partir del JSON recibido. Go solo expone la API y los archivos estáticos.

---

## 8. Tests

El proyecto se prueba en dos niveles, uno por cada capa:

| Qué se prueba | Cómo | Dónde |
|---|---|---|
| **Capa de datos** (esquema + queries de sqlc) | Tests de Go con el paquete `testing`, contra una base PostgreSQL real | [`tests/`](tests/) |
| **Handlers / API REST** | Recorrido de todos los endpoints con `curl` contra la app levantada | [`requests.sh`](requests.sh), ejecutado por [`run.sh`](run.sh) |

### Capa de datos (`go test`)

Los tests de [`tests/`](tests/) usan directamente el código generado por sqlc (crear usuario, crear evento, reservar ticket) y comparten helpers en `testing_helpers_test.go` para conectarse a la BD y crear datos de prueba.

Para ejecutar el ciclo completo (limpiar → generar sqlc → levantar Postgres → esperar → testear → limpiar):

```bash
make test
```

O manualmente, con la base ya levantada:

```bash
docker compose up -d db
go test -v ./tests/...
```

### Handlers (`requests.sh`)

```bash
./run.sh
```

Levanta todo, corre las 10 requests de `requests.sh` (crear usuario y evento, listar, obtener, editar, reservar, consultar reservas, listar tickets del usuario, cancelar y verificar) y limpia al terminar. También se puede correr `./requests.sh` solo, si la app ya está levantada en `localhost:8080`.

---

## 9. Documentación teórica

Durante el desarrollo se fueron escribiendo apuntes sobre cada capa de la aplicación:

- [`teoría_capa_datos.md`](teoría_capa_datos.md): base de datos, SQL y sqlc.
- [`teoría_capa_logica_go.md`](teoría_capa_logica_go.md): Go y handlers HTTP.
- [`teoría_capa_presentación.md`](teoría_capa_presentación.md): arquitectura web, SSR vs. CSR y frontend.