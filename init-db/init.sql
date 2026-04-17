-- TABEL MERCHANTS
-- Menyimpan data merchant/toko pemilik QRIS
CREATE TABLE merchants (
    id SERIAL PRIMARY KEY,
    qr_id TEXT NOT NULL, -- Tipe TEXT lambat untuk dicari, tidak ada UNIQUE
    merchant_name TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- TABEL TRANSACTIONS
-- Menyimpan riwayat transaksi pembayaran
CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    merchant_id INT, -- Tidak ada Foreign Key yang mengikat ketat
    amount BIGINT NOT NULL,
    status TEXT DEFAULT 'PENDING',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- DATA AWAL SIMULASI (DUMMY DATA)
-- Data ini akan otomatis terisi saat Docker Compose dijalankan
INSERT INTO merchants (qr_id, merchant_name) VALUES 
('000201010211265700', 'Kantin FILKOM UB'),
('000201010211265701', 'Toko Buku Mitra Niaga'),
('000201010211265702', 'Kopi Rural Sejahtera');