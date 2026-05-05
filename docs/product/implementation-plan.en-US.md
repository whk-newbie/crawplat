# Implementation Plan (by Service / Repository)

> Notes
>
> - This document is based on existing design documents, focusing on "how to deliver" rather than product vision.
> - Tasks are organized by repository / service for easy assignment, scheduling, and verification.
> - New development requirements from the user have been incorporated:
>   - Frontend supports Chinese and English bilingual
>   - Documentation supports Chinese and English bilingual
>   - Development comments use Chinese (function comments, file comments)
> - This plan does not reference `roadmap.md`; it is based solely on current design documents and the requirements stated above.
> - The "implemented / not implemented" status in design documents is not used as a basis; this plan focuses on "what to do, how to do it, and how to verify."

---

## 1. Plan Objectives

### 1.1 Overall Objective

Deliver a sustainable implementation plan without breaking existing API / database / inter-service communication design, covering:

1. Core capabilities for all platform services, aligned with design documents.
2. Web frontend with bilingual (zh-CN / en-US) capability across all core user interfaces.
3. Bilingual version management for product and technical documentation.
4. Unified Chinese comment standards for new and refactored code (function and file comments).
5. Consistency across APIs, data models, documentation, testing, and acceptance criteria.

### 1.2 Success Criteria

Upon completion of this plan, the following results should be achieved:

- Users can switch between Chinese and English in the frontend; page copy, navigation, form hints, error messages, empty states, and confirmation text follow the active language.
- Core design, product, and usage documentation have at least readable versions in both Chinese and English.
- New and significantly modified functions and key files have clear, maintainable Chinese comments.
- APIs, database schemas, and service communication match design documents with no significant deviation.
- Tests cover key workflows, i18n switching, documentation output, and comment convention checks.

---

## 2. Scope and Boundaries

### 2.1 In Scope

This document covers the following repositories / domains:

- Web frontend repository
- Gateway / API Gateway repository
- IAM / Project / Spider / Execution / Scheduler / Node / Datasource / Monitor backend service repositories
- Documentation repository or `docs/` directory
- Testing and QA scripts, conventions, CI configuration

### 2.2 Out of Scope

The following are not priority targets for this plan:

- Reorganizing or rewriting `roadmap.md`
- Brand-new business features not directly related to current design documents
- Large-scale refactoring of legacy code that does not affect the objectives
- UI visual redesign (unless required for i18n)

### 2.3 Principles

- Ensure core workflows are available before polishing the experience.
- Establish unified conventions before expanding to all modules.
- Build reusable foundational capabilities before modifying business pages and documentation.
- i18n, bilingual docs, and comment conventions must be planned in parallel but can be implemented in phases.

---

## 3. Overall Implementation Strategy

### 3.1 Recommended Strategy

Adopt a "foundation first + parallel service delivery + layered verification" approach:

1. First, unify i18n conventions, documentation conventions, and comment conventions.
2. Then, establish the frontend language system: resource files, language switcher, and fallback mechanism.
3. In parallel, document each backend service, clarify error responses, and add comments.
4. Finally, perform cross-repository verification, terminology unification, translation review, and regression testing.

### 3.2 Rationale

- Prevents the frontend, backend, and docs from independently defining language and terminology, avoiding rework.
- Enables reuse of translation resources, terminology tables, comment templates, and documentation templates.
- Ensures future features inherit bilingual capabilities directly instead of retrofitting each time.

---

## 4. Implementation Plan by Repository / Service

---

### 4.1 Web Frontend Repository

#### 4.1.1 Goals

Give the Web frontend complete bilingual (zh-CN / en-US) capability, ensuring:

- Pages support language switching
- Switching preserves the current page state with minimal or no refresh
- Routes, menus, forms, dialogs, empty states, error messages, and notifications fully support i18n
- New languages can be added sustainably

#### 4.1.2 Implementation Items

##### A. i18n Foundation

1. Select and confirm the frontend i18n solution.
2. Establish the language resource directory structure.
3. Set Chinese as the default language with English as a switchable alternative.
4. Create language switching entry points:
   - Top navigation
   - Settings page
   - Post-login user menu
5. Establish language persistence strategy:
   - localStorage / cookie / URL parameter (choose one or combine)
   - Clear priority rules
6. Establish language fallback rules:
   - Fallback to Chinese when English is missing; fallback to English when Chinese is missing
   - No untranslated keys visible on the page

##### B. Page Copy i18n

