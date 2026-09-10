## Dominio:

La aplicación gestiona un sistema de eventos y tickets con tres entidades:
* **Usuarios (`usuarios`):** Almacena los usuarios registrados en la plataforma.
* **Eventos (`eventos`):** Administra la información de los eventos disponibles (nombre, capacidad máxima y fecha).
* **Tickets (`tickets`):** Modela las reservas de entradas asociadas a un usuario y a un evento.

---

## 2. Estructura del Módulo

Se organiza en los siguientes archivos:

* `db/schema.sql`: Sentencias `CREATE TABLE` con el esquema de la base de datos.
* `db/queries.sql`: Consultas SQL para las operaciones CRUD.
* `sqlc.yaml`: Archivo de configuración de `sqlc`.
* `db/sqlc/`: Código Go autogenerado por `sqlc`.
* `db_test.go`: Pruebas usando el paquete `testing` de Go.
* `docker-compose.yml`: Archivo de configuración para el contenedor con postgres.

---

## 3. Requisitos Previos

Para ejecutar y probar la aplicación se requiere tener instalado en el sistema:
* [Go](https://go.dev/) (v1.20+)
* [Docker](https://www.docker.com/) y Docker Compose
* [Make](https://www.gnu.org/software/make/) (opcional, para usar el Makefile)
* [sqlc](https://sqlc.dev/)

---

## 4. Instrucciones de Ejecución

El proyecto incluye un script de automatización (`Makefile`) que realiza todo el ciclo de vida de las pruebas:
1. Limpia contenedores y volúmenes previos de Docker.
2. Genera el código Go mediante `sqlc generate`.
3. Levanta la base de datos PostgreSQL en un contenedor Docker en segundo plano.
5. Ejecuta la suite de pruebas integradas usando el paquete `testing` de Go (`go test -v ./...`).
6. Detiene y elimina los contenedores y volúmenes creados.

Para ejecutar todo el proceso, clonar el repositorio en la rama `tp2` y ejecutar:

```bash
make test