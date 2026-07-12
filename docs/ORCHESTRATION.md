# Gymkhana Core — Documento de Orquestração

> **Planning version:** Stage 4  
> **Última sincronização:** 2026-07-12  
> **Etapa atual:** Etapa 4 concluída  
> **Fonte principal de verdade:** `Pherlsz/Gymkhana-Database/docs/ORCHESTRATION.md`  
> **Responsabilidade deste repositório:** lógica Go reutilizável e independente de infraestrutura

Este documento registra tudo que afeta o `Gymkhana-Core`. Em caso de divergência sobre o produto ou modelo de dados, o documento do `Gymkhana-Database` prevalece.

## 1. Missão do repositório

Fornecer um módulo Go reutilizável para lógicas que não dependem do produto Gymkhana Database, do PostgreSQL, de HTTP, de provedores de IA ou de interface visual.

Responsabilidades esperadas:

- normalização de dados;
- representação tipada de filtros e operadores;
- AST do Query Engine;
- matching genérico;
- pontuação e explicações de duplicidade;
- contratos neutros de IA;
- orquestração neutra de ferramentas;
- formatos de resultados e referências;
- utilidades puras de datas civis, competência e valores canônicos quando realmente reutilizáveis.

O Core deve ser pequeno, testável, previsível e sem dependências de infraestrutura.

## 2. O que não pertence ao Core

Não incluir:

- repositories PostgreSQL;
- queries SQL;
- `pgx`, sqlc ou migrations;
- handlers HTTP;
- OpenAPI gerado do produto;
- autenticação e sessão;
- River ou qualquer fila;
- R2, S3 ou storage;
- Google Forms;
- SDK da OpenAI;
- SDK do Google GenAI;
- Vercel, Cloud Run, Neon ou Cloudflare;
- React ou TypeScript;
- componentes visuais;
- regras específicas de página, formulário ou Data Grid;
- entidades completas do produto sem valor real de reutilização;
- permissões de MEMBER, ADMIN e SUPERADMIN;
- tabelas e modelos de persistência do Gymkhana Database.

## 3. Princípios aprovados

- domínio e infraestrutura permanecem separados;
- interfaces somente nos limites em que há substituição real;
- nenhuma abstração antecipada sem caso de uso;
- nenhuma dependência instalada para economizar poucas linhas;
- nenhum tipo de package externo atravessa a API pública do Core sem necessidade;
- nenhuma versão atualizada apenas por ser mais nova;
- nenhum release com vulnerabilidade conhecida e aplicável;
- funções puras quando possível;
- resultados determinísticos para normalização e matching;
- explicações estruturadas em vez de porcentagens arbitrárias;
- IA nunca produz SQL arbitrário;
- o Core descreve intenção e plano, mas não executa acesso a dados.

## 4. Stack aprovada

- Go 1.26;
- módulos Go;
- biblioteca padrão como primeira opção;
- `testing`;
- fuzzing nativo;
- race detector quando houver concorrência;
- `gofmt`;
- `go vet`;
- `staticcheck`;
- `govulncheck`;
- OSV-Scanner;
- Dependabot Alerts;
- Dependency Review quando disponível.

Não usar framework de aplicação, ORM, framework de agentes ou biblioteca de assertions inicialmente.

O patch exato será fixado na inicialização do repositório após revisão de compatibilidade, segurança e changelog.

## 5. Organização conceitual

Estrutura indicativa, a ser finalizada na Etapa 5:

```text
pkg/
├── normalize/
├── query/
├── matching/
├── duplicates/
├── assistant/
├── tools/
├── result/
└── civiltime/

internal/
└── testutil/
```

A estrutura final pode usar `pkg` ou packages na raiz. A decisão física será tomada na Etapa 5.

## 6. Normalização

Normalizações reutilizáveis previstas:

- nomes;
- texto para Search;
- espaços;
- caixa;
- acentos;
- CPF;
- telefones;
- e-mail;
- números de documentos;
- identificadores alfanuméricos;
- valores de opções;
- datas civis;
- competência mensal;
- endereços textuais quando aplicável.

