const formUsuario = document.querySelector("#form-usuario");
const formEvento = document.querySelector("#form-evento");
const usuarioId = document.querySelector("#usuario-id");
const eventoId = document.querySelector("#evento-id");
const listaEventos = document.querySelector("#lista-eventos");
const listaTickets = document.querySelector("#lista-tickets");

//Función centralizada (wrapper) para hacer peticiones HTTP a la API de Go.
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
    // OBTENER LA LISTA DE EVENTOS:
    const eventos = await pedir("/eventos");

    if (!eventos) return;

    listaEventos.innerHTML = "";

    // RECORRER CADA EVENTO 
    for (const ev of eventos) {
        
        // CONSULTA CANTIDAD DE RESERVAS
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

// GET /usuarios/{id}/tickets
async function cargarTickets() {
    if (!usuarioId.value) return;
    const tickets = await pedir(`/usuarios/${usuarioId.value}/tickets`);
    if (!tickets) return;

    listaTickets.innerHTML = "";
    for (const t of tickets) {
        listaTickets.innerHTML += `
            <li class="ticket">
                <div class="ticket-cuerpo">
                    <h3>${t.EventoNombre}</h3>
                    <p class="ticket-fecha">${new Date(t.EventoFecha).toLocaleString("es-AR")}</p>
                </div>
                <div class="ticket-talon">
                    <button class="btn btn-peligro borrar" data-id="${t.TicketID}">Cancelar</button>
                </div>
            </li>`;
    }
}

// POST /usuarios
formUsuario.addEventListener("submit", async (e) => {
    e.preventDefault();
    const usuario = await pedir("/usuarios", "POST", {
        nombre: document.querySelector("#usuario-nombre").value,
        email: document.querySelector("#usuario-email").value,
    });
    if (usuario) {
        formUsuario.reset();
        usuarioId.value = usuario.ID;
        cargarTickets();
    }
});

// POST /eventos (crear) o PUT /eventos/{id} (editar)
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

// Reservar (POST /tickets) y Editar (GET /eventos/{id}) con delegación de eventos
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
        // pasamos la fecha al formato del input datetime-local
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

// DELETE /tickets/{id} (delegación de eventos sobre la lista)
listaTickets.addEventListener("click", async (e) => {
    if (!e.target.classList.contains("borrar")) return;
    const ok = await pedir(`/tickets/${e.target.dataset.id}`, "DELETE");
    if (ok) {
        cargarTickets();
        cargarEventos();
    }
});

document.querySelector("#btn-ver-tickets").addEventListener("click", cargarTickets);

// Al cargar la página
cargarEventos();