ALTER TABLE faturamento.itens_nota_fiscal DROP CONSTRAINT IF EXISTS ck_itens_valor_unitario;
ALTER TABLE faturamento.itens_nota_fiscal DROP COLUMN IF EXISTS valor;
ALTER TABLE faturamento.notas_fiscais DROP COLUMN IF EXISTS endereco_cliente;
ALTER TABLE faturamento.notas_fiscais DROP COLUMN IF EXISTS nome_cliente;
