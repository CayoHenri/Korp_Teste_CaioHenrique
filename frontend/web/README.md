# Frontend web

Aplicação Angular para operar os contextos de Estoque e Faturamento.

## Tecnologias

- Angular 22 com componentes standalone;
- TypeScript 6 em modo estrito;
- RxJS 7 para estado e fluxos assíncronos;
- Angular Material 22 e CDK para componentes;
- Material Icons para ícones;
- tema predefinido `azure-blue`;
- SCSS;
- Vitest pelo builder oficial do Angular.

As faixas de versão estão declaradas no `package.json` e as versões resolvidas
estão registradas no `package-lock.json`.

## Estado atual

Implementado nesta etapa:

- workspace Angular;
- configuração standalone;
- rotas lazy;
- shell responsivo com menu e rodapé;
- tema Angular Material e Material Icons;
- página inicial;
- página de Produtos integrada à API de Estoque;
- listagem de produtos com filtros e paginação;
- cadastro e edição de produtos;
- ativação e inativação sem exclusão;
- consulta do histórico de movimentações de cada produto em diálogo;
- feedback de carregamento, estado vazio, sucesso e erro;
- testes unitários do contrato HTTP de Produtos;
- página de Notas Fiscais integrada à API de Faturamento;
- listagem, filtros, paginação, criação e edição de notas abertas;
- impressão representada pelo fechamento assíncrono da nota;
- atualização automática de notas em processamento sem bloquear a tela;
- feedback visual de fechamento e rejeição;
- testes unitários do contrato HTTP de Faturamento;
- configuração central das URLs das APIs;
- store global e stores isolados por feature usando RxJS;
- configuração inicial de build e testes.

> A troca para Angular Material foi feita apenas no código e nas dependências.
> Build e testes não foram executados nesta etapa; valide-os após instalar os
> pacotes conforme os comandos abaixo.

Ainda não implementado:

- testes unitários dos stores e componentes visuais;
- container Docker do frontend.

## Pré-requisitos

Angular 22 requer uma versão compatível do Node. Este projeto foi validado com
Node `24.19.0` e npm `11.6.2`.

Consulte a matriz oficial antes de trocar versões:
<https://angular.dev/reference/versions>.

## Instalação e execução

PowerShell, Bash, Zsh, Git Bash ou Prompt de Comando:

```console
cd frontend/web
npm install
npm start
```

A aplicação fica em <http://localhost:4200>.

Build de produção:

```console
npm run build
```

Testes:

```console
npm test -- --watch=false
```

## Estrutura

```text
src/app/
├── core/
│   ├── config/                  tokens e URLs das APIs
│   ├── http/                    contratos e tratamento comum de erros HTTP
│   └── state/                   estado global da aplicação
├── features/
│   ├── inicio/                  visão geral
│   ├── produtos/
│   │   ├── filters/             formulário de filtros da listagem
│   │   ├── form/                formulário de criação e edição
│   │   ├── list/                tabela e ações da listagem
│   │   ├── movements/           histórico de entradas e saídas
│   │   └── *.ts                 página, model, store e cliente HTTP
│   └── notas-fiscais/
│       ├── filters/             filtros por número, cliente e status
│       ├── form/                cliente, endereço e itens da nota
│       ├── list/                tabela, edição e impressão
│       └── *.ts                 página, model, store e cliente HTTP
├── layout/                      shell e navegação
├── shared/
│   └── ui/
│       ├── confirmation-dialog/ diálogo reutilizável de confirmação
│       ├── data-feedback/       estados vazio e de erro
│       ├── page-header/         cabeçalho das páginas
│       └── pagination/          paginação independente da feature
├── app.config.ts                providers globais
└── app.routes.ts                rotas lazy
```

Cada feature deve conter apenas o que pertence ao próprio contexto. Elementos
reutilizados por mais de uma feature podem ser promovidos para `shared`.

## Estado com RxJS

RxJS é o padrão do projeto para estado e operações assíncronas. Não será usado
NgRx nesta fase inicial.

Cada store segue estas regras:

1. mantém o `BehaviorSubject` privado;
2. expõe estado por `Observable` somente leitura;
3. oferece seletores com `map` e `distinctUntilChanged`;
4. altera estado por métodos explícitos;
5. não permite que componentes chamem `next`;
6. mantém loading e erro no estado da feature;
7. templates preferem `AsyncPipe` a `subscribe` manual.

Exemplo reduzido:

```typescript
private readonly stateSubject = new BehaviorSubject<State>(initialState);

readonly state$ = this.stateSubject.asObservable();
readonly itens$ = this.state$.pipe(
  map((state) => state.itens),
  distinctUntilChanged(),
);
```

O serviço HTTP faz o transporte e converte o envelope da API. O store orquestra
estado, carregamento, filtros, paginação e mutações. Componentes ficam
responsáveis por eventos e exibição.

Erros de carregamento pertencem ao estado da listagem e aparecem no card de
feedback. Erros de mutações, como cadastrar, editar ou alterar status, retornam
ao componente e aparecem em snackbar; eles não substituem a tabela por uma
mensagem de falha de carregamento. Resultados cancelados de diálogos devem ser
validados antes de iniciar qualquer chamada HTTP.

