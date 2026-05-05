# E2E Smoke Checklist

> Final regression checklist for verifying core platform workflows and bilingual capabilities. Execute before each merge to main.

## 1. Core Business Workflow Smoke

### 1.1 Authentication Flow

- [ ] **Login**: Successful login with `admin` / `admin123`, redirected to project management
- [ ] **Register**: New user registration succeeds, redirected to login page
- [ ] **Token Persistence**: Login state persists after page refresh
- [ ] **Logout**: Protected pages inaccessible after logout

### 1.2 Project Management

- [ ] **Create Project**: Fill in project code and name, creation succeeds
- [ ] **List Display**: Project list correctly shows created projects
- [ ] **Empty State**: Empty state message shown when no projects exist

### 1.3 Spider Management

- [ ] **Create Spider**: Select project and fill in details, creation succeeds
- [ ] **List Display**: Spider list filtered by project
- [ ] **Version Management**: Create version and set as current
- [ ] **Registry Auth Ref**: Load and use registry auth refs

### 1.4 Execution Management

- [ ] **Create Execution**: Select project, spider, image, and command, creation succeeds
- [ ] **Execution Detail**: View status, node, trigger, etc.
- [ ] **Execution Logs**: Logs page correctly displays execution output
- [ ] **Resource Limits**: CPU, memory, timeout fields display correctly in create and detail pages

### 1.5 Schedule Management

- [ ] **Create Schedule**: Configure cron expression and retry strategy, creation succeeds
- [ ] **List Display**: Schedule list correctly displays status and configuration
- [ ] **Enable/Disable**: Toggle schedule enabled state

### 1.6 Node Management

- [ ] **Node List**: Display online/offline nodes with heartbeat times
- [ ] **Node Detail**: View heartbeat history and execution history
- [ ] **Sessions**: View node session aggregation view
- [ ] **Status Filter**: Filter execution history by status

### 1.7 Datasource Management

- [ ] **Create Datasource**: Select type and fill in config, creation succeeds
- [ ] **Connection Test**: Test operation returns correct status
- [ ] **Data Preview**: Preview operation returns data rows

### 1.8 Monitor Overview

- [ ] **Overview Page**: Statistics cards display execution and node counters
- [ ] **Refresh**: Click refresh button to update data
- [ ] **Raw Data**: Raw JSON payload section renders correctly

## 2. Bilingual Switching Smoke

### 2.1 Language Switching Basics

- [ ] **Default Language**: First visit defaults to Chinese
- [ ] **Switching**: Language switcher correctly toggles between Chinese and English
- [ ] **Persistence**: Language choice persists after page refresh
- [ ] **Route Preservation**: Current page route is preserved after language switch

### 2.2 Page Copy Coverage

Check bilingual switching for each page:

- [ ] Login page: title, form fields, buttons, error messages
- [ ] Register page: title, form fields, buttons, success message
- [ ] Project management: page title, table headers, buttons, dialogs, empty state
- [ ] Spider management: page title, forms, version management dialog
- [ ] Execution management: create form, detail page, logs page
- [ ] Schedule management: table, cron hints, retry field descriptions
- [ ] Node management: list, detail, sessions, online/offline labels
- [ ] Datasource management: table, test/preview results
- [ ] Monitor page: statistics card titles, raw payload section

### 2.3 Shared Components

- [ ] Navigation menu items display correctly in both languages
- [ ] Loading state text follows language
- [ ] Empty state text follows language
- [ ] Error state text and retry button follow language
- [ ] Toast notification text follows language
- [ ] Confirmation dialog text follows language
- [ ] Element Plus built-in components (pagination, date pickers, etc.) follow language

## 3. Typical Error State Smoke

### 3.1 Authentication Errors

- [ ] **Unauthenticated Access**: Protected page access shows login prompt
- [ ] **Wrong Password**: Login failure shows language-appropriate error message
- [ ] **Duplicate Registration**: Registering an existing user shows conflict message

### 3.2 Request Errors

- [ ] **404 Page**: Non-existent route shows 404 error page
- [ ] **API 404**: Non-existent resource request shows language-appropriate message
- [ ] **API 500**: Service error shows generic error message
- [ ] **Network Error**: Connection error message on network interruption

### 3.3 Form Validation Errors

- [ ] **Required Fields**: Submitting empty form shows field validation messages
- [ ] **Format Validation**: Invalid input shows language-appropriate message

## 4. Documentation and Terminology Smoke

### 4.1 Documentation Completeness

- [ ] Core product docs have bilingual versions
- [ ] Core design docs have bilingual versions
- [ ] Architecture docs have bilingual versions
- [ ] Development guides have bilingual versions

### 4.2 Terminology Consistency

- [ ] Frontend page terminology matches `platform-terminology.md`
- [ ] Documentation terminology matches frontend page terminology
- [ ] API error message terminology matches frontend mappings

### 4.3 Comment Standards

- [ ] New Go files include Chinese file comments
- [ ] Core functions include Chinese function comments
- [ ] Comments explain "why" rather than merely repeating "what"

## Execution Notes

Execute this checklist before each merge to main:

1. Start the full stack: `make migrate && make up`
2. Verify each item against the checklist, record failures
3. Switch to English and re-execute Section 2
4. Run `make down` to clean up

Failing items must be fixed before merging or explicitly documented in the PR description.
