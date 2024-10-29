FROM node:18.12.1-alpine3.15 AS development

RUN mkdir -p /opt/importer/app && chown node:node /opt/importer/app

WORKDIR /opt/importer/app

COPY --chown=node:node . .

RUN npm install

USER node

CMD ["node", "main.mjs"]
