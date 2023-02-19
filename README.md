## Importador de dados Correios eDNE_Basico

#### Dependências
 Docker

#### Tempo estimado para processamento
 Foi verificado a duração de 01 (uma) hora para realizar todo o processo 
 usando o banco do container e o arquivo original dos correios.

#### Como rodar
1) Extraia o arquivo `eDNE_Basico` cedido pelos correios em algum diretório

2) Substitua os arquivos `.TXT` contidos na pasta `eDNE/basico` e `eDNE/delta`
com os respectivos arquivos extraídos dos correios.
> Obs: Nesse script usamos os arquivos delimitados

3) Crie um um arquivo `.env` com as credenciais conforme o arquivo `.env.example`
>`cp .env.example .env`

4) Suba os containers
>`docker-compose up`

#### Erros comuns
- Se a porta `5423` já estiver em uso altere a variável `POSTGRESQL_PORT` no arquivo `.env` para uma porta diferente.