Gradually replace static text with i18n keys, page by page:

1. Login / Register pages
2. Project list / Project creation page
3. Spider list / Version management page
4. Execution creation / detail / logs pages
5. Schedule list / create / edit pages
6. Node list / detail / heartbeat status page
7. Datasource management page
8. Monitor dashboard / alert rules / alert events pages
9. Shared components:
   - Search bar
   - Table column headers
   - Pagination
   - Dialogs
   - Form validation
   - Toast / Notification
   - Loading / Empty / Error states

##### C. Routing and Navigation i18n

1. i18n for menu names.
2. i18n for breadcrumbs.
3. i18n for page titles.
4. i18n for error pages: 404 / 403 / 500.
5. Replace any hardcoded Chinese in URL paths with language-resource-driven rendering.

##### D. Form and Validation i18n

1. i18n for input labels / placeholders / helper text.
2. Map backend error messages to frontend-readable language.
3. Output form validation messages via language resources.
4. i18n for create / update / delete confirmation dialogs.

##### E. Backend Error to Frontend Mapping

1. Unified error code and error message mapping table.
2. Frontend-translatable handling for common errors:
   - 401 Not authenticated
   - 403 Forbidden
   - 404 Resource not found
   - 409 Conflict
   - 500 Internal server error
3. Preserve server-side raw error text for debugging; do not expose it as the sole display text.

##### F. Frontend Testing

1. Language switching tests:
   - Default language
   - Persistence after refresh
   - Route preservation after switching
2. Key page snapshot tests:
   - Chinese version
   - English version
3. Shared component tests:
   - Fallback for missing keys
   - Long text wrapping
   - Button width adaptation
4. Interaction tests:
   - Successful submission
   - Failed submission
   - Error message translation

##### G. Shared Components (Status: Completed 2026-05-05)

Completed shared component development and migration based on Element Plus, covering Sections A (i18n foundation), B (page copy), C (routing and navigation), D (form validation), E (error mapping), and F (frontend testing):

**Existing components migrated to Element Plus:**
1. AppLayout → `el-container` + `el-header` + `el-menu` (horizontal mode, route-linked highlighting, auto-matching activeIndex)
2. AppLanguageSwitcher → `el-select` (small size, Chinese/English labels, reactive to locale store)
3. AppEmptyState → `el-empty` (description reads i18n key via `localeStore.t()`)
4. AppLoadingState → `el-skeleton` + `v-loading` directive (skeleton / text dual mode, `rows` configurable)

**New shared components:**
1. AppTable — Generic table (`el-table` + `el-pagination`), generic typed, supports sort/pagination/loading/empty state, column headers use i18n keys
2. AppForm — Generic form (`el-form`), supports 6 field types (input / textarea / number / select / switch / date), validation and button labels use i18n
3. AppConfirmDialog — Action confirmation dialog (`ElMessageBox` wrapper), includes `confirmAction` generic function and `confirmDelete` convenience function
4. AppNotification — Toast notification utility (`ElMessage`), exports `notifySuccess` / `notifyError` / `notifyWarning` / `notifyInfo`
5. AppErrorState — Error state display + retry button (`el-result`), message/retry labels use i18n
6. AppBreadcrumb — Breadcrumb navigation (`el-breadcrumb`), items support to/labelKey

**Related enhancements:**
- `el-config-provider` language pack switches reactively with `localeStore` (App.vue provides `elLocale`)
- `api/client.ts` enhanced with `ApiError` class, HTTP status codes auto-mapped to i18n error codes
- `messages.ts` supplemented with 50+ i18n keys: nodes / schedules / errors / actions / notFound / gateway
- All 9 view pages (Login / Projects / Spiders / Executions / ExecutionDetail / Schedules / Nodes / Datasources / Monitor) migrated to Element Plus components + i18n keys
- All 22 tests passing, Vite build successful

#### 4.1.3 Frontend Deliverables

- i18n foundation module
- Chinese and English resource files
- Language switching entry point
- Page copy replacement checklist
- Frontend i18n test cases

#### 4.1.4 Risks

- Some pages may contain hardcoded Chinese strings requiring systematic inspection.
- Table column headers, status enums, and error messages may be scattered across multiple components, making them easy to miss.
- English copy is typically longer, potentially requiring layout adjustments.

---

### 4.2 Gateway Repository

#### 4.2.1 Goals

Ensure the gateway does not block frontend bilingual support, while providing stable, i18n-consumable error responses.

