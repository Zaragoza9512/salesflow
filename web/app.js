// URL de tu API en producción
//const API_URL = 'https://salesflow-7rwk.onrender.com';
// URL local para desarrollo
const API_URL = 'http://localhost:8080';

// token JWT guardado en memoria
let token = '';

// leads cargados actualmente
let leadsData = [];

// función para mostrar solo una sección y ocultar las demás
function mostrarSeccion(seccion) {
    // ocultar todas las secciones
    document.getElementById('login-section').style.display = 'none';
    document.getElementById('dashboard-section').style.display = 'none';
    document.getElementById('nuevo-lead-section').style.display = 'none';

    // mostrar solo la que necesitamos
    document.getElementById(seccion).style.display = 'block';
}

// función que se ejecuta cuando el usuario hace clic en "Entrar"
async function login() {
    // leer los valores que escribió el usuario
    const correo = document.getElementById('correo').value;
    const password = document.getElementById('password').value;

    // llamar a la API de login
    const respuesta = await fetch(API_URL + '/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ correo, password })
    });

    const datos = await respuesta.json();

    // si el login fue exitoso, guardar el token y mostrar el dashboard
    if (datos.token) {
        token = datos.token;
        mostrarSeccion('dashboard-section');
        cargarLeads();
    } else {
        // mostrar error
        document.getElementById('login-error').textContent = 'Correo o contraseña incorrectos';
    }
}

// cargar leads desde la API y mostrarlos en la tabla
async function cargarLeads() {
    const respuesta = await fetch(API_URL + '/leads', {
        headers: { 'Authorization': 'Bearer ' + token }
    });

    leadsData = await respuesta.json();
    mostrarLeads(leadsData);
}

// mostrar leads en la tabla
function mostrarLeads(leads) {
    const tbody = document.getElementById('leads-body');
    
    // si no hay leads mostrar mensaje
    if (!leads || leads.length === 0) {
        tbody.innerHTML = '<tr><td colspan="6">No hay leads todavía</td></tr>';
        return;
    }

    // construir las filas de la tabla
    tbody.innerHTML = leads.map(lead => {
        const categoria = lead.score >= 70 ? 'HOT' : lead.score >= 40 ? 'WARM' : 'COLD';
        const clase = categoria.toLowerCase();
        const emoji = categoria === 'HOT' ? '🔴' : categoria === 'WARM' ? '🟡' : '🔵';

        return `
            <tr>
                <td class="${clase}">${emoji} ${lead.score}</td>
                <td>${lead.nombre}</td>
                <td>${lead.canal}</td>
                <td>${lead.telefono}</td>
                <td>${lead.estado}</td>
                <td>
                    <button onclick="verLead('${lead.id}')">Ver</button>
                </td>
            </tr>
        `;
    }).join('');
}

// filtrar leads por categoría
function filtrarLeads(categoria) {
    if (categoria === 'todos') {
        mostrarLeads(leadsData);
        return;
    }
    const filtrados = leadsData.filter(lead => {
        const cat = lead.score >= 70 ? 'HOT' : lead.score >= 40 ? 'WARM' : 'COLD';
        return cat === categoria;
    });
    mostrarLeads(filtrados);
}

// mostrar formulario de nuevo lead
function mostrarNuevoLead() {
    mostrarSeccion('nuevo-lead-section');
}

// volver al dashboard
function mostrarDashboard() {
    mostrarSeccion('dashboard-section');
    cargarLeads();
}

// cerrar sesión
function cerrarSesion() {
    token = '';
    leadsData = [];
    mostrarSeccion('login-section');
}

// crear un nuevo lead
async function crearLead() {
    const lead = {
        nombre: document.getElementById('lead-nombre').value,
        telefono: document.getElementById('lead-telefono').value,
        correo: document.getElementById('lead-correo').value,
        canal: document.getElementById('lead-canal').value,
        tipo_credito: document.getElementById('lead-tipo-credito').value,
        monto_credito: parseFloat(document.getElementById('lead-monto').value),
        zona_interes: document.getElementById('lead-zona').value,
    };

    const respuesta = await fetch(API_URL + '/leads', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer ' + token
        },
        body: JSON.stringify(lead)
    });

    const datos = await respuesta.json();

    if (datos.id) {
        document.getElementById('lead-mensaje').textContent = 'Lead creado y calificando con IA...';
        // esperar 3 segundos y volver al dashboard
        setTimeout(() => {
            mostrarDashboard();
        }, 3000);
    } else {
        document.getElementById('lead-mensaje').textContent = 'Error al crear el lead';
    }
}

// ver detalle de un lead
async function verLead(id) {
    const respuesta = await fetch(API_URL + '/leads/' + id, {
        headers: { 'Authorization': 'Bearer ' + token }
    });
    const lead = await respuesta.json();
    alert(`
Lead: ${lead.nombre}
Score: ${lead.score}
Canal: ${lead.canal}
Tipo crédito: ${lead.tipo_credito}
Monto: $${lead.monto_credito}
Zona: ${lead.zona_interes}
Estado: ${lead.estado}
    `);
}