# Execution Listing Completion Design

Date: 2026-05-05
Topic: Complete execution listing query capability
Status: Draft for review

## Background

The roadmap still lists incomplete execution listing query capability as technical debt. Recent commits have already added several execution filters, which indicates the current delivery stream is focused on improving the operability of the execution list rather than starting a brand-new Phase 7 subsystem. This iteration should close the remaining gap and produce a stable, testable, front-to-back execution listing experience.

## Scope

This iteration completes the execution listing feature as a deliverable unit.

Included:
- Backend completion of `GET /api/v1/executions`
- Structured filtering, pagination, sorting, and response normalization
- Frontend completion of `ExecutionsView` query form and pagination workflow
- Backend and frontend tests for key list behaviors
- Small roadmap / technical-debt status update if the work fully closes the roadmap gap

Excluded:
- CSV/export support
- Realtime refresh / WebSocket push
- Advanced search DSL
- DAG/workflow orchestration
- RBAC or multi-tenant behavior changes
- Monitor page redesign

## Goals

1. Make execution list queries stable and consistent for frontend consumption.
2. Eliminate ad-hoc filter additions by introducing a single query model.
3. Deliver a usable operations page with filtering, loading, empty, and pagination states.
4. Keep implementation bounded to the current architecture and recent development direction.

## API Design

Endpoint remains:
- `GET /api/v1/executions`

### Supported query parameters

Filtering:
- `status`
- `project_id`
- `spider_id`
- `node_id`
- `trigger_source`
- `from`
- `to`

Pagination:
- `limit`
- `offset`

Sorting:
- `sort_by`
- `sort_order`

### Defaults

- `sort_by=created_at`
- `sort_order=desc`
- `limit=20`
- `offset=0`
- `limit` upper bound: 100

### Validation rules

- Parsing / format errors return `400`
- `from > to` returns `400`
- Unsupported `sort_by` returns `400`
- Unsupported `sort_order` returns `400`
- `limit < 0` or `offset < 0` returns `400`
- `limit > 100` is normalized to `100`

### Allowed sort fields

- `created_at`
- `started_at`
- `finished_at`
- `status`

### Response shape

```json
{
  "items": [],
  "total": 0,
  "limit": 20,
  "offset": 0
}
```

`total` is the total count under the active filter set, not the current page length.

## Backend Design

### Layering

#### API layer
Responsibilities:
- Parse query params
- Surface validation failures as HTTP 400
- Return normalized paginated response

#### Service layer
Responsibilities:
- Hold a structured execution list query object
- Apply defaults
- Normalize inputs
- Enforce business validation

#### Repository layer
Responsibilities:
- Build SQL predicates for supported filters
- Execute `COUNT(*)` under the same filters
- Execute paginated list query
- Apply whitelist-based sorting

### Query behavior

- Exact-match filters for `status`, `project_id`, `spider_id`, `node_id`, `trigger_source`
- Range filters for `from` and `to`
- If only `from` is provided, query `>= from`
- If only `to` is provided, query `<= to`
- If no filters are provided, return newest executions by default sort

### Compatibility

- Preserve the existing endpoint
- Preserve currently used parameter names where already introduced
- Avoid changing execution detail or execution trigger endpoints

## Frontend Design

Target page: `ExecutionsView`

### Page structure

1. Query panel
   - status
   - project
   - spider
   - node
   - trigger source
   - time range
   - search button
   - reset button

2. Results table
   - execution id
   - project / spider
   - node
   - status
   - trigger source
   - created / started / finished time
   - detail action

3. Pagination area
   - page size
   - current page
   - total count

### Interaction model

- User changes filters, then clicks search to fetch
- Reset restores default query state
- Page changes preserve active filters
- Page-size changes reset to first page
- Loading state shown during requests
- Empty state shown for zero results
- Errors surfaced with existing message mechanism

### State model

Maintain a single query state object containing:
- filters
- pagination
- sorting

This keeps the implementation ready for future URL synchronization without requiring full route-state work in this iteration.

## Parallel Worktree Plan

This implementation will use multiple worktrees with clear ownership.

### Worktree A: backend query completion
- API parameter parsing
- service query model and validation
- repository filtering / sorting / pagination / count
- backend tests

### Worktree B: frontend execution list completion
- `ExecutionsView` query form
- pagination integration
- request parameter mapping
- frontend tests

### Worktree C: integration and verification
- reconcile API/Frontend contract mismatches
- final verification runs
- optional roadmap technical-debt update if fully closed

## Testing Plan

### Backend
Cover at minimum:
- default pagination and default sorting
- single-filter queries
- combined-filter queries
- time-range queries
- invalid parameter cases
- consistency between `total` and filtered results
- empty result set behavior

### Frontend
Cover at minimum:
- filter submission triggers request with expected params
- reset restores defaults
- pagination retains filters
- page-size change resets pagination correctly
- loading / empty / error states render correctly

## Risks and Mitigations

### Risk: frontend and backend diverge on contract
Mitigation:
- lock query parameter names and response shape first
- reserve integration worktree for final alignment

### Risk: scope expands into route sync or dashboard redesign
Mitigation:
- explicitly exclude URL sync and unrelated page refactors

### Risk: parallel worktrees conflict
Mitigation:
- keep ownership split by backend / frontend / integration
- use the integration worktree as the only final merge point

## Success Criteria

This iteration is successful when:
- `GET /api/v1/executions` supports the agreed filters, pagination, and sorting rules
- the frontend execution list page provides complete query and pagination interaction
- automated tests cover the key cases on both backend and frontend
- the roadmap execution-list technical debt can be considered closed or materially reduced

## Self-review

Checklist completed:
- No placeholders or TBDs remain
- Scope is bounded to a single implementation cycle
- Backend and frontend responsibilities are consistent
- Exclusions are explicit to avoid ambiguity
