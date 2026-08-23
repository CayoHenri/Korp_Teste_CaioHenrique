ALTER TABLE estoque.produtos DROP CONSTRAINT IF EXISTS ck_produtos_valor_unitario;
ALTER TABLE estoque.produtos DROP COLUMN IF EXISTS valor;
