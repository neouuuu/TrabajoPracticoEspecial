# 🌐 Arquitectura Web Moderna: De la Red al DOM

---

## 1. Modelos de Renderizado: SSR vs. CSR

### A. Server-Side Rendering (SSR) — Lo que veíamos con templ
Cómo funciona: El servidor de Go hace todo el trabajo pesado. Recibe la petición, va a la base de datos, dibuja el HTML completo con los datos adentro, y le manda el plato servido al navegador.

El navegador: Solo recibe HTML plano y lo muestra. El servidor controla la vista.

### B. Client-Side Rendering (CSR) — Lo que te pide tu TP (HTML + CSS + JS puro)
Cómo funciona: El servidor te manda un index.html prácticamente vacío (o con una estructura básica) y un archivo app.js.

El navegador (El cliente): Descarga ese JavaScript, ejecuta un fetch() a tu API REST, recibe un JSON plano, y él mismo se pone a construir el HTML en tiempo real usando código (creando elementos, inyectándolos en el <div>, etc.).

Tu servidor Go: En este modelo, Go no sabe nada de HTML. Go es únicamente una API REST muda que devuelve datos en formato JSON.

¿Dónde encaja React (o Vue / Svelte)? ¿Es para simplificar esto?
Exactamente, diste en el clavo.

Hacer manipulación del DOM con JavaScript vainilla (puro) —como te pide la consigna del TP— se vuelve un dolor de cabeza gigante apenas tu aplicación crece un poquito:

Tenés que crear elementos a mano con document.createElement.

Tenés que escuchar eventos a mano con delegación compleja.

Si borrás un elemento, tenés que acordarte de actualizar la lista visualmente sin romper nada.

¿Qué hace React?
React es una librería que vino a simplificar y automatizar el Client-Side Rendering. En lugar de escribir código imperativo de cómo manipular el HTML paso a paso con JavaScript puro, en React declarás "cómo debería verse la pantalla según los datos (estado)", y React se encargara de actualizar el DOM por vos automáticamente.

Resumen de tu situación actual:
Tu API de Go maneja los datos (/usuarios, /eventos, /tickets).

Tu index.html + app.js hacen el trabajo sucio en el navegador (consumen la API y dibujan la interfaz).

Estás haciendo Client-Side Rendering a mano (con JavaScript puro) para entender las bases de cómo vive una aplicación web moderna por dentro, antes de usar frameworks mágicos como React.

Cuando el navegador descarga tu archivo index.html, no se queda leyendo texto plano todo el tiempo. Lo lee una sola vez y construye en la memoria RAM un árbol de objetos que representa a cada etiqueta, atributo y texto de tu página. A ese árbol se le llama DOM (Document Object Model).

¿Cómo interviene JavaScript en todo esto?
El HTML estático inicial: El navegador pinta la pantalla por primera vez basándose en el DOM inicial que armó con el index.html.

El poder de JavaScript: El motor de JavaScript tiene acceso directo a ese objeto global llamado document.

La manipulación: Cuando en tu app.js hacés un document.createElement('div') o un contenedor.appendChild(...), no estás escribiendo sobre el archivo index.html de tu computadora; estás modificando en tiempo real el árbol DOM que vive en la memoria del navegador.

El reflejo visual: Tan pronto como JavaScript modifica ese árbol, el navegador automáticamente redibuja (repaint/reflow) los píxeles en la pantalla. Por eso la página se actualiza sin necesidad de recargar (sin hacer F5).

El ciclo completo con tu API de Go:
La página carga el index.html básico (un contenedor vacío).

Se ejecuta el app.js y hace un fetch() a tu API REST de Go.

Go responde con un JSON (ej. una lista de eventos).

JavaScript procesa ese JSON, manipula el DOM creando elementos <li> o <div> con esos datos, y los inyecta en el contenedor.

¡Básicamente acabas de entender el principio de funcionamiento de todo el desarrollo frontend!




Más teoría: El navegador, el motor de JavaScript y el asíncronismo:

JavaScript is single-threaded: tiene una sola Call Stack (pila de ejecución). Para no bloquear la interfaz gráfica cuando hacés un fetch o esperas una respuesta, 
el navegador utiliza Web APIs, la Task Queue (Macrotareas), la Microtask Queue y el Event Loop.

