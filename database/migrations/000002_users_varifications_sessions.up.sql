CREATE TABLE
  "users" (
    id VARCHAR(10) PRIMARY KEY,

    first_name VARCHAR(25) NOT NULL,
    middle_name VARCHAR(25),
    last_name VARCHAR(25) NOT NULL,
    other_name VARCHAR(100),

    email VARCHAR(80) UNIQUE NOT NULL,
    contact VARCHAR(25) UNIQUE NOT NULL,

    password VARCHAR(255) NOT NULL,

    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_locked BOOLEAN NOT NULL DEFAULT FALSE,
    mfa_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    failed_attempts INT NOT NULL DEFAULT 0,
    notes TEXT,

    is_contact_verified BOOLEAN NOT NULL DEFAULT FALSE,
    is_email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    last_security_check TIMESTAMPTZ NULL,    
    last_password_change TIMESTAMPTZ NULL,
    last_login TIMESTAMPTZ NULL,

    accepted_terms BOOLEAN NOT NULL DEFAULT FALSE,

    role VARCHAR(25) NOT NULL DEFAULT 'USER', -- USER, STAFF, 
    type VARCHAR(25) NOT NULL DEFAULT 'EXTERNAL', -- INTERNAL, EXTERNAL

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
    updated_at TIMESTAMPTZ
  );

CREATE TABLE
 "user_profiles" (
    id VARCHAR(10) PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(10) ,
    gender VARCHAR(40) NOT NULL,
    avatar VARCHAR(255) UNIQUE ,
    date_of_birth DATE ,    
    country VARCHAR(60) NOT NULL DEFAULT 'KENYA',
    identity_document VARCHAR(60) NOT NULL DEFAULT 'NATIONAL',
    identity_number VARCHAR(60) NOT NULL,
    id_front VARCHAR(255) UNIQUE , 
    id_back VARCHAR(255) UNIQUE ,
    network_provider VARCHAR(60) NULL, -- AIRTEL , MTN, used to inticate provider
    accepted_marketing BOOLEAN NOT NULL DEFAULT FALSE,
    bio TEXT ,
    other JSON,
    region VARCHAR(80) ,    
    province VARCHAR(80) , 
    district VARCHAR(80) ,     
    city VARCHAR(80) ,     
    address VARCHAR(80) ,     
    latitude DOUBLE NOT NULL DEFAULT 0,
    longitude DOUBLE NOT NULL DEFAULT 0,   
    signature VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
    updated_at TIMESTAMPTZ
);


CREATE TABLE 
"branch_users" (
  branch BIGINT NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
  user_id VARCHAR(10) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  PRIMARY KEY (branch, user_id)
);

CREATE TABLE
"organization_users" (
  organization BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  user_id VARCHAR(10) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  PRIMARY KEY (organization, user_id)
);


CREATE TABLE 
"verifications" (
    id BIGSERIAL PRIMARY KEY,
    user_id VARCHAR(10) NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    otp VARCHAR(60) NOT NULL,
    purpose VARCHAR(60) NOT NULL DEFAULT 'LOGIN',
    channel VARCHAR(60) NOT NULL DEFAULT 'EMAIL',
    is_used BOOLEAN NOT NULL DEFAULT FALSE,
    expiry TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
);


CREATE TABLE
  "sessions" (
    id uuid PRIMARY KEY,
    user_id VARCHAR(10) NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    refresh_token VARCHAR NOT NULL,
    user_agent VARCHAR NOT NULL,
    client_ip VARCHAR NOT NULL,
    description TEXT,
    other JSON,
    platform VARCHAR(20),
    device VARCHAR(50),
    is_blocked BOOLEAN NOT NULL DEFAULT false,
    expiry TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
  );

