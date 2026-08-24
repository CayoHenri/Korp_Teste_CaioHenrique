export type NotaFiscalStatus = 'ABERTA' | 'PROCESSANDO' | 'FECHADA';

export interface ItemNotaFiscal {
  id: string;
  produtoId: string;
  codigoProduto: string;
  descricaoProduto: string;
  quantidade: number;
  valor: number;
  valorTotal: number;
}

export interface NotaFiscal {
  id: string;
  numero: number;
  status: NotaFiscalStatus;
  nomeCliente: string;
  enderecoCliente: string;
  motivoRejeicao?: string;
  quantidadeTotal: number;
  valorTotal: number;
  itens: ItemNotaFiscal[];
  dataCadastro: string;
  dataAtualizacao: string;
  dataFechamento?: string;
}

export interface NotaFiscalItemInput {
  codigoProduto: string;
  quantidade: number;
}

export interface NotaFiscalInput {
  nomeCliente: string;
  enderecoCliente: string;
  itens: NotaFiscalItemInput[];
}

export interface NotaFiscalFiltros {
  numero: number | null;
  status: NotaFiscalStatus | null;
  nomeCliente: string;
  pagina: number;
  tamanhoPagina: number;
}

export interface NotasFiscaisPaginadas {
  itens: NotaFiscal[];
  total: number;
  pagina: number;
  tamanhoPagina: number;
  totalPaginas: number;
}

export interface ResultadoFechamento {
  nota: NotaFiscal;
  sucesso: boolean;
}
