export interface Produto {
  id: string;
  codigo: string;
  descricao: string;
  saldo: number;
  valor: number;
  ativo: boolean;
  dataCadastro: string;
  dataAtualizacao: string;
}

export interface CriarProdutoInput {
  codigo: string;
  descricao: string;
  saldo: number;
  valor: number;
}

export interface AtualizarProdutoInput {
  descricao: string;
  saldo: number;
  valor: number;
}

export interface ProdutoFiltros {
  codigo: string;
  descricao: string;
  ativo: boolean | null;
  pagina: number;
  tamanhoPagina: number;
}

export interface ProdutosPaginados {
  itens: Produto[];
  total: number;
  pagina: number;
  tamanhoPagina: number;
  totalPaginas: number;
}

export type TipoMovimentacao = 'ENTRADA' | 'SAIDA';

export interface ProdutoMovimentacao {
  id: string;
  produtoId: string;
  tipo: TipoMovimentacao;
  quantidade: number;
  referencia?: string;
  dataMovimentacao: string;
}