#### 4.2.2 Implementation Items

1. Standardize error response structure so the frontend can map to language.
2. Maintain request-id and access log traceability across the request chain.
3. Emit stable error codes for JWT validation failures, internal token validation failures, rate limiting, etc.
4. Add Chinese comments and necessary documentation for any gateway-layer copy.
5. Verify that API version routing has no side effects on frontend language switching.

#### 4.2.3 Acceptance Criteria

- Gateway error responses can be uniformly handled by the frontend.
- Language switching does not affect authentication, rate limiting, or route forwarding.
- Gateway introduces no new hardcoded copy.

---

### 4.3 IAM / Auth Repository

#### 4.3.1 Goals

Adapt login, registration, and authentication pages and APIs for bilingual support, ensuring authentication errors are translatable.

#### 4.3.2 Implementation Items

1. Chinese / English copy switching for the login page.
2. Chinese / English copy switching for the registration page.
3. Unified mapping for authentication failure messages.
4. Form validation copy i18n.
5. Account status, password rules, and disable notices i18n.
6. Add Chinese comments to authentication-related functions and core files.

#### 4.3.3 Acceptance Criteria

- Not-logged-in, login failure, registration failure, and permission denied messages support language switching.
- Authentication pages display stably in both languages.
- Key authentication logic has readable Chinese comments.

---

### 4.4 Project Service Repository

#### 4.4.1 Goals

Ensure project management pages and APIs are fully expressible in the bilingual frontend system.

#### 4.4.2 Implementation Items

1. i18n for project list, create, and detail pages.
2. i18n for project empty states, hints, validation, and operation confirmations.
3. Consistent use of language resources and formatters for project name, description, status fields.
4. Add Chinese comments to key project service functions.

#### 4.4.3 Acceptance Criteria

- No hardcoded language in the project management main workflow.
- Core operations completable in both Chinese and English environments.

---

### 4.5 Spider Service Repository

#### 4.5.1 Goals

Clearly express Spider management, version management, and registryAuthRef capabilities in frontend and documentation.

#### 4.5.2 Implementation Items

1. i18n for Spider list / create / edit / version pages.
2. Language mapping for version number, registry credential reference, version status fields.
3. i18n for version list and version creation form validation messages.
4. Add Chinese comments to relevant business logic functions.
5. Add Chinese file documentation for version inheritance logic if present.

#### 4.5.3 Acceptance Criteria

- Spider version pages support Chinese / English switching.
- registryAuthRef display and explanation are consistent across docs and pages.
- Key version selection logic has improved readability.

---

### 4.6 Execution Service Repository

#### 4.6.1 Goals

Correctly express Execution creation, detail, logs, retry, and resource limit pages and docs in a bilingual environment.

#### 4.6.2 Implementation Items

1. Execution creation page supports Chinese and English hints.
2. Execution detail page supports bilingual display:
   - Status
   - Retry info
   - Resource limits
   - Version reference
   - registryAuthRef reference
3. Execution logs page supports bilingual UI copy.
4. Resource limit and retry parameter forms support i18n.
5. Add Chinese comments to key execution flow functions, scheduling entry points, and retry materialization logic.
6. Add Chinese file comments for core files related to Mongo logs, Redis queues, and PG state sync.

#### 4.6.3 Acceptance Criteria

- All pages and messages in the execution main workflow support language switching.
- Logs page has no hardcoded Chinese UI copy in English mode.
- Retry, failure, and completion status descriptions are unified.

---

### 4.7 Scheduler Service Repository

#### 4.7.1 Goals

Maintain consistency in schedule management capabilities across documentation, frontend display, and backend comments.

#### 4.7.2 Implementation Items

1. i18n for schedule list, create, and edit pages.
2. i18n for cron explanations, retry strategy explanations, version and credential reference explanations.
3. Add Chinese comments to schedule materialization, deduplication, and trigger logic.
4. Unify schedule error to frontend error message mapping.

#### 4.7.3 Acceptance Criteria

- Chinese and English explanations of schedule tasks are consistent.
- Key materialization flow functions have complete comments.

---

### 4.8 Node Service Repository

#### 4.8.1 Goals

Support bilingual display for node monitoring, heartbeats, sessions, and execution history pages and APIs.

#### 4.8.2 Implementation Items

