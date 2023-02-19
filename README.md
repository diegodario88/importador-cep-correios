## Importador de dados Correios eDNE_Basico

#### Dependências
 Docker

#### Tempo estimado para processamento
 Foi verificado a duração de 01 (uma) hora para realizar todo o processo usando o banco do container

#### Como rodar
1) Extraia o arquivo `eDNE_Basico` cedido pelos correios na raiz do projeto
>`unzip /caminho/para/o/arquivo/caminho_do_arquivo.zip`
Certifique-se de substituir "caminho_do_arquivo.zip" pelo caminho real do arquivo que deseja extrair.

2) Crie um um arquivo `.env` com as credenciais conforme o arquivo `.env.example`
>`cp .env.example .env`

2) Suba os containers
>`docker-compose up`

#### Erros comuns
- Se a porta `5423` já estiver em uso altere a variável `POSTGRESQL_PORT` no arquivo `.env` para uma porta diferente.
