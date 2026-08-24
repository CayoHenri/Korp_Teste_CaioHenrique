export type NotaFiscalStatus = 'ABERTA' | 'PROCESSANDO' | 'FECHADA';

export interface NotaFiscalResumo {
  id: string;
  numero: number;
  nomeCliente: string;
  quantidadeTotal: number;
  valorTotal: number;
  status: NotaFiscalStatus;
}
