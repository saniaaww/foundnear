DELETE FROM categories WHERE name IN
    ('Dompet', 'Tas', 'Elektronik', 'Dokumen', 'Kunci', 'Pakaian', 'Lainnya');
DELETE FROM roles WHERE name IN ('admin', 'officer', 'user');