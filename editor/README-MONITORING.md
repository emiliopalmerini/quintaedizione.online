# Sistema di Monitoraggio e Performance

Documentazione per il sistema avanzato di monitoraggio, caching e performance ottimizzata implementato nell'editor D&D 5e SRD.

## 🎯 Caratteristiche Implementate

### 1. Connection Manager MongoDB Avanzato
- **Pool di connessioni configurabile** (min: 10, max: 50)
- **Monitoring automatico** comandi MongoDB con metriche
- **Gestione timeout** e retry automatici
- **Health checks continui** ogni 30 secondi
- **Creazione indici ottimizzata** in background

```python
# Configurazione connection manager
ConnectionConfig(
    max_pool_size=50,
    min_pool_size=10,
    server_selection_timeout_ms=5000,
    retry_writes=True
)
```

### 2. Sistema Caching Redis + Fallback
- **Cache Redis** con fallback in-memory automatico
- **TTL configurabile** per diversi tipi di dati
- **Invalidazione intelligente** per pattern di chiavi
- **Metriche cache** (hit rate, response time, errori)
- **Health monitoring** Redis con failover trasparente

```python
# Cache automatica con decorator
@cached(ttl=300)
async def expensive_operation(param):
    return await heavy_computation(param)
```

### 3. Query Ottimizzate e Batch Operations
- **Aggregation pipeline** per paginazione efficiente
- **Batch counting** collezioni in parallelo
- **Search ottimizzata** con relevance scoring
- **Navigation queries** per prev/next documenti
- **Cache risultati** search e aggregazioni

```python
# Batch operations per performance
counts = await query_service.get_collection_counts_batch(
    collections=["incantesimi", "oggetti_magici", "mostri"],
    lang="it"
)
```

### 4. Health Checks Avanzati
- **Multi-layer monitoring**: Database, Cache, Application, Performance, Dependencies
- **Status granulari**: Healthy, Degraded, Unhealthy
- **Timeout configurabili** per ogni check
- **Metriche dettagliate** per debugging
- **Endpoint multipli** per load balancer e monitoring

## 📊 Endpoint di Monitoraggio

### `/healthz` - Health Check Semplice
```bash
curl http://localhost:8000/healthz
# Response: "ok" (200) | "unhealthy" (503)
```

### `/health` - Status JSON Basico
```bash
curl http://localhost:8000/health
```
```json
{
  "status": "healthy",
  "timestamp": 1234567890.123,
  "uptime_seconds": 3600.45,
  "summary": {
    "total_checks": 5,
    "healthy": 5,
    "degraded": 0,
    "unhealthy": 0
  },
  "version": "1.0.0"
}
```

### `/health/detailed` - Diagnostica Completa
```bash
curl http://localhost:8000/health/detailed
```
```json
{
  "status": "healthy",
  "timestamp": 1234567890.123,
  "uptime_seconds": 3600.45,
  "version": "1.0.0",
  "checks": [
    {
      "name": "database",
      "status": "healthy", 
      "message": "database is healthy",
      "duration_ms": 15.3,
      "error": null
    },
    {
      "name": "cache",
      "status": "healthy",
      "message": "cache is healthy", 
      "duration_ms": 2.1,
      "error": null
    }
  ],
  "summary": {
    "detailed_status": {
      "database": {
        "ping_ms": 12.5,
        "server_version": "7.0.4",
        "total_commands": 1523,
        "failed_commands": 2,
        "connection_pool": {
          "max_pool_size": 50,
          "min_pool_size": 10
        }
      },
      "cache": {
        "type": "redis",
        "stats": {
          "hits": 856,
          "misses": 234,
          "sets": 445,
          "avg_response_time_ms": 1.8
        }
      }
    }
  }
}
```

## ⚙️ Configurazione Ambiente

### Variabili Ambiente
```bash
# Database
MONGO_URI=mongodb://admin:password@localhost:27017/?authSource=admin
MONGO_MAX_POOL_SIZE=50
MONGO_MIN_POOL_SIZE=10

# Cache Redis
REDIS_URL=redis://localhost:6379
CACHE_DEFAULT_TTL=300
CACHE_ENABLED=true

# Logging
LOG_LEVEL=INFO
STRUCTURED_LOGGING=true
LOG_FILE=/var/log/dnd5e-editor.log

# Performance
QUERY_TIMEOUT_MS=30000
HEALTH_CHECK_INTERVAL=60
```

### Docker Compose con Redis
```yaml
# Copia docker-compose.override.example.yml -> docker-compose.override.yml
services:
  redis:
    image: redis:7-alpine
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
    volumes:
      - redis_data:/data
  
  editor:
    environment:
      - REDIS_URL=redis://redis:6379
      - CACHE_ENABLED=true
    depends_on:
      redis:
        condition: service_healthy
```

## 📈 Metriche e Performance

### 1. Database Performance
- **Connection pooling**: Riduce latenza connessioni
- **Batch operations**: N query → 1 aggregation
- **Index hints**: Query ottimizzate su indici specifici
- **Query timeout**: Prevenzione query lente (30s max)

### 2. Cache Performance  
- **Hit rate target**: >80% per collection counts
- **Response time**: <5ms per cache hit
- **Fallback automatico**: Redis down → in-memory cache
- **Memory limit**: 256MB Redis + 1000 entry fallback

### 3. Monitoring Overhead
- **Health checks**: <50ms totali
- **Background monitoring**: 30s interval
- **Log structured**: JSON per parsing automatico
- **Error tracking**: Stack traces + context

## 🔧 Troubleshooting

### Database Issues
```bash
# Check connection health
curl http://localhost:8000/health/detailed | jq '.summary.detailed_status.database'

# Common problems:
# - High connection count: Increase max_pool_size
# - Slow queries: Check indexes and query timeout
# - Connection timeout: Check server_selection_timeout_ms
```

### Cache Issues  
```bash
# Check cache status
curl http://localhost:8000/health/detailed | jq '.summary.detailed_status.cache'

# Common problems:
# - Redis down: App continues with fallback cache
# - High miss rate: Check TTL settings and cache keys
# - Memory issues: Check Redis maxmemory policy
```

### Performance Monitoring
```bash
# Check response times
curl http://localhost:8000/health/detailed | jq '.checks[] | {name, duration_ms}'

# Monitor query performance
tail -f /var/log/dnd5e-editor.log | jq 'select(.operation == "paginated_find")'
```

## 🚀 Best Practices Produzione

### 1. Load Balancer Configuration
```nginx
upstream dnd5e_backend {
    server app1:8000;
    server app2:8000;
}

location /healthz {
    proxy_pass http://dnd5e_backend;
    proxy_connect_timeout 5s;
    proxy_read_timeout 10s;
}
```

### 2. Monitoring Stack
- **Prometheus**: Scrape `/health` endpoint
- **Grafana**: Dashboard per metriche
- **AlertManager**: Alert su health check failures
- **ELK Stack**: Log aggregation e analysis

### 3. Scaling Considerations
- **Horizontal scaling**: Stateless app + shared Redis/MongoDB
- **Cache warming**: Pre-populate cache dopo deploy
- **Circuit breaker**: Health check based routing
- **Blue-green deployment**: Health checks per traffic switching

### 4. Security
- **Redis AUTH**: Password protection per Redis
- **MongoDB TLS**: Connessioni crittografate
- **Health endpoint**: Rate limiting per `/health/detailed`
- **Log sanitization**: Mask password in connection strings