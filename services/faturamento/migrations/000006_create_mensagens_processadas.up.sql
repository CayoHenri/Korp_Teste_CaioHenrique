CREATE TABLE faturamento.mensagens_processadas (
    correlation_id UUID PRIMARY KEY,
    event_id UUID NOT NULL UNIQUE,
    processed_at TIMESTAMPTZ NOT NULL
);
