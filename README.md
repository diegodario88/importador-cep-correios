# Importador de CEPs dos Correios

## Descrição

Este projeto é uma aplicação CLI escrita em Go que importa arquivos do
Diretório Nacional de Endereços (DNE) dos Correios para um banco de dados
PostgreSQL. A aplicação foi construída com foco em desempenho, concorrência
e baixo consumo de memória.

## Visão geral

- Lê arquivos da base completa `eDNE/basico` no formato `.TXT`, com layout delimitado por `@`, conforme o padrão dos Correios.
- Processa os dados em paralelo, arquivo por arquivo.
- Utiliza `pgx.CopyFrom` para inserções em lote no PostgreSQL.
- Exibe barras de progresso em tempo real com a biblioteca `mpb`.
- Registra métricas como tempo total de execução, total de registros e total de CEPs inseridos e armazena em `correios.importacao_relatorio`
- Implementa uma função no banco de dados PostgreSQL para facilitar consultas por CEP, com interface simples e desempenho otimizado. Exemplo de uso:

  ```sql
  SELECT * FROM correios.consulta_cep('87020025');
  ```

  ```json
  {
    "uf": "PR",
    "localidade": "Maringá",
    "cep": "87020025",
    "ibge": "4115200",
    "bairro": "Zona 07",
    "complemento": "- de 701/702 ao fim",
    "logradouro": "Avenida Duque de Caxias"
  }
  ```

O propósito deste projeto é importar a base completa de CEPs para um banco PostgreSQL e, a partir disso, executar um `dump`
do schema `correios`, permitindo seu `restore` em ambientes de produção. Esse processo pode ser repetido periodicamente para manter
a sincronização com as atualizações quinzenais publicadas pelos Correios.

Optou-se por não utilizar os arquivos do diretório **Delta**, visto que a importação completa é suficientemente rápida e elimina
a complexidade de gerenciar operações de `UPDATE` e `DELETE`. Além disso, essa abordagem permite recuperar facilmente a sincronização
caso uma atualização quinzenal seja perdida, bastando importar novamente a base completa mais recente.

## Dependências

Para executar este projeto, você precisará de:

- Go (para rodar localmente, opcional)
- Docker
- Docker Compose
- Arquivos da base `eDNE/basico` (modelo atual disponível [aqui](https://www2.correios.com.br/sistemas/edne/default.cfm?s=true))

## Tempo estimado de processamento

Na versão 2.\*, a importação da base completa levou cerca de 25 segundos, com o hardware e software descritos abaixo.

- **Processador:** AMD Ryzen™ 9 3900 (24 threads)
- **Memória:** 64 GiB
- **Sistema operacional:** Fedora Linux 42 (Workstation Edition)
- **Kernel:** Linux 6.14.2-300.fc42.x86_64

## Como usar

1. Extraia os arquivos da base `eDNE_Basico` fornecida pelos Correios para um diretório local.

2. Substitua os arquivos `.TXT` existentes na pasta `eDNE/basico` pelo conteúdo extraído dos Correios.

   > Observação: Este projeto trabalha com arquivos delimitados por `@`.

3. Crie um arquivo `.env` com as credenciais do banco. Use o modelo `.env.example` como base:

   ```bash
   cp .env.example .env
   ```

````

4. Construa os containers

   ```bash
   docker compose build
   ```

5. Execute a aplicação
   ```bash
   docker compose run --rm importer
   ```

#### Erros comuns

- _Porta em uso:_ Se a porta `5432` já estiver ocupada no seu sistema, altere a variável `POSTGRESQL_PORT` no arquivo `.env`
  para uma porta diferente (ex: `6432`)

- _Barras de progresso não aparecem:_ Isso é esperado ao rodar via `docker compose logs`. As barras só são exibidas corretamente quando
  o terminal é interativo (ex: `go run`, `docker exec -it`, etc).

## Planos futuros

- Automatizar o processo de download e extração dos arquivos da base dos Correios utilizando a biblioteca `chromedp`, eliminando a etapa manual de obtenção dos dados.
- Criar uma tabela de relatório (`correios.importacao_relatorio`) contendo os dados consolidados da execução, como total de registros inseridos, total de CEPs distintos, data/hora de execução e versão da base importada.
- Gerar automaticamente o arquivo de `dump` do schema `correios` no formato binário (`.dump`) ao final da importação, facilitando restaurações e integrando com pipelines de produção.
- Adicionar uma etapa de confirmação interativa antes de iniciar o processo de importação, garantindo que o usuário esteja ciente das operações que serão executadas, especialmente em ambientes sensíveis.
- Exibir informações detalhadas sobre a conexão com o banco de dados no início da execução, incluindo a versão do PostgreSQL, para facilitar a validação de compatibilidade com ambientes de produção e evitar falhas em operações de restore.

<a href='https://ko-fi.com/Y8Y8Q12UV' target='_blank'><img height='36'
style='border:0px;height:36px;' src='https://cdn.ko-fi.com/cdn/kofi1.png?v=3'
border='0' alt='Buy Me a Coffee at ko-fi.com' /></a>
````