Regras importantes:

- preservar zeros à esquerda;
- números de documentos permanecem texto;
- letras canônicas em maiúsculas quando aplicável;
- separadores visuais não fazem parte do valor canônico;
- nomes devem preservar o valor de exibição fora da função de normalização de Search;
- e-mail em minúsculas;
- data civil não é um instante UTC;
- `YEAR_MONTH` é diferente de `DATE` e `DATETIME`.

O Core não decide como esses valores são persistidos. Ele apenas oferece tipos e funções neutras.

## 7. Tipos temporais neutros

Separação obrigatória:

- `CivilDate`: ano, mês e dia sem timezone;
- `YearMonth`: ano e mês;
- `Instant`: instante real tratado pelo produto com `time.Time` e UTC.

O Core pode fornecer:

- parsing estrito;
- validação;
- comparação;
- avanço ou recuo de mês;
- ano bissexto;
- serialização canônica;
- formatação neutra quando não depender de UI.

Não deve conter regras visuais específicas de `pt-BR`; a camada de apresentação decide a exibição final.

## 8. Query Engine

O Core define a linguagem estruturada de consulta, não o SQL.

Modelo conceitual:

```go
type QueryPlan struct {
    Entity   EntityRef
    Filters  []Filter
    Includes []RelationRef
    Sort     []Sort
    Limit    int
}
```

Elementos previstos:

- entidade alvo;
- campo;
- operador;
- valor tipado;
- composição lógica;
- relações incluídas;
- ordenação;
- limite;
- projeção de campos;
- agrupamento quando necessário;
- metadados de origem da intenção.

Operadores genéricos previstos:

- equal;
- not equal;
- contains;
- starts with;
- ends with;
- in;
- not in;
- greater than;
- greater or equal;
- less than;
- less or equal;
- between;
- is empty;
- is not empty;
- matches normalized value;
- approximate match quando o executor suportar.

O Core deve validar:

- operador compatível com o tipo do campo;
- valores obrigatórios;
- limites válidos;
- composição lógica válida;
- ausência de campos ou operadores desconhecidos;
- profundidade e complexidade máximas configuráveis.

O Core não conhece nomes de colunas PostgreSQL, schemas, índices ou joins.

## 9. Catálogo de campos

O Query Engine recebe do produto um catálogo neutro de campos disponíveis.

Exemplo conceitual:

```go
type FieldDefinition struct {
    Key        string
    Label      string
    Type       ValueType
    Operators  []Operator
    Searchable bool
    Sortable   bool
    Filterable bool
}
```

O produto monta o catálogo a partir de:

- campos nativos;
- tipos de documentos;
- tipos de contas;
- entidades customizadas;
- campos customizados;
- permissões do usuário.

O Core valida o plano contra esse catálogo, mas não busca o catálogo no banco.

## 10. Resultados e referências

Formato neutro de resultado de Search ou consulta:

```go
type ResultReference struct {
    EntityType string
    EntityID   string
    Label      string
    FieldKey   string
}

type MatchResult struct {
    Reference   ResultReference
    MatchedValue string
    Highlight    string
    Relevance    float64
}
```

O produto pode acrescentar URLs e permissões depois.

O Core não cria rotas web, links absolutos ou referências a páginas específicas.

## 11. Matching e duplicatas

O Core oferece regras genéricas de comparação e explicações estruturadas.

Níveis aprovados no produto:

- `VERY_STRONG`;
- `PROBABLE`;
- `POSSIBLE`.

O Core não deve retornar uma porcentagem apresentada como certeza.

A saída deve conter razões:

```go
type DuplicateReason struct {
    RuleKey   string
    FieldKey  string
    Strength  Strength
    MessageKey string
}
```

Exemplos de sinais reutilizáveis:

- CPF idêntico;
- documento idêntico;
- telefone idêntico;
- e-mail idêntico;
- nome normalizado semelhante;
- data de nascimento igual;
- filiação semelhante;
- endereço semelhante;
- combinação de sinais fracos.

