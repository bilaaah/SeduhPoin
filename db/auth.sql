-- Autentikasi dan Session SeduhPoin
USE seduhpoin;

CREATE TABLE IF NOT EXISTS users (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    nama VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role ENUM('ADMIN', 'KASIR', 'MEMBER') NOT NULL DEFAULT 'MEMBER',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS sessions (
    token VARCHAR(64) PRIMARY KEY,
    user_id BIGINT NOT NULL,
    expires_at DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- AKUN SISTEM UNTUK SIMULASI (Password: admin123, kasir123, member123)
INSERT INTO users (id, nama, email, password_hash, role) VALUES
(1, 'Pemilik Kedai (Admin)', 'admin@seduhkopi.com', '$2a$10$KPQoj7P.EAmlSLoSAAfFtuMF42IOayaW2jtlLMMGacBwFja/xk6Me', 'ADMIN'),
(2, 'Kasir Kedai Seduh', 'kasir@seduhkopi.com', '$2a$10$sPRcFmGKAwipGlTTJ/62UO7NRWEfB108zfbB34wKqfQ61hDEre6N6', 'KASIR'),
(3, 'Budi (Member Loyal)', 'member@seduhkopi.com', '$2a$10$qIbJFbrmTU.sRWp7JC5RiO6bOvntjkPlUf9Iswa/PtiKKl.RUMox6', 'MEMBER')
ON DUPLICATE KEY UPDATE nama=VALUES(nama), email=VALUES(email), password_hash=VALUES(password_hash), role=VALUES(role);
