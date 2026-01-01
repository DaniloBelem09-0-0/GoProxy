import Redis from 'ioredis';

export class NexusLib {
  constructor(redisUrl) {
    this.redis = new Redis(redisUrl);
  }

  /**
   * Registra ou atualiza uma rota no service mesh.
   * @param {string} path 
   * @param {string[]} backends 
   */
  async registerService(path, backends) {
    const config = { path, backends };
    const payload = JSON.stringify(config);

    const pipeline = this.redis.pipeline();
    
    pipeline.set(`route:${path}`, payload);
    
    pipeline.publish('config_updates', payload);

    await pipeline.exec();
    console.log(`[Nexus Lib] Comando enviado: ${path} -> ${backends.length} backends`);
  }

  async listServices() {
    const keys = await this.redis.keys('route:*');
    if (keys.length === 0) return [];
    
    const values = await this.redis.mget(...keys);
    return values.map(v => JSON.parse(v));
  }
}