Nome nunca pode ser o único critério para uma sugestão forte.

O produto decide:

- quais campos existem;
- quais pesos ou regras estão ativos;
- quando criar revisão;
- como persistir candidatos;
- como executar merge;
- como aplicar permissões.

O Core calcula e explica.

## 12. Regras de versão das duplicatas

As configurações de matching podem ser versionadas.

A avaliação deve aceitar:

- versão das regras;
- versão dos dados comparados;
- configuração de normalização;
- limiares por nível.

Isso permite ao produto decidir se uma revisão descartada deve reaparecer quando dados ou regras mudarem.

O Core não persiste versões nem revisões.

## 13. Contratos neutros de IA

O Core não depende de SDK de provedor.

Contratos conceituais:

```go
type GenerateRequest struct {
    Messages []Message
    Tools    []ToolDefinition
    Schema   *OutputSchema
}

type GenerateResult struct {
    Text       string
    ToolCalls  []ToolCall
    Structured any
    Usage      Usage
}
```

O produto implementa adapters para OpenAI, Google ou outros provedores.

Nenhum tipo de `openai-go` ou `go-genai` pode aparecer na API pública do Core.

## 14. Orquestração neutra do AI Chat

Fluxo aprovado:

1. receber pergunta;
2. classificar ou interpretar intenção;
3. construir plano tipado;
4. validar plano contra catálogo e política;
5. solicitar execução a uma tool neutra;
6. receber resultado estruturado;
7. sintetizar resposta;
8. produzir referências.

O Core pode fornecer uma máquina de estados simples e interfaces neutras.

Não usar:

- LangChain;
- LangGraph;
- CrewAI;
- Semantic Kernel;
- framework genérico de agentes;
- SQL arbitrário;
- credenciais de banco;
- acesso direto a repositories.

Exemplo de tool neutra:

```go
type QueryExecutor interface {
    Execute(ctx context.Context, plan QueryPlan) (QueryResult, error)
}
```

O `Gymkhana-Database` implementa essa interface usando seus módulos e SQL seguro.

## 15. Tool calling

Tools devem ser pequenas, explícitas e tipadas.

Exemplos futuros:

- consultar Profiles;
- consultar documentos;
- consultar contas;
- consultar entidades customizadas;
- obter detalhes de um registro;
- gerar visão tabular navegável.

Não criar uma tool genérica `execute_sql`.

Cada tool deve declarar:

- nome estável;
- descrição;
- schema de entrada;
- schema de saída;
- permissões exigidas como metadado neutro;
- limites de resultados;
- erros técnicos estáveis.

O produto é responsável por autenticação e autorização reais.

## 16. Streaming

O produto usará SSE sobre HTTP, mas o Core não conhece SSE.

O Core pode expor eventos neutros:

- `message.started`;
- `text.delta`;
- `tool.started`;
- `tool.completed`;
- `references`;
- `message.completed`;
- `message.failed`;
- `message.cancelled`.

A camada HTTP do `Gymkhana-Database` converte esses eventos para SSE.

## 17. OCR

O Core pode oferecer:

- schema neutro de extração;
- validação de resultado estruturado;
- comparação entre valor atual e sugerido;
- normalização dos valores extraídos;
- classificação de campos aceitos, rejeitados ou conflitantes.

O Core não:

- baixa arquivos;
- acessa R2;
- chama OpenAI ou Google;
- persiste resultados;
- altera entidades;
- decide autorização.

Revisão humana obrigatória permanece regra do produto.

## 18. Custom fields

O Core pode conhecer tipos genéricos de valor:

- text;
- long text;
- number;
- money;
- civil date;
- datetime;
- boolean;
- single select;
- multi select;
- email;
- phone;
- URL;
- attachment reference.

Pode validar compatibilidade entre operador e tipo.

Não conhece:

- IDs de banco;
- tabelas `custom_fields`;
- owners;
- storage;
- widgets React;
- regras de import específicas do produto.

## 19. Imports

O Core pode fornecer funções puras para:

