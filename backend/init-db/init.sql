CREATE TABLE merchants (
    id SERIAL PRIMARY KEY,
    qr_id TEXT NOT NULL, -- Tipe TEXT lambat untuk dicari, tidak ada UNIQUE
    merchant_name TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    merchant_id INT, -- Tidak ada Foreign Key yang mengikat ketat
    amount BIGINT NOT NULL,
    status TEXT DEFAULT 'PENDING',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO merchants (qr_id, merchant_name, created_at) VALUES 
('TEST001', 'Kantin FILKOM UB', NOW()),
('TEST002', 'TESTING STORE', NOW());