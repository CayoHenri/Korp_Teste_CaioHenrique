CREATE TABLE IF NOT EXISTS estoque.produtos (
    id UUID PRIMARY KEY,
    codigo VARCHAR(50) NOT NULL,
    descricao VARCHAR(255) NOT NULL,
    saldo INTEGER NOT NULL,
    data_cadastro TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    data_atualizacao TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_produtos_codigo UNIQUE (codigo),
    CONSTRAINT ck_produtos_codigo_nao_vazio CHECK (BTRIM(codigo) <> ''),
    CONSTRAINT ck_produtos_descricao_nao_vazia CHECK (BTRIM(descricao) <> ''),
    CONSTRAINT ck_produtos_saldo_nao_negativo CHECK (saldo >= 0)
);

CREATE INDEX IF NOT EXISTS idx_produtos_descricao
    ON estoque.produtos (descricao);
