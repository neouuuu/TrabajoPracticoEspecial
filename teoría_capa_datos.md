# 🚀 Arquitectura Backend: Del Frontend a Go, SQL y `sqlc`

Antes que nada un poco de teoría antes de la implementación de los Handler porque se me está haciendo un re quilombo en la cabeza y andar escribiendo el proceso me ayuda a asimilarlo.

Según el diseño REST para APIs, el Javascript en el front-end va a ser la interfaz con el usuario y por ejemplo cuando este haga click en un botoncito de reservar va a ejecutarse esto:

```javascript
fetch("/ticket", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ usuario_id: 5, evento_id: 12 })
})
```

Es acá de donde salen las URL para que el router sepa qué handler invocar.

En el diseño REST, el prefijo de la URL define el recurso principal sobre el que estás consultando:
* `/tickets/{id}`: Debe devolver la información de un ticket en específico (cuyo ID es el del ticket).
* `/usuarios/{id}/tickets`: Devuelve la lista de tickets pertenecientes a un usuario.
* `/eventos/{id}/tickets`: Devuelve los tickets (o cantidad de reservas) pertenecientes a un evento.

---

## 📦 El Paquete HTTP en Movimiento

Lo que viaja por la red es un paquete de texto HTTP:

### Si es un request:
```http
POST /ticket HTTP/1.1
Content-Type: application/json

{"usuario_id": 5, "evento_id": 12}
```

### Si es un response:
```http
HTTP/1.1 201 Created
Content-Type: application/json

{"id": 104, "usuario_id": 5, "evento_id": 12}
```

---

## 🔄 Decodificación del JSON (De Texto a Go)

Bien, ahora como entre el back y el front se comunican por mensajes JSON, hay que extraer ese JSON del request y convertirlo a tipos de datos para poder meterlo en variables de Go, ejemplo:

```go
var req struct {
        Nombre string `json:"nombre"`
        Email  string `json:"email"`
    }

if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    http.Error(w, "JSON inválido", http.StatusBadRequest)
    return
}
```

Por suerte `"encoding/json"` es una librería que lo hace por nosotros. Así es que conseguimos un **DTO**. Recién ahora podemos invocar las funciones que nos generó sqlc.

---

## 🛠️ SQL Puro en Go vs. `sqlc`

Pero antes, explicando un poco más porque leyendo me volví a perder jajaj, no entendía la ventaja de sqlc en el sentido del mapeo porque si ya tenés un struct para decodificar el json, ¿por qué no lo reusás para la consulta en la BD (`sqlc` te lo vuelve a mapear)? Pero va queriendo:

### 1. Conexión a la Base de Datos
`db` es un POOL de conexiones:
```go
db, err := sql.Open("postgres", "postgres://user:pass@localhost:5432/dbname?sslmode=disable")
```

### 2. Entender los métodos del `db`:
* `db.ExecContext`: Operaciones que no devuelven filas (`UPDATE`, `DELETE` o `INSERT` sin `RETURNING`).
* `db.QueryRowContext`: Operaciones que devuelven 1 fila (`SELECT LIMIT 1` o `INSERT` con `RETURNING`).
* `db.QueryContext`: Operaciones que devuelven 0, 1 o varias filas (`SELECT *`).

### 3. DTO vs. Entity:
* **DTO (Data Transfer Object):** Es el objeto que se usa exclusivamente para transportar datos entre el cliente (Frontend) y tu API (Backend). Describe exactamente lo que la API espera recibir o enviar en un JSON.
* **Entity:** Es la representación exacta de una tabla de tu base de datos en el código de Go. Refleja todas las columnas, claves primarias, claves foráneas y tipos de datos del esquema SQL.

> *Estos DEBIERAN ser diferentes, el usuario no manda todos los datos que aparecen en la tabla de la BD, como el ID que se genera solo.*

### 4. Lo que devuelve la BD vs. lo que entiende Go (`.Scan()`):
La base de datos y el runtime de Go tienen sistemas de tipos, memoria y protocolos de red distintos. La función `.Scan()` actúa como el puente de conversión y asignación en memoria entre ambos:
* La BD envía: Bytes crudos con metadatos de columnas.
* Go tiene: Un struct `Usuario` alojado en la memoria RAM con tipos estáticos fuertemente tipados.

### 5. Qué es el `context`:
La interfaz `context.Context` (del paquete estándar `context`) es una estructura diseñada para transmitir señales de cancelación, límites de tiempo (*timeouts/deadlines*) y metadatos a lo largo de la cadena de llamadas de un programa o entre distintas goroutines. *(Por ejemplo, `r.Context()` es el Contexto de la petición HTTP actual).*

---

## 💻 Ejemplo de Operación de INSERT con RETURNING (SQL Puro)

```go
// Definimos el DTO
type CrearUsuarioDTO struct {
    Nombre string `json:"nombre"`
    Email  string `json:"email"`
}

var dto CrearUsuarioDTO

// Decodificamos el JSON que viene en el Body del request
err := json.NewDecoder(r.Body).Decode(&dto)
if err != nil {
    http.Error(w, "JSON inválido", http.StatusBadRequest)
    return
}

// Definimos la ENTITY de la BD
type Usuario struct {
    ID       int64     `json:"id"`
    Nombre   string    `json:"nombre"`
    Email    string    `json:"email"`
    CreadoEn time.Time `json:"creado_en"`
}

// Definimos la QUERY como string 
query := `
    INSERT INTO usuarios (nombre, email) 
    VALUES ($1, $2) 
    RETURNING id, nombre, email, creado_en
`

var usuarioGuardado Usuario

// QueryRowContext usa todo lo que definimos y devuelve ese *sql.Row
row := db.QueryRowContext(r.Context(), query, dto.Nombre, dto.Email)

// .Scan() mapea valor por valor dentro de los punteros
err = row.Scan(
    &usuarioGuardado.ID,
    &usuarioGuardado.Nombre,
    &usuarioGuardado.Email,
    &usuarioGuardado.CreadoEn,
)

w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusCreated)
json.NewEncoder(w).Encode(usuarioGuardado)
```

---

## ⚡ Lo que `sqlc` nos salva de hacer:

BUEEEEEEEEEEEEEEEEEEEEEEEEEEEEENO después de saber todo esto vemos todas las cosas que sqlc nos ahorra:
* **NO** definimos la Entity, por ende el `.Scan()` **TAMPOCO**.
* Lo que sí hacemos es escribir consultas en un archivo aparte y con `sqlc generate` se arma todo automáticamente.
* Después solo hay que invocar las funciones que genera pasándole los parámetros (ej. `db.CrearUsuarioParams`), y te devuelven directamente la Entity lista para usar en el `Encode`.

