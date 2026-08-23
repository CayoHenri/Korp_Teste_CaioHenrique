UPDATE faturamento.outbox_events
SET event_type = 'faturamento.nota.fechamento_solicitado'
WHERE event_type = 'estoque.baixa.solicitada'
  AND published_at IS NULL;