1. i18n for node list / detail / session pages.
2. i18n for node online, offline, and health status messages.
3. i18n for execution history filter labels.
4. Add Chinese comments to node heartbeat handling, session windows, and execution filtering logic.
5. Add Chinese file comments for state calculation methods, stating module responsibilities.

#### 4.8.3 Acceptance Criteria

- Node monitoring pages support Chinese / English switching.
- Online / offline / session statistics descriptions are clear.

---

### 4.9 Datasource Service Repository

#### 4.9.1 Goals

Unify datasource management, test, and preview functionality in the bilingual UI and documentation.

#### 4.9.2 Implementation Items

1. i18n for datasource list / create / edit pages.
2. i18n for test / preview return descriptions and frontend hints.
3. Map real probe failure reasons to translatable messages.
4. Add Chinese comments to datasource connection testing and preview result formatting.

#### 4.9.3 Acceptance Criteria

- Test and preview messages are clear in both Chinese and English interfaces.
- Datasource core logic is documented.

---

### 4.10 Monitor Service Repository

#### 4.10.1 Goals

Support bilingual display for monitor overview, alert rules, and alert events, ensuring clear alert semantics.

#### 4.10.2 Implementation Items

1. i18n for overview page, rules page, and events page.
2. i18n for alert types, statuses, trigger conditions, and action configurations.
3. i18n for Webhook failure, event records, and rule enable/disable messages.
4. Add Chinese comments to alert matching, polling, and event persistence logic.
5. Add Chinese file comments to monitor aggregation logic.

#### 4.10.3 Acceptance Criteria

- Alert rules and events are accurately explained in both Chinese and English.
- Alert trigger and display flows are unambiguous.

---

### 4.11 Documentation Repository / docs Directory

#### 4.11.1 Goals

Establish a bilingual documentation system so that design, product, and usage documentation can be understood by both Chinese and English readers.

#### 4.11.2 Implementation Items

##### A. Documentation Structure

Recommended structure:

- `docs/zh-CN/`: Chinese primary docs
- `docs/en-US/`: English primary docs
- `docs/shared/`: Terminology table, diagram sources, shared templates
- `docs/product/`: Product documentation
- `docs/design/`: Design documentation
- `docs/architecture/`: Architecture documentation
- `docs/howto/`: Usage guides and operations manuals

##### B. Bilingual Document Mapping

At minimum, ensure the following documents have bilingual versions or bilingual bodies:

1. Product overview
2. MVP / phase descriptions
3. API design
4. Database design
5. Inter-service communication design
6. Frontend design
7. Operations / deployment / local development docs
8. Changelog / release notes

##### C. Documentation Writing Standards

1. When Chinese is the primary language, English documents must maintain the same structure.
2. Terminology table provides unified management to avoid multiple translations for the same concept.
3. Heading levels, list structures, and table fields must be consistent for bidirectional maintenance.
4. Design diagrams, tables, and API examples should have bilingual descriptions where possible.
5. All new documents must have at minimum:
   - Chinese description
   - English summary or full English text

##### D. Maintenance Approach

1. Create bilingual versions simultaneously for new documents.
2. When modifying core design docs, update bilingual files in sync.
3. If full translation is not immediately possible, provide at minimum an English summary, with a clear annotation of the pending completion task.

#### 4.11.3 Acceptance Criteria

- Core documents can be read independently per language directory.
- Terminology does not drift between the Chinese and English versions.
- Document structure and content can be traced between versions.

---

### 4.12 Code Comment Standards and Enforcement

#### 4.12.1 Goals

Ensure new and significantly modified code has readable, maintainable Chinese comments.

#### 4.12.2 Implementation Items

##### A. Function Comments

The following types of functions should have Chinese comments:

1. Core business functions
2. Data transformation functions
3. State calculation functions
4. External call wrapper functions
5. Complex conditional branching functions
6. Retry / scheduling / message delivery / sync logic functions

Comments should at minimum explain:

- What the function does
- Input and output semantics
- Key side effects or boundary conditions
- Why it was implemented this way, if non-obvious

##### B. File Comments

The following types of files should have Chinese file comments:

1. Business modules with a clear single responsibility
2. Adaptation layers interacting with external systems
3. Complex state machines / queues / scheduling / alert processing modules
4. Shared utility classes, shared configuration, and shared constant files

File comments should explain:

- What the file is responsible for
- Who it interacts with
- What it is NOT responsible for

##### C. Comment Boundaries

