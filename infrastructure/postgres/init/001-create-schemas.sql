-- Este script roda somente na primeira inicializacao do volume do PostgreSQL.
-- Cada microsservico sera responsavel exclusivamente por seu proprio schema.

CREATE SCHEMA IF NOT EXISTS estoque;
CREATE SCHEMA IF NOT EXISTS faturamento;

COMMENT ON SCHEMA estoque IS
    'Dados pertencentes exclusivamente ao microsservico de Estoque';

COMMENT ON SCHEMA faturamento IS
    'Dados pertencentes exclusivamente ao microsservico de Faturamento';
