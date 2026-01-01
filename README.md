# 🚀 GoProxy - High Performance Service Mesh

![Go Version](https://img.shields.io/badge/go-1.25-blue.svg)
![Node Version](https://img.shields.io/badge/node-20.x-green.svg)
![Docker](https://img.shields.io/badge/docker-compose-blue)
![License](https://img.shields.io/badge/license-MIT-lightgrey)

**GoProxy** é uma infraestrutura de Service Mesh leve e de alto desempenho, projetada com arquitetura **Event-Driven** para gerenciamento dinâmico de tráfego entre microserviços.

Diferente de proxies tradicionais (como Nginx) que exigem "reload" de arquivos de configuração, o GoProxy utiliza um sistema de **Pub/Sub via Redis** para atualizar rotas em tempo real, sem downtime, com latência na casa dos nanossegundos.

---

## 🏗️ Arquitetura do Sistema

O projeto segue o padrão **Control Plane / Data Plane**, comum em grandes orquestradores como Kubernetes e Istio.

```mermaid
graph LR
    User["Client/User"] -->|HTTP Request| DP["Data Plane (Go)"]
    DP -->|"Round Robin"| Backend1["Microservice A"]
    DP -->|"Round Robin"| Backend2["Microservice B"]
    
    Admin["Admin/Dev"] -->|"REST API"| CP["Control Plane (Node.js)"]
    CP -->|"Publish Config"| Redis[("Redis Pub/Sub")]
    Redis -->|"Subscribe Update"| DP