1. Do not add pointless comments to obvious single-line code.
2. Do not write "translation comments" that merely repeat what the code says.
3. Prioritize adding explanations for complex logic, public boundaries, and easily misunderstood areas.
4. Comments must be updated in sync with implementation changes; avoid stale comments.

#### 4.12.3 Acceptance Criteria

- New core functions are readable.
- New complex files have Chinese descriptions.
- Comments cover "why this approach" rather than merely repeating "what it does."

---

## 5. Unified Language System Design

### 5.1 Frontend Language Strategy

Recommended language system:

- Default language: Chinese `zh-CN`
- Secondary language: English `en-US`
- Switching strategy: User can manually switch; browser language can serve as the first-visit default
- Persistence strategy: Remember the user's last selection

### 5.2 Terminology Strategy

Establish a unified terminology table covering at minimum:

- Project / 项目
- Spider / 爬虫
- Spider Version / 爬虫版本
- Execution / 执行
- Schedule / 调度任务
- Node / 节点
- Datasource / 数据源
- Alert Rule / 告警规则
- Alert Event / 告警事件
- Registry Auth Ref / 镜像仓库凭据引用

### 5.3 Copy Style Requirements

Chinese:

- Concise, clear, avoid colloquial ambiguity

English:

- Concise, standard, avoid direct word-for-word translation from Chinese

### 5.4 Translation Boundaries

- Status values, enum values, error codes, and field names should remain stable; only the display layer is translated.
- Do not directly translate database field names or API parameter names.
- Frontend copy and documentation must use the same terminology table.

---

## 6. Recommended Development Order

### 6.1 Phase 1: Foundation and Conventions

1. Establish frontend i18n foundation.
2. Establish bilingual documentation directory structure and templates.
3. Establish Chinese comment convention examples.
4. Establish terminology table.
5. Unify error messages and translation mapping principles.

### 6.2 Phase 2: Frontend Core Workflow i18n

1. Login / Register
2. Project management
3. Spider / Execution / Schedule / Node / Datasource / Monitor
4. Shared components and error pages

### 6.3 Phase 3: Backend Documentation and Comments

1. Add Chinese comments to core logic files in each service
2. Bilingual API documentation
3. Bilingual database, communication, deployment, and development documentation

### 6.4 Phase 4: Unified Acceptance

1. Chinese / English switching regression tests
2. Documentation consistency checks
3. Comment convention checks
4. Key workflow smoke tests

---

## 7. Testing and Verification Plan

### 7.1 Frontend Tests

- i18n initialization tests
- Language switching tests
- Route preservation tests
- Empty state / error state tests
- Long copy adaptation tests
- Component snapshot tests

### 7.2 Backend Tests

- API response structure tests
- Error code mapping tests
- registryAuthRef inheritance / propagation tests
- Resource limit field tests
- Schedule / retry / alert flow tests

### 7.3 Documentation Tests

- Bilingual directory pairing checks
- Terminology table consistency checks
- Core document language version existence checks
- Code examples and API description consistency checks

### 7.4 Comment Inspection

- Core functions have Chinese comments
- File responsibilities have Chinese descriptions
- No obviously stale comments

---

## 8. Milestone-Based Delivery Checklist

### 8.1 Milestone A: Conventions Established

Deliverables:

- Frontend i18n convention
- Bilingual documentation convention
- Chinese comment convention
- Terminology table

### 8.2 Milestone B: Bilingual Frontend Usable

Deliverables:

- Login / Register Chinese / English switching
- Main menu / core workflow page Chinese / English switching
- Component-level i18n capability

### 8.3 Milestone C: Bilingual Documentation Readable

Deliverables:

- Core design documents bilingual
- Product documents bilingual
- Development / deployment documentation bilingual

### 8.4 Milestone D: Code Maintainability Improved

Deliverables:

- Key files have Chinese comments
- Core functions have Chinese comments
- Comment convention checks in place

### 8.5 Milestone E: Unified Acceptance Passed

Deliverables:

- Chinese / English bilingual regression passed
- Documentation consistency confirmed
- Comment conventions confirmed
- Key workflow smoke tests passed

---

## 9. Risks and Mitigations

### 9.1 Risk: Translation Resource Bloat

**Symptom**: Many pages, much copy, complex key management.
**Mitigation**:

- Establish naming conventions
- Split namespaces by service
- Create shared common keys
- Avoid duplicate keys

### 9.2 Risk: Terminology Inconsistency

**Symptom**: Same feature translated differently across pages and docs.
**Mitigation**:

