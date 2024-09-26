CREATE TABLE
  "users" (
    user_id BIGSERIAL PRIMARY KEY,
    first_name VARCHAR(25) NOT NULL,
    last_name VARCHAR(25) NOT NULL,
    email VARCHAR(80) UNIQUE NOT NULL,
    gender VARCHAR(10) CHECK (gender IN ('male', 'female', 'non-binary')) NOT NULL,
    hashed_password VARCHAR(255) NOT NULL,
    password_changed_at TIMESTAMPTZ,
    date_of_birth DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
    updated_at TIMESTAMPTZ,
    is_email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    is_account_active BOOLEAN NOT NULL DEFAULT FALSE
  );

CREATE TABLE
  "sessions" (
    id uuid PRIMARY KEY,
    email VARCHAR(80) NOT NULL REFERENCES users (email),
    refresh_token VARCHAR NOT NULL,
    user_agent VARCHAR NOT NULL,
    client_ip VARCHAR NOT NULL,
    is_blocked BOOLEAN NOT NULL DEFAULT false,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
  );

CREATE TABLE
  "varify_email" (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users (user_id) ON DELETE CASCADE,
    email VARCHAR(60) NOT NULL,
    secret_code VARCHAR(60) NOT NULL,
    is_used BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
    expired_at TIMESTAMPTZ NOT NULL DEFAULT (now () + interval '1 day')
  );