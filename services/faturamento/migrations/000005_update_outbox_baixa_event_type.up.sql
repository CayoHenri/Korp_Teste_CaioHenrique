UPDATE faturamento.outbox_events
SET event_type = 'estoque.baixa.solicitada'
WHERE event_type = 'faturamento.nota.fechamento_solicitado'
  AND published_at IS NULL;