- normalizar cabeçalhos;
- converter valores textuais em tipos;
- validar uma linha contra um schema neutro;
- representar erros por campo;
- gerar candidatos de duplicidade a partir de dados normalizados.

Não inclui:

- Excelize;
- R2;
- `pgx.CopyFrom`;
- staging no PostgreSQL;
- jobs;
- progresso;
- Google Forms.

Essas responsabilidades permanecem no `Gymkhana-Database`.

## 20. Erros

O Core usa erros técnicos estáveis e neutros.

Exemplo:

```go
type FieldError struct {
    Field string
    Code  string
    Args  map[string]any
}
```

Mensagens finais em português podem ser montadas pela aplicação.

O Core não retorna envelopes HTTP nem status codes diretamente.

## 21. Testes

### 21.1 Unitários

Cobertura prioritária:

- normalização Unicode;
- acentos e espaços;
- CPF e documentos;
- telefones;
- e-mail;
- datas civis;
- competência;
- operadores;
- validação de AST;
- matching;
- níveis de duplicidade;
- tool schemas;
- transições da orquestração.

### 21.2 Fuzzing

Fuzzing nativo para:

- Unicode;
- documentos com caracteres arbitrários;
- telefones;
- datas;
- parser de query;
- normalizadores;
- valores de filtro;
- schemas de tools.

### 21.3 Determinismo

Mesma entrada, configuração e versão das regras devem produzir a mesma saída.

Testes devem evitar dependência de relógio global, locale da máquina ou ordem de map.

## 22. Dependências e segurança

Antes de adicionar uma package:

- comprovar necessidade;
- revisar licença;
- revisar manutenção;
- revisar dependências transitivas;
- verificar vulnerabilidades;
- avaliar se a biblioteca padrão resolve;
- impedir vazamento do tipo externo pela API pública.

Pipeline:

```bash
gofmt

go vet ./...
staticcheck ./...
go test ./...
go test -race ./...
govulncheck ./...
```

Supply chain:

- versões fixadas;
- `go.sum` versionado;
- OSV-Scanner;
- Dependabot Alerts;
- Dependency Review quando disponível;
- Actions por commit SHA;
- nenhum merge automático de atualização.

## 23. Relação com Gymkhana-Database

O produto consome uma versão fixa do Core.

O `Gymkhana-Database` fornece implementações de:

- catálogo de campos;
- QueryExecutor;
- repositories;
- autorização;
- adapters de IA;
- persistência de threads e runs;
- storage;
- HTTP e SSE;
- jobs;
- configuração.

O Core fornece modelos e regras neutras.

## 24. Relação com Gymkhana-UI

Não há dependência direta.

Core é Go e UI é React/TypeScript.

Não criar package multi-linguagem para “compartilhar tudo”.

A comunicação entre backend e frontend ocorre por OpenAPI no repositório do produto.

Conceitos equivalentes podem existir nas duas linguagens, mas são contratos separados e gerados quando apropriado.

## 25. Versionamento

A estratégia física será definida na Etapa 5.

Regras já aprovadas:

- tags ou versões exatas;
- nunca depender de branch flutuante;
- breaking changes explícitas;
- changelog para mudanças públicas;
- API pública pequena;
- packages internos por padrão;
- exportar somente o que tiver consumidor real.

## 26. Decisões adiadas

- package layout definitivo;
- política exata de semantic versioning;
- publicação pública ou privada;
- catálogo completo de operadores;
- parser de linguagem natural próprio;
- regras finais de pontuação de duplicidade;
- schemas finais de tools;
- suporte a provedores além dos adapters do produto;
- otimizações baseadas em benchmark.

## 27. Próxima etapa

Na Etapa 5 serão definidos:

- estrutura física definitiva;
- nome do módulo Go;
- packages públicos e internos;
- estratégia de tags e releases;
- CI;
- consumo pelo `Gymkhana-Database`;
- política de compatibilidade;
- ownership e revisão de mudanças públicas.

Este documento deve ser atualizado novamente ao final da Etapa 5.
