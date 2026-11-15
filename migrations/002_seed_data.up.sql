INSERT INTO roles (name) VALUES ('admin'), ('officer'), ('user')
ON CONFLICT (name) DO NOTHING;

INSERT INTO categories (name, description) VALUES
    ('Dompet', 'Dompet, tempat kartu, dan sejenisnya'),
    ('Tas', 'Tas, ransel, dan koper'),
    ('Elektronik', 'Ponsel, laptop, earphone, dan perangkat elektronik lain'),
    ('Dokumen', 'KTP, SIM, kartu mahasiswa, dan dokumen lain'),
    ('Kunci', 'Kunci kendaraan, rumah, atau loker'),
    ('Pakaian', 'Jaket, topi, dan aksesori pakaian'),
    ('Lainnya', 'Barang yang tidak masuk kategori lain')
ON CONFLICT (name) DO NOTHING;
