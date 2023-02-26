# Importador de dados Correios eDNE_Basico

## Descrição

Este repositório contém um script para importar dados do Correios eDNE_Basico para um banco de dados PostgreSQL usando Docker e Docker Compose. Este script usa arquivos delimitados para importação.

## Dependências

Para executar este script, você precisará das seguintes dependências:

- Docker
- Docker Compose

## Tempo estimado para processamento

Foi verificado que o processo completo leva cerca de 3 horas, usando o banco de dados no container e o arquivo original dos Correios.

## Como usar

Siga as instruções abaixo para executar o script:

1. Extraia o arquivo `eDNE_Basico` cedido pelos correios em algum diretório.

2. Substitua os arquivos `.TXT` contidos nas pastas `eDNE/basico` e `eDNE/delta` com os respectivos arquivos extraídos dos Correios.

   > Nota: Neste script, usamos arquivos delimitados.

3. Crie um arquivo `.env` com as credenciais conforme o arquivo `.env.example`.
   >`cp .env.example .env`

4) Suba os containers
   >`docker-compose up`

#### Erros comuns
- Se a porta `5423` já estiver em uso altere a variável `POSTGRESQL_PORT` no arquivo `.env` para uma porta diferente.
