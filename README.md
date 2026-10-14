## 1. Dominio

La aplicación gestiona un sistema de eventos y tickets con tres entidades principales:
- **Usuarios (`usuarios`):** Almacena los usuarios registrados en la plataforma.
- **Eventos (`eventos`):** Administra la información de los eventos disponibles (nombre, capacidad máxima y fecha).
- **Tickets (`tickets`):** Modela las reservas de entradas asociadas a un usuario y a un evento.

---

## 2. Estructura del Módulo

El proyecto se organiza en los siguientes archivos y directorios:

- `db/schema.sql`: Sentencias `CREATE TABLE` con el esquema de la base de datos.
- `db/queries.sql`: Consultas SQL para las operaciones CRUD.
- `sqlc.yaml`: Archivo de configuración de `sqlc`.
- `db/sqlc/`: Código Go autogenerado por `sqlc`.
- `handlers/`: Handlers HTTP que implementan los **endpoints REST** de la API.
- `db_test.go`: Pruebas usando el paquete `testing` de Go.
- `Dockerfile`: Archivo de configuración para compilar y empaquetar la aplicación en Go.
- `docker-compose.yml`: Configuración para orquestar los contenedores de la aplicación Go y la base de datos PostgreSQL.
- `requests.sh`: Script ejecutable que prueba los endpoints REST de la API utilizando peticiones `curl`.
- `run.sh`: Script de orquestación que levanta los contenedores en Docker, ejecuta `requests.sh` y al finalizar apaga y limpia el entorno.
- `Makefile`: Automatización de tareas de desarrollo y testing.

---

## 3. Requisitos Previos

Para ejecutar y probar la aplicación se requiere tener instalado en el sistema:
- [Go](https://go.dev/) (v1.20+)
- [Docker](https://www.docker.com/) y Docker Compose
- [Make](https://www.gnu.org/software/make/) (opcional, para usar el Makefile)
- [sqlc](https://sqlc.dev/)
- [cURL](https://curl.se/) y `bash` para la ejecución de scripts

---

## 4. Instrucciones de Ejecución

Para esta entrega se incorporaron los endpoints REST de la API y un flujo automatizado de pruebas end-to-end con Docker Compose.

### A. Ejecución del Ciclo de Pruebas Integradas (API REST + Contenedores)
El proyecto incluye un script `run.sh` que automatiza todo el proceso de prueba de la API:
1. Levanta los contenedores de Go y PostgreSQL en segundo plano mediante `docker compose`.
2. Ejecuta el script `requests.sh` para validar el funcionamiento de los endpoints REST usando `curl`.
3. Cierra y elimina los contenedores y volúmenes creados.

Para ejecutar todo el ciclo, clonar el repositorio en la rama `tp2` y ejecutar:

```bash
./run.sh
