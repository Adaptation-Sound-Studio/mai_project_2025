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

INSERT INTO users (name, login, password) VALUES
  ('Alice Smith', 'alice', '$2a$10$A/OiukfL8fbMpnC2mv.L6uD5OwooxR8oG3IxbMDZaSvm3vR8amn4e'),
  ('Bob Johnson', 'bob', '$2a$10$A/OiukfL8fbMpnC2mv.L6uD5OwooxR8oG3IxbMDZaSvm3vR8amn4e'),
  ('Charlie Brown', 'charlie', '$2a$10$A/OiukfL8fbMpnC2mv.L6uD5OwooxR8oG3IxbMDZaSvm3vR8amn4e');

INSERT INTO roles (role) VALUES ('artist') ON CONFLICT DO NOTHING;
INSERT INTO roles (role) VALUES ('admin') ON CONFLICT DO NOTHING;

INSERT INTO user_role (user_id, role_id) VALUES
  (1, 1),  
  (2, 2),  
  (3, 1),  
  (3, 2); 
