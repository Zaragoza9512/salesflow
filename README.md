# SalesFlow

Sistema inteligente de gestión y calificación de leads inmobiliarios construido con Go, PostgreSQL, Redis e inteligencia artificial.

## ¿Por qué existe este proyecto?

Trabajé más de 3 años como SDR en empresas de PropTech como Houm y TuHabi, calificando leads manualmente todos los días. El problema real no era conseguir leads — era saber a cuál llamar primero y cuándo hacer seguimiento.

Construí SalesFlow para resolver ese problema: un CRM que califica automáticamente cada lead basado en reglas de negocio reales del mercado inmobiliario mexicano, genera un análisis en lenguaje natural con IA, y te dice exactamente a quién contactar hoy.

## Features

- Autenticación segura con JWT y bcrypt
- CRUD completo de leads con estados y seguimiento
- Motor de scoring automático basado en reglas de negocio reales
- Reasoning generado por IA (Groq/Llama) explicando por qué califica cada lead
- Cache con Redis para consultas rápidas
- Multi-tenancy — múltiples asesores con datos aislados
- CI/CD automatizado con GitHub Actions
- Deploy en producción con Docker

## Stack tecnológico

| Capa | Tecnología |
|---|---|
| Lenguaje | Go (Golang) |
| Router | Chi |
| Base de datos | PostgreSQL |
| Cache | Redis |
| IA / Reasoning | Groq API (Llama 3.3) |
| Autenticación | JWT + bcrypt |
| Contenedores | Docker |
| Infra como código | Terraform |
| CI/CD | GitHub Actions |
| Deploy | Render.com |

## Arquitectura

```
Cliente (curl / app)
        ↓
    Chi Router
        ↓
JWT Middleware (rutas protegidas)
        ↓
    Handlers
    ↙        ↘
PostgreSQL    Redis (cache)
                ↓
           Groq API (scoring IA)
```

## Cómo correrlo localmente

### Requisitos
- Go 1.23+
- Docker y Docker Compose

### Instalación

```bash
git clone https://github.com/Zaragoza9512/salesflow.git
cd salesflow
cp .env.example .env
docker compose up -d
go run cmd/api/main.go
```

### Variables de entorno

```env
APP_PORT=8080
APP_ENV=development
DB_HOST=localhost
DB_PORT=5432
DB_USER=salesflow
DB_PASSWORD=salesflow123
DB_NAME=salesflow_db
REDIS_HOST=localhost
REDIS_PORT=6379
JWT_SECRET=tu_clave_secreta
GROQ_API_KEY=tu_groq_key
```

## Endpoints

### Autenticación
```
POST /auth/register   Registrar nuevo usuario
POST /auth/login      Iniciar sesión → devuelve JWT
```

### Leads (requieren JWT)
```
POST   /leads          Crear lead → score automático con IA
GET    /leads          Listar leads ordenados por score
GET    /leads/{id}     Detalle de un lead
PUT    /leads/{id}     Actualizar lead
DELETE /leads/{id}     Archivar lead
```

### Ejemplo de respuesta

```json
{
  "id": "df6c7dfe-e4ad-4780-91e9-acd5dabc72ad",
  "nombre": "Carlos Mendez",
  "canal": "WhatsApp",
  "tipo_credito": "INFONAVIT",
  "monto_credito": 780000,
  "zona_interes": "Ecatepec",
  "score": 80,
  "estado": "nuevo"
}
```

### Reasoning generado por IA

> "Carlos Mendez tiene un alto score de 80/100 y se encuentra en la categoría HOT. Su monto de crédito de $780,000 es apenas superior al mínimo requerido para una propiedad nueva en Edomex, lo que sugiere que debe actuar con urgencia ya que los precios suben constantemente."

## Motor de Scoring

El score se calcula con reglas basadas en experiencia real del mercado:

| Factor | Puntos |
|---|---|
| Tiene tipo de crédito (INFONAVIT/FOVISSSTE) | +30 |
| Monto mayor o igual a $900,000 | +20 |
| Monto mayor o igual a $750,000 | +15 |
| Monto mayor o igual a $650,000 | +10 |
| Canal Referido | +20 |
| Canal WhatsApp | +15 |
| Canal Instagram/Facebook | +10 |
| Zona de interés definida | +10 |
| Nombre proporcionado | +10 |

**Categorías:**
- HOT: score mayor o igual a 70 — Contactar hoy
- WARM: score mayor o igual a 40 — Esta semana
- COLD: score menor a 40 — Cuando haya tiempo

## Roadmap

- [ ] Seguimiento automatizado con WhatsApp
- [ ] Agente conversacional para primer contacto
- [ ] Dashboard web para asesores
- [ ] Sistema de pagos (Stripe)
- [ ] Notificaciones push

## Producción

API disponible en: https://salesflow-7rwk.onrender.com

Health check: https://salesflow-7rwk.onrender.com/health

## Autor

Luis Eduardo Zaragoza Hernández — Backend Developer

[GitHub](https://github.com/Zaragoza9512) | [LinkedIn](https://linkedin.com/in/luis-zaragoza95)