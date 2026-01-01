FROM node:20-alpine

WORKDIR /app

# 1. Copia os manifestos (Note que agora é 'goproxy-client')
COPY package*.json ./
COPY packages/goproxy-client/package*.json ./packages/goproxy-client/
COPY services/control-plane/package*.json ./services/control-plane/

# 2. Instala dependências
RUN npm install

# 3. Copia o código fonte (Aqui também precisa mudar!)
COPY packages/goproxy-client ./packages/goproxy-client
COPY services/control-plane ./services/control-plane

# 4. Define o diretório de trabalho final
WORKDIR /app/services/control-plane

EXPOSE 4000

CMD ["node", "index.js"]