1. La Call Stack 
Es donde JavaScript ejecuta tu código de forma secuencial LIFO.

Si una operación síncrona tarda, la pila se congela y la pestaña del navegador deja de responder.

2. Las Web APIs (El "Backstage" del Navegador)
Cuando en tu código ejecutás un fetch() o un setTimeout(), JavaScript no se queda esperando.

JavaScript libera la Call Stack inmediatamente y sigue ejecutando el resto del código. La operación de red ocurre fuera del hilo principal de JS, en los hilos de fondo del navegador (Web APIs).
(Esto significa que el motor de JavaScript es solo uno de los componentes del navegador, tiene varios más como se ve a continuación)

3. Microtareas vs. Macrotareas (Las Colas de Espera)
Cuando las operaciones asíncronas terminan su ejecución en los hilos de las Web APIs, sus callbacks (las funciones que debían ejecutarse después) no vuelven directo a la pila. Van a unas salas de espera 
divididas en dos jerarquías:

A. Microtask Queue — Prioridad Alta
Aquí van las promesas: los .then(), el código que está después de un await y queueMicrotask.

El motor las procesa apenas se vacía la Call Stack, antes de cualquier otra cosa.

B. Macrotask Queue — Prioridad Normal
Aquí van los setTimeout, setInterval, eventos de I/O, clics del mouse, y el procesamiento inicial de eventos de red como las respuestas completas de un fetch.

4. El Event Loop (El Director de Orquesta)
El Event Loop (Bucle de Eventos) es un ciclo infinito que hace una sola pregunta todo el tiempo:

"¿Está la Call Stack vacía?"

Si la respuesta es sí, el Event Loop ejecuta el siguiente algoritmo estricto:

Vaciar la Microtask Queue: Toma todas las microtareas acumuladas (promesas, await) y las manda a la Call Stack para que se ejecuten hasta que la cola de microtareas quede en cero. Si una microtarea genera otra microtarea,
se ejecuta también antes de pasar al siguiente paso.

Renderizado (opcional): El navegador aprovecha para actualizar la pantalla, repintar píxeles o calcular layouts si hace falta.

Tomar una Macrotask: Toma una sola macrotarea de la Macrotask Queue (por ejemplo, el callback de un setTimeout), la sube a la Call Stack y la ejecuta.

Repite el ciclo infinitamente.

¿Cómo se aplica esto a tu código (pedir)?
Cuando hacés esto en tu aplicación:

async function pedir(url, metodo = "GET", datos) {
    const res = await fetch(url, { ... }); // 1. Pausa asíncrona
    return res.json();                     // 2. Microtarea
}

La posta entre Macrotarea y Microtarea
La Macrotasla (El evento de red físico):
Cuando el servidor de Go termina de enviar los datos por internet, el navegador recibe los paquetes físicos en su hilo de red. El navegador registra eso como una tarea de I/O (Entrada/Salida) y coloca una 
Macrotarea en la Task Queue para procesar la llegada de esa respuesta.

La ejecución de la Macrotarea:
El Event Loop agarra esa macrotasla y la ejecuta. ¿Qué hace esa macrotasla? Básicamente, procesa los datos que llegaron y resuelve la Promesa del fetch (cambia su estado de "pendiente" a "éxito").

La Microtarea (Despertar tu código):
En el preciso instante en que esa promesa se resuelve, el motor dice: "¡Che, la promesa ya está lista!" e inmediatamente encola una Microtarea en la Microtask Queue.

El salto al código:
Como las microtareas tienen prioridad VIP, el Event Loop las corre al instante, sacando de la pausa a tu await y ejecutando la línea siguiente de tu función pedir.



Entonces el navegador posee los siguientes componentes principales:

El motor de JavaScript, un hilo que ejecuta segun la Call Stack y donde corre el Event Loop junto a las queues de microtareas y macrotareas.
Las Web APIs (fetch, setTimeout, etc.).
El Motor de Renderizado para pintar la pantalla y calcular layouts.
Capa de Almacenamiento: LocalStorage, SessionStorage, etc.


Finalmente: Como se utiliza Javascript para actualizar dinámicamente el DOM:

Arquitectura y Comunicacion entre HTML y JavaScript:

1. El concepto de "Evento": Dos mundos paralelos

