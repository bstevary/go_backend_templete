CREATE TABLE 
"permissions" (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(60) NOT NULL UNIQUE,
    code VARCHAR(55) NOT NULL UNIQUE,
    description TEXT,
    resource VARCHAR(60) NOT NULL,
    privacy VARCHAR(20) NOT NULL DEFAULT 'GE', -- GE: General, BRC: Branch, SYS: System ORG: Organization
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NULL DEFAULT NULL
);

CREATE TABLE 
"roles" (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(60) NOT NULL,
    reference BIGINT NOT NULL, -- branch id or organization id
    scope VARCHAR(10) NOT NULL DEFAULT 'ORG', -- Possible values: 'BRANCH', 'ORG'
    description TEXT NOT NULL,
    tag VARCHAR(60) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NULL DEFAULT NULL,
    UNIQUE (name, reference)   
    );


CREATE TABLE
 "privileges" (
    role BIGINT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission BIGINT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (role, permission)
);

CREATE TABLE 
"authorities" (
    user_id VARCHAR(10) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reference BIGINT NOT NULL, -- branch id or organization id
    scope VARCHAR(10) NOT NULL DEFAULT 'ORG', -- Possible values: 'BRANCH', 'ORG'    
    role BIGINT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, role)
);
