DELETE FROM faturamento.notas_fiscais
WHERE id::TEXT LIKE '30000000-0000-4000-8000-%';

SELECT SETVAL(
    'faturamento.nota_fiscal_numero_seq',
    COALESCE((SELECT MAX(numero) FROM faturamento.notas_fiscais), 1),
    EXISTS (SELECT 1 FROM faturamento.notas_fiscais)
);