El Evento del DOM (Navegador): Es una señal que emite el navegador cuando ocurre una accion fisica o logica por parte del usuario o del sistema (por ejemplo, hacer un click, enviar un formulario con submit, o terminar de cargar la pagina con load). Estos eventos se escriben por convencion en minusculas y entre comillas.

2. Escuchadores de Eventos (addEventListener)
Para que JavaScript pueda reaccionar a lo que hace el usuario tiene que estar a la espera de inputs.

El Vínculo Inicial: Mediante selectores como document.querySelector, JavaScript va a buscar un elemento fisico de la estructura HTML (un boton, un formulario, una lista) y lo guarda en una variable en memoria.

El Escuchador (Listener): Al invocar elemento.addEventListener("tipoDeEvento", (e) => { ... }), colocamos un observador atado a ese elemento. Espera a que suceda el evento especificado.

El Objeto de Evento (e): En el milisegundo en que ocurre la accion, el navegador empaqueta toda la informacion sobre el suceso en un objeto (por convencion llamado e). Este objeto contiene datos vitales como el target
del impacto y acciones como e.preventDefault(), que frena el comportamiento nativo del navegador (como evitar que un formulario recargue la pagina entera).

3. Delegación de Eventos: Eficiencia y Escalabilidad
Cuando tenes listas dinámicas de elementos asignar un escuchador individual a cada boton nuevo se vuelve ineficiente.

Aqui es donde entra la Delegación de Eventos:

El Principio del Burbujeo: Cuando haces clic en un elemento interno, el evento no se queda ahi; sube en forma de burbuja hacia su padre directo, luego hacia su abuelo, hasta llegar a la raiz del documento.

El Patrón: En lugar de ponerle un listener a cada boton hijo, le ponemos uno solo al elemento padre estático (por ejemplo, el contenedor general de la lista).

El Filtro Inteligente: Cuando el padre recibe la burbuja del clic, utiliza e.target para averiguar exactamente quien recibio el golpe original. Mediante condiciones como e.target.classList.contains("reservar") o leyendo atributos personalizados como data-id="${ev.ID}", el sistema decide si debe ejecutar una accion o ignorarla.

4. El Puente Invisible: Hidden IDs y la Reutilización de Formularios
¿Cómo sabe un formulario si estamos creando un elemento nuevo desde cero o si estamos editando uno existente?

La Memoria Oculta: En tu HTML colocas un campo invisible (<input type="hidden" id="evento-id">). Este campo no se dibuja en la pantalla para el usuario, pero actua como una caja fuerte o un puente de memoria temporal en JavaScript.

El Modo Creación: Al arrancar la pagina, ese input esta vacio (""). Cuando envias el formulario, el codigo evalua el estado y, al ver que no hay ID, ejecuta una peticion de creacion (POST).

El Modo Edición: Cuando el usuario hace clic en el boton "Editar" de un evento especifico, la aplicacion rescata el ID de ese elemento y se lo inyecta a escondidas al input oculto (eventoId.value = ev.ID), rellenando ademas los campos visuales con los datos actuales y cambiando temporalmente los textos del boton (de "Crear evento" a "Guardar cambios").

La Bifurcación Inteligente: Al hacer clic nuevamente en guardar, el formulario simplemente consulta: "¿Hay un ID guardado en el input oculto?". Si hay un numero, ejecuta una actualizacion (PUT al endpoint correspondiente); si esta vacio, ejecuta una insercion (POST). Al terminar, limpia todo y el sistema vuelve a su estado original.



RESUMIENDO:

En resumen, una SPA (Single Page Application) funcional basada en esta arquitectura opera bajo un ciclo claro y predecible:

Estado Visual y Datos: El HTML provee la estructura estatica y los contenedores limpios. El index.html simple que provee el servidor de archivos del Backend.

Peticiones Asíncronas: Funciones centralizadas basadas en fetch, promesas y await se comunican con el backend para traer o enviar datos en formato JSON, manejando los estados de red de forma segura (res.ok, control de errores con res.text()). Por esto es crucial que el Backend maneje bien los errores y los códigos de operación, también adjuntando mensajes personalizados de error. Esto lo puede hacer una función wrapper que centralice las peticiones al Backend (el pedir() de mi codigo).

