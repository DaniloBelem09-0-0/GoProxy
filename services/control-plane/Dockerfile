FROM node:20-alpine

WORKDIR /app

COPY package*.json ./
COPY packages/goproxy-client/package*.json ./packages/goproxy-client/
COPY services/control-plane/package*.json ./services/control-plane/

RUN npm install

COPY packages/goproxy-client ./packages/goproxy-client
COPY services/control-plane ./services/control-plane

WORKDIR /app/services/control-plane

EXPOSE 4000

CMD ["node", "index.js"]