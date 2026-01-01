import Redis from 'ioredis';

export class NexusClient {
  constructor(options = {}) {
    this.redisUrl = options.redisUrl || 'redis://localhost:6379';
    this.redis = new Redis(this.redisUrl);
    this.channel = 'config_updates';
    
    this._setupErrorHandling();
  }

  _setupErrorHandling() {
    this.redis.on('error', (err) => {
      console.error('[Nexus-Client] Erro na conexão com Redis:', err.message);
    });
  }

  async registerService(path, backends) {
    if (!path || !Array.isArray(backends) || backends.length === 0) {
      throw new Error('Parâmetros inválidos: path e backends (array) são obrigatórios.');
    }

    const payload = JSON.stringify({ path, backends });

    const pipeline = this.redis.pipeline();
    pipeline.set(`route:${path}`, payload);
    pipeline.publish(this.channel, payload);

    try {
      await pipeline.exec();
      return { status: 'success', data: { path, backends } };
    } catch (error) {
      throw new Error(`Falha ao registrar serviço no Redis: ${error.message}`);
    }
  }

  async listServices() {
    try {
      const keys = await this.redis.keys('route:*');
      if (keys.length === 0) return [];

      const values = await this.redis.mget(...keys);
      return values.map((v) => JSON.parse(v));
    } catch (error) {
      throw new Error(`Erro ao listar serviços: ${error.message}`);
    }
  }

  async disconnect() {
    return this.redis.quit();
  }
}
