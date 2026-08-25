DELETE FROM estoque.movimentacoes_estoque
WHERE produto_id::TEXT LIKE '10000000-0000-4000-8000-%';

DELETE FROM estoque.produtos
WHERE id::TEXT LIKE '10000000-0000-4000-8000-%';
