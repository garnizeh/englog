**Task:** Analyze the software architecture documentation contained in the `docs/architecture/` folder and extract a detailed definition for the product evolution phases: MVP, Version 2 (V2), and Version 3 (V3).

**Context:** You are a technical Product Manager reviewing an architecture document to create an executive summary of the development roadmap. Your analysis should go beyond finding keywords and should synthesize information from component descriptions, sequence diagrams, non-functional requirements, and business objectives.

**Response Format:**
For each version (MVP, V2, V3), provide a structured analysis containing:

---

**Phase: [Phase Name - e.g., MVP]**

- **Summary Definition:** (In one sentence, what is the essence of this version?)
- **Business Objective:** (What business goal does this version aim to achieve? E.g., Validate hypothesis X, Increase retention by Y%, Enter market Z.)
- **Main Functional Scope:** (List in bullet point format the main epics or functionalities that define the scope of this delivery.)
- **Key Architecture Decisions:** (List the most impactful architecture decisions for this phase. E.g., "Adoption of monolithic architecture to accelerate initial delivery", "Introduction of payments microservice", "Database migration to support higher data volume".)
- **Readiness Criteria (If mentioned):** (What defines that this phase is complete? E.g., "Capability to support 1,000 concurrent users", "Test coverage of 80% in core module".)
- **Explicit Negative Scope:** (What does the documentation state will NOT be part of this phase?)

---

(Repeat the above structure for V2 and V3)

**Final Instruction:** If the documentation does not use the labels "MVP", "V2", or "V3", infer the phases from the logical progression of functionalities described in the roadmap or document structure.
