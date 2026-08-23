ALTER TABLE estoque.produtos
ADD COLUMN valor DOUBLE PRECISION NOT NULL DEFAULT 0,
ADD CONSTRAINT ck_produtos_valor_unitario CHECK (valor > 0);
