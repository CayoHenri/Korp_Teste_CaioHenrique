ALTER TABLE faturamento.notas_fiscais
ADD COLUMN nome_cliente VARCHAR(200) NOT NULL DEFAULT '',
ADD COLUMN endereco_cliente VARCHAR(500) NOT NULL DEFAULT '';

ALTER TABLE faturamento.itens_nota_fiscal
ADD COLUMN valor DOUBLE PRECISION NOT NULL DEFAULT 0,
ADD CONSTRAINT ck_itens_valor_unitario CHECK (valor >= 0);
