CREATE TABLE estoque.movimentacoes_estoque (
    id UUID PRIMARY KEY,
    produto_id UUID NOT NULL,
    tipo VARCHAR(20) NOT NULL,
    quantidade INTEGER NOT NULL,
    referencia UUID NULL,
    data_movimentacao TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_movimentacoes_produto
        FOREIGN KEY (produto_id) REFERENCES estoque.produtos (id),
    CONSTRAINT ck_movimentacoes_tipo
        CHECK (tipo IN ('ENTRADA', 'SAIDA')),
    CONSTRAINT ck_movimentacoes_quantidade_positiva
        CHECK (quantidade > 0)
);

CREATE INDEX idx_movimentacoes_produto_data
    ON estoque.movimentacoes_estoque (produto_id, data_movimentacao DESC);

CREATE INDEX idx_movimentacoes_referencia
    ON estoque.movimentacoes_estoque (referencia)
    WHERE referencia IS NOT NULL;
