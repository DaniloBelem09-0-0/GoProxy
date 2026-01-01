import Redis from 'ioredis';

/**
 * NexusClient - Biblioteca core para gestão do Nexus-Mesh.
 * Responsável por persistência e sinalização via Redis.
 */
export class NexusClient {
  constructor(options = {}) {
    // Configurações padrão para o ambiente Docker ou Local
    this.redisUrl = options.redisUrl || 'redis://localhost:6379';
    this.redis = new Redis(this.redisUrl);
    this.channel = 'config_updates';
    
    this._setupErrorHandling();
  }

  /**
   * Configura listeners básicos de erro para evitar que a lib derrube a aplicação.
   */
  _setupErrorHandling() {
    this.redis.on('error', (err) => {
      console.error('[Nexus-Client] Erro na conexão com Redis:', err.message);
    });
  }

  /**
   * Registra uma rota no Data Plane (Go).
   * @param {string} path - O endpoint (ex: '/api/v1')
   * @param {string[]} backends - Lista de servidores (ex: ['http://localhost:3001'])
   */
  async registerService(path, backends) {
    if (!path || !Array.isArray(backends) || backends.length === 0) {
      throw new Error('Parâmetros inválidos: path e backends (array) são obrigatórios.');
    }

    const payload = JSON.stringify({ path, backends });

    // Padrão Unit of Work: Garante que o SET e o PUBLISH ocorram juntos
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

  /**
   * Recupera todos os serviços ativos registrados no Redis.
   */
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

  /**
   * Encerra a conexão de forma segura.
   */
  async disconnect() {
    return this.redis.quit();
  }
}