CREATE TABLE 
"branches" (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    country VARCHAR(60) NOT NULL DEFAULT 'KENYA',
    region VARCHAR(100), 
    province VARCHAR(100), 
    district VARCHAR(100), 
    city VARCHAR(100),
    address TEXT,
    contact VARCHAR(60),
    email VARCHAR(100),
    other JSON,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    subscription VARCHAR(60) NOT NULL DEFAULT 'FREE',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NULL DEFAULT NULL
);

CREATE TABLE 
"organizations" (
    id BIGSERIAL PRIMARY KEY,
    branch BIGINT NOT NULL REFERENCES branches(id), 
    name VARCHAR(100) NOT NULL,
    code  VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    contact VARCHAR(60),
    email VARCHAR(100),
    country VARCHAR(60) NOT NULL DEFAULT 'KENYA',
    region VARCHAR(100), 
    province VARCHAR(100), 
    district VARCHAR(100), 
    city VARCHAR(100),
    address TEXT,
    other JSON,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NULL DEFAULT NULL
);