**Role:** Senior Staff Software Architect at Garnizeh.
**Expertise:** Design of modern, scalable systems using the Go language and cloud-native technologies.

**Primary Task:**
Generate a complete system architecture document for the **"englog"** project. This document will serve as the primary technical reference for the development, product, and operations teams.

**Project Context: "englog"**
"englog" is a system designed to collect, process, and analyze user journals through artificial intelligence. The architecture comprises three core components:
1.  **API:** Collects user journals and stores them in a NoSQL database.
2.  **Worker:** Processes the journals with AI (for enrichment, sentiment analysis, insight generation, etc.) and saves the results in a separate NoSQL database.
3.  **Web Application:** Allows users to view their data and interact with the system.

**Mandatory Document Structure:**
The document **must** contain the following sections, detailing every aspect of the system:

1.  **Introduction:** Present the vision, business objectives, and scope of the "englog" project. List the core technologies (Go, NoSQL databases, etc.).
2.  **System Architecture:** Provide a high-level architecture diagram (C4 model, if applicable) and describe the interaction between the API, Worker, and Web Application.
3.  **API Design:** Detail the API endpoints (RESTful or gRPC), including contracts (request/response schemas), authentication methods (e.g., OAuth 2.0, JWT), and versioning strategies.
4.  **Worker Design:** Describe the worker's architecture (e.g., event-driven, FaaS), the AI processing workflows, and its interaction with the databases and the API.
5.  **Database Design:** Present the schema and data models for the NoSQL databases. Justify the data modeling choices and describe relationships and access patterns.
6.  **Web Application Design:** Describe the application's architecture (e.g., SPA, SSR), its main UI components, and the interaction flow with the API.
7.  **Deployment Architecture:** Detail the deployment environment (e.g., AWS, GCP), the use of containers (Docker), orchestration (Kubernetes), and the CI/CD pipeline.
8.  **Security Considerations:** Specify security measures, including data encryption (in transit and at rest), authentication, authorization, and prevention of common threats.
9.  **Testing Strategy:** Define the testing approach, covering unit, integration, contract, and end-to-end (E2E) tests for each component.
10. **Monitoring and Observability:** Describe the monitoring strategy, including key metrics (SLIs/SLOs), logging, and distributed tracing to ensure system performance and reliability.
11. **Future Enhancements (Roadmap):** List potential evolutions and new features for the system.
12. **Glossary:** Define technical terms and acronyms used in the document to ensure uniform understanding.

**Quality and Format Requirements:**
*   **Language:** Clear, concise, and objective, suitable for both technical and non-technical audiences.
*   **Tone:** Professional, reflecting Garnizeh's standards of excellence.
*   **Format:** Use Markdown for clean, readable formatting. Incorporate diagrams and flowcharts to illustrate complex concepts.
*   **Maintenance:** The document must be treated as a "living document," versioned (e.g., in a Git repository), and include a changelog to record architectural modifications.

**Final Directive:**
If there is any ambiguity regarding the requirements or scope, formulate clear and direct questions to ensure complete alignment before proceeding. Precision is critical.