- Unified terminology table
- Unified reviewer
- Freeze core terminology before bulk translation

### 9.3 Risk: Copy Length Breaking Layout

**Symptom**: English buttons, labels, and descriptions overflow layout.
**Mitigation**:

- Run English-length regression testing early
- Use adaptive layouts
- Reserve expansion space for key buttons and table columns

### 9.4 Risk: Inconsistent Comment Quality

**Symptom**: Comments too many, too few, or stale.
**Mitigation**:

- Only require comments for complex logic and core modules
- Establish comment templates
- Review comments alongside code changes

### 9.5 Risk: High Bilingual Document Sync Cost

**Symptom**: English version falls behind after Chinese document updates.
**Mitigation**:

- Commit bilingual documents as pairs
- Prioritize dual-writing for core documents
- Use templates and checklists to reduce omissions

---

## 10. Final Acceptance Criteria

This implementation plan is considered complete when all of the following are met:

1. Frontend core functional pages support Chinese / English switching.
2. Core documentation has at least readable versions in both languages.
3. New and significantly modified code has function and file comments in Chinese, and they are maintainable.
4. Key APIs, database fields, and communication paths match design documents.
5. Core workflow tests and i18n regression tests pass.
6. Core terminology is consistent across pages and documentation.

---

## 11. Appendix: Recommended Task Breakdown Template

Each repository is recommended to further break down tasks using this format:

- Objective
- Scope of impact
- List of changes
- New files
- Modified files
- Dependencies
- Acceptance method
- Rollback method

This ensures changes in any repository can be clearly tracked and easily managed as work tickets.

---

## 12. Assignable Task List by Repository

> Note: The following list is broken down by the criteria of "assignable, schedulable, verifiable." Tasks within each repository can be further split into multiple tickets. Recommended execution order: "Foundation → Pages/APIs → Testing → Documentation/Comments → Acceptance."

### 12.1 Web Frontend Task List

#### 12.1.1 i18n Foundation

- Select and implement the frontend i18n solution with a unified language pack loading approach.
- Define the language resource directory, namespaces, and key naming convention.
- Implement language persistence read/write with clear priority: user manual selection > local cache > browser language > default Chinese.
- Add a global language switching entry point that preserves the current page state as much as possible.
- Create a common translation utility supporting placeholders, plurals, dynamic interpolation, and fallback.

#### 12.1.2 Shared Component i18n

- Update buttons, labels, dialogs, toasts, form validation, pagination, empty states, loading states, and error pages to read from language resources.
- Create unified translation wrappers for table column headers, status labels, time formats, and quantity units.
- Find and replace hardcoded Chinese strings with language keys.
- Reserve style space for longer English copy to avoid button and title overflow.

#### 12.1.3 Page-Level i18n

- Replace all static text with bilingual copy on login, registration, and password recovery pages (if they exist).
- Replace static text module by module for Project, Spider, Execution, Schedule, Node, Datasource, and Monitor core pages.
- Unify i18n for route titles, breadcrumbs, menu names, permission hints, and empty state messages.
- Create a unified "status code → display text" mapping for page-level enum statuses.

#### 12.1.4 Error and Exception Handling

- Map common errors (401 / 403 / 404 / 409 / 500 / 502) to frontend-translatable copy.
- Unify the display style and copy strategy for network errors, timeout errors, permission errors, and validation errors.
- Preserve backend raw error fields for debugging but do not use them as the sole display text.

#### 12.1.5 Frontend Test Tasks

- Add i18n initialization and fallback unit tests.
- Add page state preservation tests after language switching.
- Add Chinese / English snapshot tests for core pages.
- Add adaptation tests for long copy, narrow screens, pagination, and table column overflow.
- Add detection for untranslated keys, or build-time checks.

#### 12.1.6 Frontend Acceptance Criteria

- Language switching entry point is usable and persistent.
- Core pages have no hardcoded visible Chinese or English residues.
- Key interaction workflows complete successfully in both languages.
- Missing translation keys have clear fallback behavior; no blank copy appears.

---

### 12.2 Gateway Task List

#### 12.2.1 Unified Error Responses

- Fix the gateway error response structure so the frontend can map uniformly.
- Define stable error semantics for auth failures, internal token failures, rate limiting, and version mismatches.
- Check for visible copy in the gateway; replace with Chinese comments or standardized error codes.

#### 12.2.2 Route and Auth Compatibility

