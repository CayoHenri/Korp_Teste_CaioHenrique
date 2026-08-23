CREATE SEQUENCE faturamento.nota_fiscal_numero_seq START WITH 1 INCREMENT BY 1;

CREATE TABLE faturamento.notas_fiscais (
    id UUID PRIMARY KEY,
    numero BIGINT NOT NULL UNIQUE,
    status VARCHAR(20) NOT NULL,
    data_cadastro TIMESTAMPTZ NOT NULL,
    data_atualizacao TIMESTAMPTZ NOT NULL,
    data_fechamento TIMESTAMPTZ NULL,
    CONSTRAINT ck_notas_fiscais_status CHECK (status IN ('ABERTA', 'PROCESSANDO', 'FECHADA'))
);

CREATE TABLE faturamento.itens_nota_fiscal (
    id UUID PRIMARY KEY,
    nota_fiscal_id UUID NOT NULL REFERENCES faturamento.notas_fiscais(id) ON DELETE CASCADE,
    produto_id UUID NOT NULL,
    codigo_produto VARCHAR(100) NOT NULL,
    descricao_produto VARCHAR(255) NOT NULL,
    quantidade INTEGER NOT NULL,
    CONSTRAINT ck_itens_nota_fiscal_quantidade CHECK (quantidade > 0)
);

CREATE INDEX idx_itens_nota_fiscal_nota_id ON faturamento.itens_nota_fiscal(nota_fiscal_id);
