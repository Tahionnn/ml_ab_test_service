# A/B Testing System for ML Models Pet Project

## Overview

The A/B Testing System for ML Models is a distributed platform designed for controlled online experiments in production environments. It enables parallel comparison of multiple model versions by dynamically splitting user traffic, collecting inference events, and providing management APIs for experiments, model registries, and user access control. The system supports deterministic traffic bucketing, asynchronous event collection, and soft-delete history preservation.

## Architecture

The system follows a microservice architecture with two processing contours:

- **Synchronous Contour**: Handles incoming prediction requests with low latency. Components include an API Gateway and a Traffic Splitter that caches experiment configurations and deterministically assigns requests to model variants.
- **Asynchronous Contour**: Manages model deployment lifecycle and experiment event collection via a message broker. Components include Model Registry, Experiment Registry, User Registry, and Model Deployment Service.

### Communication Patterns

- **RESTful APIs** (FastAPI) for entity management (experiments, models, metrics, users)
- **gRPC** (Go) for low-latency communication between API Gateway and Traffic Splitter
- **Asynchronous messaging** (Apache Kafka) for deployment events and experiment tracking

### Data Storage

- **PostgreSQL** 
- **Redis** 

## Technology Stack

| Component               | Technology                                                                 |
|-------------------------|----------------------------------------------------------------------------|
| API Gateway & Splitter  | Go (chi router, gRPC-go, kafka-go)                                        |
| Management Services     | Python (FastAPI, FastStream, SQLAlchemy, Ray Serve)                        |
| Relational Database     | PostgreSQL                                                                 |
| Cache                   | Redis                                                                      |
| Message Broker          | Apache Kafka      |
| Model Serving           | Ray Serve                      |

## Deployment and Scalability

- Each microservice can be scaled independently.
- Kafka partitions ensure message ordering per experiment/user key.
- Redis caching reduces database load for traffic split decisions.
- Ray Serve provides dynamic scaling for ML model inference.

## Running locally 

Clone the repo. Generate proto and swagger API

```bash
make gen-proto      # compile protobuf for Go
make gen-swagger    # generate OpenAPI docs for api-gateway
```

Spin up whole stack with command

```bash
make up
```
