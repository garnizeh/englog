# EngLog Architecture Documentation

**Version:** 1.0
**Date:** August 4, 2025
**Author:** Senior Staff Software Architect, Garnizeh
**Status:** Initial Design

---

## 📋 Overview

This directory contains the complete architectural documentation for the EngLog system, organized into specific modules to facilitate navigation and maintenance. The documentation has been structured following **modular architecture** and **separation of concerns** principles.

### 🎯 System Vision

**EngLog** is an innovative personal journal management system that leverages artificial intelligence to transform how users collect, process, and analyze their personal reflections and experiences. The platform empowers users to gain deeper insights into their thoughts, emotions, and personal growth patterns through intelligent data processing and analysis.

### 🏗️ Core Architecture

The system follows a **microservices architecture** pattern with three main components:

1. **API Service (Go)** - Robust backend service for secure journal collection and storage
2. **Worker Pool (Go)** - AI-powered processing engine that enriches journal entries
3. **Web Application** - Intuitive web interface for data visualization and interaction

---

## 📚 Documentation Structure

### 🔹 Main Documents

| Document                                             | Description                                | Status      |
| ---------------------------------------------------- | ------------------------------------------ | ----------- |
| [`OVERVIEW.md`](./OVERVIEW.md)                       | Executive overview and business objectives | ✅ Complete |
| [`SYSTEM_ARCHITECTURE.md`](./SYSTEM_ARCHITECTURE.md) | Original complete document (reference)     | ✅ Complete |

### 🔹 System Components

| Component           | Document                                                           | Description                         |
| ------------------- | ------------------------------------------------------------------ | ----------------------------------- |
| **API Service**     | [`components/API_SERVICE.md`](./components/API_SERVICE.md)         | Detailed REST and gRPC API design   |
| **Worker Pool**     | [`components/WORKER_POOL.md`](./components/WORKER_POOL.md)         | Distributed processing architecture |
| **Database**        | [`components/DATABASE.md`](./components/DATABASE.md)               | Schema and storage strategies       |
| **Web Application** | [`components/WEB_APPLICATION.md`](./components/WEB_APPLICATION.md) | Frontend and user interface         |

### 🔹 Design & Strategies

| Area               | Document                                                 | Description                            |
| ------------------ | -------------------------------------------------------- | -------------------------------------- |
| **Authentication** | [`design/AUTHENTICATION.md`](./design/AUTHENTICATION.md) | OAuth 2.0, OTP and security strategies |
| **Tags & Context** | [`design/TAGS_CONTEXT.md`](./design/TAGS_CONTEXT.md)     | Tag system and context store           |
| **Groups**         | [`design/GROUPS.md`](./design/GROUPS.md)                 | Collaboration and group analytics      |
| **AI Integration** | [`design/AI_INTEGRATION.md`](./design/AI_INTEGRATION.md) | AI services integration                |

### 🔹 Operations

| Area           | Document                                                 | Description                              |
| -------------- | -------------------------------------------------------- | ---------------------------------------- |
| **Deployment** | [`operations/DEPLOYMENT.md`](./operations/DEPLOYMENT.md) | Deployment strategies and infrastructure |
| **Security**   | [`operations/SECURITY.md`](./operations/SECURITY.md)     | Security considerations                  |
| **Testing**    | [`operations/TESTING.md`](./operations/TESTING.md)       | Testing strategy                         |
| **Monitoring** | [`operations/MONITORING.md`](./operations/MONITORING.md) | Observability and monitoring             |

---

## 🚀 Technology Stack

### Backend

- **Language:** Go 1.24+
- **API Framework:** Gin
- **Database:** PostgreSQL with JSONB + pgvector
- **Cache:** Redis
- **Communication:** gRPC (API ↔ Workers)

### Frontend

- **Framework:** Vanilla JavaScript
- **CSS:** Bootstrap/Tailwind (via CDN)
- **Architecture:** Lightweight, dependency-free SPA

### AI & ML

- **Providers:** OpenAI GPT, Azure Cognitive Services, Ollama
- **Embeddings:** pgvector for semantic search
- **Processing:** Distributed asynchronous pipeline

### DevOps

- **Containerization:** Docker & Docker Compose
- **Cloud:** Multi-provider (AWS, GCP, Azure)
- **Authentication:** JWT + OAuth 2.0

---

## 📖 How to Use This Documentation

### 👥 For Developers

1. Start with [`OVERVIEW.md`](./OVERVIEW.md) to understand the context
2. Review [`components/API_SERVICE.md`](./components/API_SERVICE.md) for endpoints and contracts
3. Check [`components/WORKER_POOL.md`](./components/WORKER_POOL.md) for processing
4. See [`operations/DEPLOYMENT.md`](./operations/DEPLOYMENT.md) for local setup

### 👔 For Product Managers

1. Read [`OVERVIEW.md`](./OVERVIEW.md) for business objectives
2. Review [`design/GROUPS.md`](./design/GROUPS.md) for collaborative features
3. Check [`design/AI_INTEGRATION.md`](./design/AI_INTEGRATION.md) for AI capabilities

### 🔧 For DevOps/SRE

1. Focus on [`operations/DEPLOYMENT.md`](./operations/DEPLOYMENT.md)
2. Review [`operations/SECURITY.md`](./operations/SECURITY.md)
3. Check [`operations/MONITORING.md`](./operations/MONITORING.md)

### 🏛️ For Architects

1. Start with [`OVERVIEW.md`](./OVERVIEW.md) for system overview
2. Review all documents in [`components/`](./components/)
3. Check [`design/`](./design/) for architectural decisions

---

## 🔄 Versioning and Updates

| Version | Date       | Major Changes                          |
| ------- | ---------- | -------------------------------------- |
| 1.0     | 2025-08-04 | Initial modular documentation creation |

### 📅 Review Schedule

- **Monthly Review:** First Monday of each month
- **Major Review:** Every major release
- **Next Review:** 2025-09-04

---

## 🤝 Contributions

For documentation updates:

1. 📝 **Small corrections:** Edit the specific file directly
2. 🏗️ **Architectural changes:** Create RFC (Request for Comments) first
3. 📋 **New components:** Add to appropriate directory structure

### 📧 Contacts

- **Principal Architect:** Senior Staff Software Architect
- **Reviews:** team@garnizeh.com

---

## 📋 Document Status

| Document        | Status         | Last Update |
| --------------- | -------------- | ----------- |
| Overview        | ✅ Complete    | 2025-08-04  |
| API Service     | 🚧 In Progress | 2025-08-04  |
| Worker Pool     | 🚧 In Progress | 2025-08-04  |
| Database        | 🚧 In Progress | 2025-08-04  |
| Web Application | 🚧 In Progress | 2025-08-04  |
| Authentication  | 🚧 In Progress | 2025-08-04  |
| Tags & Context  | 🚧 In Progress | 2025-08-04  |
| Groups          | 🚧 In Progress | 2025-08-04  |
| AI Integration  | 🚧 In Progress | 2025-08-04  |
| Deployment      | 🚧 In Progress | 2025-08-04  |
| Security        | 🚧 In Progress | 2025-08-04  |
| Testing         | 🚧 In Progress | 2025-08-04  |
| Monitoring      | 🚧 In Progress | 2025-08-04  |

**Legend:**

- ✅ Complete
- 🚧 In Progress
- ⏳ Planned
- ❌ Pending

---

_This directory serves as the single source of truth for all architectural documentation of the EngLog project. Keep it updated as the system evolves._
