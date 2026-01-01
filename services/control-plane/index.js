import fastify from 'fastify';
import { NexusClient } from '@goproxy/client'; // Importação limpa

const app = fastify();
const nexus = new NexusClient({ 
  redisUrl: process.env.REDIS_URL || 'redis://localhost:6379' 
});

app.post('/services', async (request, reply) => {
  const { path, backends } = request.body;
  const result = await nexus.registerService(path, backends);
  return result;
});

app.get('/services', async () => {
  return await nexus.listServices();
});

app.listen({ port: 4000, host: '0.0.0.0' });