- Verify that language switching does not affect route forwarding.
- Verify consistent behavior of JWT validation, internal token validation, and rate limiting across languages.
- Add Chinese file comments for gateway response headers or log fields that need explanation.

#### 12.2.3 Gateway Acceptance Criteria

- Gateway error responses are stable and predictable.
- Frontend can uniformly handle gateway errors without depending on Chinese or English original text.
- Gateway introduces no new UI copy coupling.

---

### 12.3 IAM / Auth Task List

#### 12.3.1 Auth Page i18n

- Full bilingual support for login page, register page, password rule hints, and account status messages.
- Form validation, button text, error messages, and success messages all use language packs.
- Unified frontend mapping for session expired, credential invalid, and permission denied messages.

#### 12.3.2 Auth Logic and Documentation

- Add Chinese function comments to core authentication functions.
- Add Chinese file comments to main auth module files, stating responsibility boundaries and dependencies.
- Add behavior descriptions for token refresh, session management, and password policy logic if present.

#### 12.3.3 IAM Acceptance Criteria

- Login and registration workflows are fully functional in both Chinese and English.
- Authentication failures, form errors, and permission errors display the correct language copy.
- Key authentication logic has maintainable Chinese comments.

---

### 12.4 Project Service Task List

#### 12.4.1 Project Management Page i18n

- Unified i18n for project list, create, detail, edit, and delete confirmation dialogs.
- Language-resource-controlled display for project status, labels, description, member count, and time fields.
- Bilingual empty-state copy for no projects, no permissions, and no search results.

#### 12.4.2 Project Service Logic Comments

- Add Chinese comments to project create, query, update, and delete core functions.
- Add Chinese file descriptions for project domain models, DTO transformations, and validation logic.

#### 12.4.3 Project Acceptance Criteria

- No hardcoded language on main project management pages.
- Core project logic is readable and maintainable.

---

### 12.5 Spider Service Task List

#### 12.5.1 Spider Page and Version Management i18n

- Full i18n for Spider list, create, edit, version list, and version publish pages.
- Unified handling of display copy for registryAuthRef, version number, version status, and inheritance relationships.
- Bilingual messages for version selection, version conflicts, and version-not-found states.

#### 12.5.2 Spider Business Logic Comments

- Add Chinese comments to version creation, version inheritance, version query, and credential reference resolution functions.
- Add Chinese file comments for files related to version history, version snapshots, and publish workflows.

#### 12.5.3 Spider Acceptance Criteria

- Spider and Spider Version display terminology is consistent across pages and docs.
- Version-related error messages support Chinese / English switching.

---

### 12.6 Execution Service Task List

#### 12.6.1 Execution Management Page i18n

- Full bilingual support for execution create, detail, logs, retry, failure reason, and completion status.
- Bilingual display of resource limit fields (CPU, memory, timeout) in forms and detail pages.
- Unified translation of hints for spiderVersion, registryAuthRef, retryLimit, and related fields.

#### 12.6.2 Execution Flow Core Logic Comments

- Add Chinese comments to execution create, claim, start, log append, complete, fail, and retry materialization functions.
- Add Chinese file comments for Redis queue handling, Mongo log writing, and PG state sync files.
- Add behavior descriptions for task retry strategy, resource limit validation, and execution state machine.

#### 12.6.3 Execution Acceptance Criteria

- The execution main workflow can be completed in both Chinese and English.
- Logs, detail, and status pages have no visible hardcoded Chinese residues.
- Execution state transition logic has clear Chinese comment explanations.

---

### 12.7 Scheduler Service Task List

#### 12.7.1 Schedule Management i18n

- i18n for schedule list, create, edit, and detail pages.
- Unified translation for cron explanations, retry strategy, trigger conditions, and enable/disable states.
- Unified handling of materialization tasks, deduplication strategy, and trigger window hints.

#### 12.7.2 Schedule Logic Comments

- Add Chinese comments to cron parsing, task scanning, materialization generation, and deduplication advancement logic.
- Add Chinese file comments for schedule task files, clarifying responsibility boundaries.

#### 12.7.3 Scheduler Acceptance Criteria

- Schedule function pages and error messages support bilingual display.
- Complex scheduling logic has readable Chinese explanations.

---

### 12.8 Node Service Task List

#### 12.8.1 Node Monitoring i18n

- Full i18n for node list, node detail, node sessions, and execution history filter criteria.
- Bilingual status messages for online, offline, abnormal, healthy, and heartbeat timeout.
- Unified handling of node session, execution count, and execution time range copy.

