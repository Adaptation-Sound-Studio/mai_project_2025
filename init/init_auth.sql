CREATE TABLE IF NOT EXISTS users (
    user_id    BIGSERIAL PRIMARY KEY,
    name       VARCHAR(100) NOT NULL,
    login      VARCHAR(100) NOT NULL UNIQUE,
    password   VARCHAR(130) NOT NULL,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS roles (
    role_id BIGSERIAL PRIMARY KEY,
    role    VARCHAR(50) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS user_role (
    ur_id    BIGSERIAL PRIMARY KEY,
    user_id  BIGINT NOT NULL,
    role_id  BIGINT NOT NULL,
    
    CONSTRAINT fk_user_role_user FOREIGN KEY (user_id)
        REFERENCES users (user_id) ON DELETE CASCADE,
    CONSTRAINT fk_user_role_role FOREIGN KEY (role_id)
        REFERENCES roles (role_id) ON DELETE CASCADE,
    CONSTRAINT uq_user_role UNIQUE (user_id, role_id)
);

INSERT INTO roles (role) VALUES ('artist');
INSERT INTO roles (role) VALUES ('admin');