## Angular Material

Angular Material usa o tema `azure-blue` configurado no `angular.json`.
Componentes são importados por módulos específicos no array `imports` de cada
componente standalone, evitando um módulo global com toda a biblioteca.

Customizações devem usar primeiro o sistema de temas e as variáveis próprias da
aplicação. `::ng-deep` deve ser evitado porque cria acoplamento com detalhes
internos dos componentes.

O layout adota densidade compacta para priorizar os dados operacionais: toolbar
de 48px, contexto, título e descrição na mesma linha, filtros de 36px, linhas de
tabela de 34px e paginador de 40px. Estados de hover, bordas e contraste mantêm
a leitura clara com menos espaço. Em telas estreitas, os componentes voltam a
empilhar os controles para preservar legibilidade e área de toque.

Tabelas usam separadores suaves, colunas numéricas alinhadas, códigos em fonte
monoespaçada com truncamento e tooltip, além de ações compactas com descrição.
Formulários em diálogo definem a largura pelo `MatDialog`, nunca pelo conteúdo
interno, evitando rolagem horizontal e mantendo cabeçalho e ações consistentes.

`PortuguesePaginatorIntl` traduz rótulos, navegação e intervalos de todas as
paginações. Confirmações de ações não usam `window.confirm`: devem abrir
`ConfirmationDialog`, mantendo aparência, acessibilidade e textos consistentes
com o restante da aplicação.

Documentação oficial: <https://material.angular.dev/>.

## Rotas

| Rota             | Feature     |
| ---------------- | ----------- |
| `/inicio`        | Visão geral |
| `/produtos`      | Estoque     |
| `/notas-fiscais` | Faturamento |

A raiz redireciona para `/inicio`. Rotas desconhecidas também retornam ao início.

## Configuração das APIs

As URLs iniciais ficam em `src/environments/environment.ts` e são expostas por
`API_CONFIG`, um `InjectionToken` tipado:

```text
Estoque:     http://localhost:8081
Faturamento: http://localhost:8082
```

Serviços HTTP devem injetar esse token em vez de repetir URLs literais.

O frontend chama as APIs diretamente. Estoque e Faturamento liberam a origem do
Angular por CORS usando, respectivamente, `ESTOQUE_CORS_ALLOWED_ORIGINS` e
`FATURAMENTO_CORS_ALLOWED_ORIGINS`. Não existe proxy no servidor de
desenvolvimento do frontend.

## Fluxo da feature Produtos

```text
ProdutosPage → ProdutosStore → ProdutoHttpService → API de Estoque
```

- `produto.model.ts` descreve respostas, filtros e comandos;
- `produto-http.service.ts` conhece endpoints e envelopes HTTP;
- `produtos.store.ts` mantém o estado RxJS e recarrega a listagem após mutações;
- `form/produto-form-dialog.ts` reutiliza o formulário para criação e edição;
- `filters/produtos-filters.ts` mantém e normaliza os controles de busca;
- `list/produtos-list.ts` apresenta a tabela e emite ações para o container;
- `movements/produto-movimentacoes-dialog.ts` consulta e apresenta o histórico;
- `produtos-page.ts` trata somente eventos e apresentação.

Features com mais de uma responsabilidade visual devem ser divididas em pastas
como `form/`, `list/` e `details/`. Componentes sem conhecimento do domínio,
como paginação e feedback de dados, ficam em `shared/ui` e recebem toda a
configuração por inputs, comunicando eventos por outputs.

## Fluxo da feature Notas Fiscais

```text
NotasFiscaisPage → NotasFiscaisStore → NotaFiscalHttpService → API de Faturamento
```

O botão de impressão não gera documento nesta etapa. Ele confirma a operação e
chama `POST /notas-fiscais/{id}/fechamento`. A resposta muda a nota para
`PROCESSANDO` e o store inicia um polling RxJS a cada 1,5 segundo:

1. apenas notas em processamento são monitoradas;
2. cada ID possui no máximo um fluxo de acompanhamento;
3. `exhaustMap` impede requisições sobrepostas quando a API demora;
4. somente a linha consultada é atualizada, preservando filtros e paginação;
5. o fluxo termina em `FECHADA` ou quando a rejeição devolve a nota para
   `ABERTA` com `motivoRejeicao`;
6. `takeUntil` cancela todos os acompanhamentos ao sair da página.

O botão de impressão permanece visível em todos os estados, mas só é habilitado
para notas `ABERTA`. Durante o fechamento, spinner e status `Processando`
fornecem feedback sem bloquear os demais controles da tela.

## Convenções

- nomes de arquivos em kebab-case;
- componentes standalone e rotas lazy;
- componentes de UI com arquivos `.ts`, `.html` e `.scss` separados;
- models como interfaces ou tipos imutáveis;
- estado localizado na feature que o utiliza;
- imports RxJS diretamente de `rxjs`;
- nenhuma regra de negócio duplicada do backend;
- módulos Angular Material importados sob demanda;
- textos e mensagens de usuário em português;
- testes próximos aos arquivos testados.

## Próximas etapas

1. criar testes dos stores e dos componentes visuais;
2. adicionar o frontend ao Docker Compose.