#### 12.8.2 Node State Logic Comments

- Add Chinese comments to heartbeat write, online status determination, session window, and history filtering functions.
- Add Chinese file comments for the main node management file, explaining how node state is calculated.

#### 12.8.3 Node Acceptance Criteria

- Node status display is consistent and accurate in both Chinese and English.
- Node session and execution history filtering logic has explanatory comments.

---

### 12.9 Datasource Service Task List

#### 12.9.1 Datasource Page i18n

- Full bilingual support for datasource list, create, edit, test, and preview pages.
- i18n for test / preview success, failure, timeout, and format error messages.
- Unified translation of connection info, test results, and preview result descriptions.

#### 12.9.2 Datasource Logic Comments

- Add Chinese comments to connection test, preview request, result normalization, and error classification functions.
- Add Chinese file comments for datasource adapter layer, probe logic, and result parsing files.

#### 12.9.3 Datasource Acceptance Criteria

- Test and preview messages are clear and understandable in both Chinese and English.
- Datasource core logic has comments and maintainability.

---

### 12.10 Monitor Service Task List

#### 12.10.1 Monitor and Alert Page i18n

- Full i18n for overview page, alert rules page, and alert events page.
- Unified translation for alert types, trigger conditions, alert statuses, event statuses, and Webhook messages.
- Consistent display copy for failed execution and offline node alert rules.

#### 12.10.2 Alert Logic Comments

- Add Chinese comments to rule matching, event persistence, Webhook sending, failure retry, and offline detection functions.
- Add Chinese file comments for monitor aggregation, polling scheduling, and event construction files.

#### 12.10.3 Monitor Acceptance Criteria

- Alert rules and events have consistent meaning in both Chinese and English, with no ambiguity.
- Core alert processing logic is readable and traceable.

---

### 12.11 Documentation Repository / docs Directory Task List

#### 12.11.1 Bilingual Documentation Structure

- Create Chinese primary document directory and corresponding English directory.
- Create unified templates for product docs, design docs, architecture docs, deployment docs, and development docs.
- Create a terminology table; freeze core term translations to prevent version drift.

#### 12.11.2 Core Document Bilingualization

- Provide bilingual versions for API design, database design, service communication design, frontend design, MVP description, and usage documentation.
- Perform bilingual review of images, tables, API examples, and field explanations.
- Gradually complete existing documents by "structure first, then translate section by section."

#### 12.11.3 Document Maintenance Process

- Generate Chinese and English versions simultaneously for new documents.
- Update corresponding language versions when modifying core documents.
- If full translation is not possible in the short term, at minimum provide an English summary with an explicit marker of pending content.

#### 12.11.4 Documentation Acceptance Criteria

- Core documents can be read independently per language.
- Chinese and English versions have consistent structure, consistent terminology, and consistent examples.
- No cases where one version is updated while the other lags behind long-term.

---

### 12.12 Code Comment Standards and QA Task List

#### 12.12.1 Comment Convention Implementation

- Create function comment templates with a unified Chinese description style.
- Create file comment templates describing module responsibilities, dependencies, and boundaries.
- Define which functions must have comments: core flows, state transitions, data transformations, external calls, retry/scheduling, error handling.

#### 12.12.2 Comment Quality Checks

- Add comment check items to code reviews.
- Create a manual checklist for new and significantly modified files.
- Avoid "repeating what the code does" comments; require comments that explain "why it was done this way."

#### 12.12.3 Comment Acceptance Criteria

- Core modules have Chinese file comments.
- Complex functions have Chinese function comments.
- Comments are in sync with implementation; no obviously stale descriptions.

---

### 12.13 Cross-Repository Shared Task List

#### 12.13.1 Unified Terminology and Translation Glossary

- Create a shared terminology table and freeze core terms.
- Unify translations for Project / Spider / Execution / Schedule / Node / Datasource / Alert and other terms.
- Define conventions for error code display text and status enum translations.

#### 12.13.2 Unified Testing and Acceptance

- Create frontend Chinese / English switching smoke tests.
- Create core page snapshot regression tests.
- Create a bilingual documentation consistency checklist.
- Create a comment checklist and code review requirements.

#### 12.13.3 Shared Task Acceptance Criteria

- Global terminology is consistent.
- Global language switching behavior is consistent.
- Global documentation maintenance approach is consistent.
- Global comment convention enforcement is consistent.
