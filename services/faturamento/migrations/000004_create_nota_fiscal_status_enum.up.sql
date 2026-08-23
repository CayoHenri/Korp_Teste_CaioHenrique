CREATE TYPE faturamento.nota_fiscal_status AS ENUM (
    'ABERTA',
    'PROCESSANDO',
    'FECHADA'
);

ALTER TABLE faturamento.notas_fiscais
DROP CONSTRAINT IF EXISTS ck_notas_fiscais_status;

ALTER TABLE faturamento.notas_fiscais
ALTER COLUMN status TYPE faturamento.nota_fiscal_status
USING status::TEXT::faturamento.nota_fiscal_status;