Renderizado Dinámico: Se procesan los arreglos de datos utilizando plantillas de texto e interpolacion (innerHTML += ...) para inyectar elementos interactivos dotados de identificadores (data-id).

Interacción Centralizada: Los eventos del usuario son capturados mediante delegación en elementos padres, actualizando las variables globales o los inputs ocultos para coordinar las acciones de la interfaz de punta a punta.


EJEMPLOS DE CODIGO:

```javascript
// Elementos del DOM (los elementos que vas declarando en el HTML), ahora se le asignan a variables para poder manipularlos desde JS. Se buscan por el id declarado en el HTML.
const formUsuario = document.querySelector("#form-usuario");
const formEvento = document.querySelector("#form-evento");
const usuarioId = document.querySelector("#usuario-id");
const eventoId = document.querySelector("#evento-id");
const listaEventos = document.querySelector("#lista-eventos");
const listaTickets = document.querySelector("#lista-tickets");


// Función centralizada (wrapper) para hacer peticiones HTTP a la API de Go.
async function pedir(url, metodo = "GET", datos) {
    const res = await fetch(url, {
        method: metodo,
        headers: { "Content-Type": "application/json" },
        body: datos ? JSON.stringify(datos) : undefined,
    });

    if (!res.ok) {
        const mensajeError = await res.text();
        alert(mensajeError);
        return null;
    }

    return res.status === 204 ? true : await res.json();
}


async function cargarEventos() {
    const eventos = await pedir("/eventos");
    if (!eventos) return;

    listaEventos.innerHTML = "";

    for (const ev of eventos) {
        const r = await pedir(`/eventos/${ev.ID}/reservas`);
        const reservas = r ? r.total_reservas : "?";
        
        listaEventos.innerHTML += `
            <li class="ticket">
                <div class="ticket-cuerpo">
                    <h3>${ev.Nombre}</h3>
                    <p class="ticket-fecha">${new Date(ev.Fecha).toLocaleString("es-AR")}</p>
                    <p>Reservas: ${reservas} / ${ev.Capacidad}</p>
                </div>
                <div class="ticket-talon">
                    <button class="btn btn-primario reservar" data-id="${ev.ID}">Reservar</button>
                    <button class="btn btn-secundario editar" data-id="${ev.ID}">Editar</button>
                </div>
            </li>`;
    }
}


listaEventos.addEventListener("click", async (e) => {
    const id = e.target.dataset.id;

    if (e.target.classList.contains("reservar")) {
        if (!usuarioId.value) {
            alert("Primero creá un usuario o escribí tu ID");
            return;
        }
        
        const ticket = await pedir("/tickets", "POST", {
            id_usuario: Number(usuarioId.value),
            id_evento: Number(id),
        });
        
        if (ticket) {
            cargarTickets();
            cargarEventos();
        }
    }

    if (e.target.classList.contains("editar")) {
        const ev = await pedir(`/eventos/${id}`);
        if (!ev) return;

        const fecha = new Date(ev.Fecha);
        fecha.setMinutes(fecha.getMinutes() - fecha.getTimezoneOffset());

        eventoId.value = ev.ID;
        document.querySelector("#evento-nombre").value = ev.Nombre;
        document.querySelector("#evento-capacidad").value = ev.Capacidad;
        document.querySelector("#evento-fecha").value = fecha.toISOString().slice(0, 16);
        document.querySelector("#evento-titulo").textContent = "Editar evento";
        document.querySelector("#btn-evento").textContent = "Guardar cambios";
    }
});


formEvento.addEventListener("submit", async (e) => {
    e.preventDefault();
    const datos = {
        nombre: document.querySelector("#evento-nombre").value,
        capacidad: Number(document.querySelector("#evento-capacidad").value),
        fecha: new Date(document.querySelector("#evento-fecha").value).toISOString(),
    };
    const id = eventoId.value;
    const evento = id
        ? await pedir(`/eventos/${id}`, "PUT", datos)
        : await pedir("/eventos", "POST", datos);
    if (evento) {
        formEvento.reset();
        eventoId.value = "";
        document.querySelector("#evento-titulo").textContent = "Nuevo evento";
        document.querySelector("#btn-evento").textContent = "Crear evento";
        cargarEventos();
    }
});