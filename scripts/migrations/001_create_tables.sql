CREATE TABLE tenants (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    nombre VARCHAR(255),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE users (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    tenant_id UUID REFERENCES tenants(id),
    nombre VARCHAR(255),
    correo VARCHAR(255) UNIQUE NOT NULL,
    contrasena_hash VARCHAR(255) NOT NULL,
    whatsapp VARCHAR(20),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE leads (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    nombre VARCHAR(255),
    telefono VARCHAR(20),
    correo VARCHAR(255),
    curp VARCHAR(18),
    canal VARCHAR(50),
    estado VARCHAR(50) DEFAULT 'nuevo',
    monto_credito DECIMAL,
    tipo_credito VARCHAR(50),
    zona_interes VARCHAR(255),
    caracteristicas_vivienda TEXT,
    fecha_visita TIMESTAMP,
    score INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE scores (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    lead_id UUID REFERENCES leads(id),
    score INTEGER,
    categoria VARCHAR(10),
    reasoning TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE interactions (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    lead_id UUID REFERENCES leads(id),
    user_id UUID REFERENCES users(id),
    tipo VARCHAR(50),
    nota TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE follow_ups (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    lead_id UUID REFERENCES leads(id),
    user_id UUID REFERENCES users(id),
    fecha_programada TIMESTAMP,
    estado VARCHAR(50) DEFAULT 'pendiente',
    nota TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);