# 🐹 El Cerebro del Backend: Guía Definitiva de Go (Golang) para Desarrollo Web

## 1. ¿Qué es Go y cuál es su filosofía?
Go (o Golang) es un lenguaje de programación creado por Google con un propósito muy claro: combinar la **velocidad de ejecución y compilación de C/C++** con la **agilidad y facilidad de desarrollo de Python o JavaScript**. 

Para la web, Go es una bestia porque:
* Es **compilado y fuertemente tipado**: No se ejecuta interpretado línea por línea, sino que se compila a un único archivo binario nativo (sin dependencias locas como un entorno de ejecución pesado tipo Node.js o JVM).
* Es **minimalista**: No tiene clases, herencia tradicional ni metaprogramación compleja. Todo se basa en funciones, estructuras (`struct`) y composición.
* Tiene **concurrencia nativa**: Diseñado desde el día uno para manejar miles de conexiones simultáneas sin fundir la CPU.

---

## 2. Las Rarezas de la Sintaxis de Go
Cuando venís de lenguajes como JavaScript, Python o Java, la sintaxis de Go te va a chocar un montón. Hay tres cosas fundamentales que rompen los esquemas al principio:

### A. Múltiples valores de retorno (y el famoso `, err`)
En Go, las funciones pueden devolver **más de un valor a la vez**. No existen las excepciones de la forma tradicional (`try / catch` con `throw`). En su lugar, por convención, cualquier operación que pueda fallar (abrir un archivo, hacer una consulta SQL, leer la red) **devuelve el resultado y un error como segundo parámetro**.

Ejemplo clásico:
```go
resultado, err := hacerAlgoQuePuedeFallar()
if err != nil {
    // Manejas el error acá mismo, sin excepciones mágicas
    return err
}
```

### B. El orden de los tipos (de derecha a izquierda)
En Go, primero declarás el nombre de la variable y al final el tipo de dato:
```go
var nombre string = "Juan"
// O usando inferencia de tipos:
edad := 25 
```
Y en las funciones, los parámetros van con el nombre primero y el tipo después:
```go
func sumar(a int, b int) int {
    return a + b
}
```

---

## 3. Estructuras (`struct`) y Métodos (El "No-POO" de Go)
Go **no tiene clases ni objetos** en el sentido clásico de la Programación Orientada a Objetos (no hay `class`, ni `extends`, ni `private/public` tradicionales).

En su lugar, usás **Structs** (que son solo moldes de datos puros sin lógica) y luego les **adherís funciones asociadas** mediante una sintaxis especial llamada *Method Receiver*.

### ¿Cómo funciona?
1. **Definís el Struct (Los datos):**
   ```go
   type Server struct {
       db  *sql.DB
       mux *http.ServeMux
   }
   ```
2. **Le asignas una función (El comportamiento / El método):**
   Para transformar ese struct en un "objeto" que tiene métodos propios, le ponés un *receptor* `(s *Server)` antes del nombre de la función:
   ```go
   // Esto convierte a 'Iniciar' en un método del struct Server
   func (s *Server) Iniciar() {
       http.ListenAndServe(":8080", s.mux)
   }
   ```
   *¿Por qué el `(s *Server)`?* Es simplemente el equivalente al `this` de JavaScript o `self` de Python. Te permite acceder a las propiedades internas del servidor (`s.db`, `s.mux`) desde adentro de la función.

---

## 4. El Servidor Web: `net/http`, Puertos y el Enrutador (`ServeMux`)
Go trae en su **biblioteca estándar** un servidor HTTP completo y potentísimo. No necesitás instalar frameworks gigantes como Express.js en Node o Spring Boot en Java para levantar una API.

### A. Escuchando en un puerto (`ListenAndServe`)
Para levantar el servidor y abrir el puerto de tu PC, usás `http.ListenAndServe`:
```go
func main() {
    // El primer parámetro es el puerto (""), el segundo es el enrutador (Handler)
    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        panic(err)
    }
}
```

### B. El Mux (Enrutador) y la diferencia entre `Handle` y `HandleFunc`
El `ServeMux` es el cartero del servidor: recibe una URL entrante y decide a qué función (Handler) delegarla.

* **`http.HandleFunc`**: Es una función de atajo. Le pasás la ruta y una **función común** que cumpla con la firma de un handler `func(w http.ResponseWriter, r *http.Request)`:
  ```go
  mux.HandleFunc("/usuarios", func(w http.ResponseWriter, r *http.Request) {
      w.Write([]byte("Hola usuarios"))
  })
  ```
* **`http.Handle`**: Es más general. En lugar de aceptar una función directa, acepta **cualquier estructura que implemente la interfaz `http.Handler`** (es decir, cualquier objeto que tenga un método `ServeHTTP(w, r)`).

### C. Sirviendo archivos estáticos (`FileServer`)
Para servir tu frontend (el `index.html`, los CSS y el `app.js`) directamente desde Go sin vueltas, usás un FileServer de la biblioteca estándar:
```go
// Expone toda la carpeta "./public" bajo la ruta "/static/"
fs := http.FileServer(http.Dir("./public"))
mux.Handle("/static/", http.StripPrefix("/static/", fs))
```

---

## 5. El Pool de Conexiones (`database/sql`)
Cuando te conectás a una base de datos en Go usando `database/sql`, la función `sql.Open` **no abre una conexión física instantánea**. Lo que hace es inicializar un **Pool de Conexiones (Connection Pool)** gestionado de forma automática y concurrente por Go.

* **¿Cómo funciona?** El pool mantiene un conjunto de conexiones abiertas y listas en la memoria. Cuando llega un request HTTP, Go toma una conexión libre del pool, ejecuta la query, y apenas termina, **devuelve la conexión al pool** para que otro request la reutilice.
* **Seguridad de hilos (Thread-safe):** El objeto `*sql.DB` está diseñado específicamente para ser compartido de manera segura entre múltiples goroutines al mismo tiempo sin que se pisen las conexiones.

---

## 6. Concurrencia Real: Las Gorutines y `net/http`
Esta es la joya de la corona de Go. 

Una **Goroutine** es un hilo de ejecución ultra liviano manejado por el runtime de Go (no por el sistema operativo). Mientras que un hilo tradicional de Java o C consume varios MB de memoria RAM, una goroutine arranca consumiendo apenas unos pocos KBs.

### ¿Cómo interviene esto en tu servidor web?
No tenés que preocuparte por configurar hilos ni callbacks asíncronos complejos. **Por cada petición HTTP que entra a tu servidor Go, el paquete `net/http` lanza automáticamente una nueva Goroutine dedicada exclusivamente a atender ese request.**

Si se conectan 10.000 usuarios al mismo tiempo:
1. El servidor recibe los 10.000 requests en paralelo.
2. Go dispara 10.000 goroutines independientes de forma instantánea.
3. El planificador interno de Go distribuye esas goroutines de manera eficiente entre los núcleos de tu procesador.
4. Tu código se ejecuta de arriba a abajo de forma síncrona y súper limpia, ¡pero por debajo corre de forma masivamente concurrente!