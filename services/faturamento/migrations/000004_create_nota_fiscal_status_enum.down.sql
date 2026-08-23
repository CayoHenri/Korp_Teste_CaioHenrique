ALTER TABLE faturamento.notas_fiscais
ALTER COLUMN status TYPE VARCHAR(20)
USING status::TEXT;

ALTER TABLE faturamento.notas_fiscais
ADD CONSTRAINT ck_notas_fiscais_status
CHECK (status IN ('ABERTA', 'PROCESSANDO', 'FECHADA'));

DROP TYPE faturamento.nota_fiscal_status;
