# Business Branches

| Area | Branch | Status | Notes |
|------|--------|--------|-------|
| MBTI full report generation | document.xml contains static body runs with w14:textFill / w14:props3d | ✅ | Triggered by production tofu-box issue; now covered by FB-042 fallback |
| MBTI full report generation | document.xml contains risky static body font families such as HYYakuHei / 汉仪雅酷黑 | ✅ | Covered by FB-043 font-family normalization fallback |
| MBTI full report generation | document.xml contains East Asian static body runs with only w:hint and no explicit font family | ✅ | Triggered by production ESTP "功利型/凭借" tofu-box issue; covered by FB-044 |
| MBTI full report generation | styles.xml / fontTable.xml declare unstable CJK fonts used by body fallback | ✅ | Covered by FB-044 style/font-table normalization |

## RuoYi Administration Stub Inventory

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| User profile | update profile/password/avatar and query/update assigned roles | P1 | ✅ | FB-088～092: authenticated self-service, bcrypt/session invalidation, decoded safe avatar and permission-checked transactional role replacement |
| Role authorization | status/data scope, allocated/unallocated users, cancel/select authorization | P1 | ✅ | FB-091～092: exact RuoYi permissions, super-admin protection, bounded paging and transactional validated relationships |
| Public registration | `POST /register` | P1 | ✅ | FB-093/095: config gate, atomic one-time captcha, strict credentials, bcrypt, transactional optional common role and unique-index migration |
| Cache administration | config/dict refresh | P2 | ✅ | FB-096/097: real permission-checked SCAN refresh handlers replace both success stubs and preserve login/captcha keys |
| Monitoring administration | force logout, jobs/job logs, operlog/logininfor cleanup | P2 | ✅ | FB-098/099 retire online/job/jobLog modules and unsupported audit mutations; real server and audit list routes remain |
| Code generation | `/tool/gen/*` | P3 | ✅ | FB-098/099 remove the unsupported backend routes, frontend module, hidden edit route, and menu access |
| User repo/wrong-book generic CRUD | 18 generated GET/POST routes | P3 | ✅ | FB-098 removes all generated stub routes and four confirmed unconsumed wrappers; `el_user_book` model/table/data remain |

### Administration Stub Closure Branch Matrix (2026-07-27)

| Function | Branch | Priority | Coverage |
|----------|--------|----------|----------|
| Profile update | authenticated current user updates valid nickname/email/phone/sex | P0 | ✅ |
| Profile update | forged user ID, malformed JSON, invalid field format, duplicate email/phone, missing user or DB failure | P0 | ✅ |
| Profile password | correct old password and strong new password | P0 | ✅ |
| Profile password | wrong old password, weak/same password, malformed JSON, hash/update failure | P0 | ✅ |
| Profile avatar | valid JPEG/PNG within size limit is stored under configured profile directory | P0 | ✅ |
| Profile avatar | empty/oversized/fake image, path manipulation, file/DB failure | P0 | ✅ |
| Profile avatar | persisted `/profile/...` URL is read through the nginx static location without an API-prefix rewrite | P0 | ✅ | FB-100: frontend store and upload completion keep the backend URL unchanged; regression test rejects `/prod-api/profile/...` |
| User role assignment | authorized caller reads or transactionally replaces valid active roles | P0 | ✅ |
| User role assignment | unauthorized caller, missing user/role, duplicate/invalid role ID, protected admin mutation or DB rollback | P0 | ✅ |
| Role authorization | authorized caller changes status/data scope and assigns or removes users transactionally | P0 | ✅ |
| Role authorization | unauthorized caller, malformed input, protected admin role/user, invalid IDs or DB rollback | P0 | ✅ |
| Registration | enabled registration with valid one-time captcha, unique username and strong password | P0 | ✅ |
| Registration | disabled registration, invalid/replayed captcha, malformed/duplicate username, weak/mismatched password or DB rollback | P0 | ✅ |
| Config/dict cache | cache miss/hit, exact write invalidation and prefix refresh | P1 | ✅ | FB-096/097: one-hour config read-through; config and dict mutations invalidate old/new exact keys; prefix refresh uses batched SCAN |
| Config/dict cache | Redis unavailable or database query failure returns controlled behavior without false success | P1 | ✅ | FB-096/097: config/dict reads fall back from Redis to DB; non-not-found DB failures and refresh failures return controlled errors; empty dict is `[]` |
| Dictionary routing | public type/batch reads bypass JWT while all type/data management routes retain authenticated login context | P0 | ✅ | FB-101: method-specific public reads remain anonymous; broad `/system/dict/` prefix removed so management handlers receive the authenticated login context |
| Optional modules | retired job/code-generation menus and hidden routes are absent | P1 | ✅ | FB-098/099 source and frontend tests; menu IDs disabled by explicit primary-key migration |
| Audit pages | operation/login audit lists remain available while unsupported mutation controls/routes are absent | P1 | ✅ | Read-only list routes/wrappers/pages retained; delete/clean/export controls and routes removed |
| Dead generic routes | user repo/wrong-book generated stubs return 404 and historical data remains untouched | P1 | ✅ | Authenticated HTTP tests return 404; no table/model/data migration is performed |
| Stub inventory | production router has zero `Stub`/`AjaxStub`/`TableStub` registrations and no `_todo` success response | P0 | ✅ | Production route source has no stub caller/helper; `internal/handler/stub.go` deleted |

## Competency Assessment — Phase 1A Security and Dispatch Baseline

> Scope: explicit assessment dispatch, exact anonymous routes, participant/paper token validation, and legacy API isolation.  
> Rule: all planned branches start as uncovered. P0 branches must become ✅ before Phase 1A can be accepted.

### A. Assessment Type Dispatch

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Assessment dispatch | `assessment_type=legacy` and `scoring_mode=legacy` | P0 | ✅ | Covered by `TestValidateAssessmentMode/legacy_pair` |
| Assessment dispatch | `assessment_type=competency` and `scoring_mode=competency_average` | P0 | ✅ | Covered by `TestValidateAssessmentMode/competency_pair` |
| Assessment dispatch | competency type with legacy scoring mode | P0 | ✅ | Covered by `TestValidateAssessmentMode/competency_with_legacy_scoring` |
| Assessment dispatch | legacy type with competency scoring mode | P0 | ✅ | Covered by `TestValidateAssessmentMode/legacy_with_competency_scoring` |
| Assessment dispatch | unknown or empty assessment type on a new competency request | P0 | ✅ | Covered by unknown/empty cases in `TestValidateAssessmentMode` |
| Assessment dispatch | existing database row has no new type before migration backfill | P1 | ✅ | Staging migration backfilled all legacy rows and repeated idempotent execution without changing valid combinations |

### B. Exact Anonymous Route Matching

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Anonymous routing | exact `POST /exam/api/competency/participant/create-paper` | P0 | ✅ | Exact matcher and HTTP middleware call covered |
| Anonymous routing | exact `POST /exam/api/competency/participant/paper-detail` | P0 | ✅ | Covered by `TestCompetencyParticipantRoutesUseExactMethodAndPath` |
| Anonymous routing | exact `POST /exam/api/competency/participant/fill-answer` | P0 | ✅ | Covered by `TestCompetencyParticipantRoutesUseExactMethodAndPath` |
| Anonymous routing | exact `POST /exam/api/competency/participant/submit` | P0 | ✅ | Covered by `TestCompetencyParticipantRoutesUseExactMethodAndPath` |
| Anonymous routing | same participant path with a different HTTP method | P0 | ✅ | GET/PUT/DELETE/PATCH negative cases covered |
| Anonymous routing | participant path with an added suffix or child path | P0 | ✅ | Suffix negatives and HTTP 401 middleware call covered |
| Anonymous routing | competency dimensions/questions/exams management endpoints | P0 | ✅ | Covered by `TestCompetencyManagementRoutesRequireAdminJWT` |
| Anonymous routing | competency results/export/admin report endpoints | P0 | ✅ | Covered by `TestCompetencyManagementRoutesRequireAdminJWT` |
| Anonymous routing | exact internal report-data endpoint | P0 | ✅ | Exact GET plus method/path negatives covered; handler token check remains a later report-handler branch |
| Anonymous routing | existing login/captcha/MBTI/legacy participant paths | P0 | ✅ | Existing anonymous behavior retained by current middleware tests; rerun after change |
| Anonymous routing | existing tester/qu/system management paths | P0 | ✅ | Existing non-anonymous behavior retained by current middleware tests; rerun after change |

### C. Participant and Paper Token Validation

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Participant token | valid HS512 signature, unexpired token, expected purpose and matching participant/exam | P0 | ✅ | Round-trip and binding tests pass |
| Participant token | token missing | P0 | ✅ | Covered by missing-token case |
| Participant token | malformed token | P0 | ✅ | Covered by malformed-token case |
| Participant token | wrong signature or non-HS512 algorithm | P0 | ✅ | Wrong-secret case plus existing JWT algorithm tests pass |
| Participant token | expiration claim missing | P0 | ✅ | Covered by missing-expiration case |
| Participant token | expired token | P0 | ✅ | Covered by expired-token case |
| Participant token | wrong `purpose` claim | P0 | ✅ | Covered by wrong-purpose case |
| Participant token | participant ID does not match request/database owner | P0 | ✅ | Covered by `ValidateBinding/wrong_participant` |
| Participant token | exam ID does not match requested exam | P0 | ✅ | Covered by `ValidateBinding/wrong_exam` |
| Paper token | valid token matches participant, exam, and paper | P0 | ✅ | Covered by round-trip and valid binding |
| Paper token | paper ID does not match request | P0 | ✅ | Covered by `ValidateBinding/wrong_paper` |
| Paper token | participant or exam claim differs from paper ownership | P0 | ✅ | Participant and exam mismatch cases covered |
| Token handling | token or signing secret appears in log/error response | P0 | ✅ | Rejection tests assert generic errors contain neither token nor secret; implementation logs neither |

### D. Legacy Paper API Isolation

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Legacy `CreatePaper` | exam does not exist | P0 | ✅ | Controlled not-found mapping covered by guard tests; staging legacy smoke verified valid paths remain reachable |
| Legacy `CreatePaper` | exam is competency | P0 | ✅ | Runtime mode test plus source-order test prove rejection before `createPaperTx` |
| Legacy `CreatePaper` | exam is legacy 001/002/003 | P0 | ✅ | Staging created temporary 00101/00201/00301 exams and two-question papers |
| Legacy paper read APIs | paper does not exist | P0 | ✅ | Guard tests cover controlled not-found; staging valid reads confirm no false competency rejection |
| Legacy paper read APIs | paper belongs to competency exam | P0 | ✅ | Guard/order tests cover `paper-detail`, `paperQu-detail`, `qu-detail`, `paper-result`, and `stand-score` |
| Legacy paper read APIs | paper belongs to legacy exam | P0 | ✅ | Staging paper-detail/qu-detail succeeded for 00101/00201/00301 |
| Legacy `FillAnswer` | paper belongs to competency exam | P0 | ✅ | Guard executes before empty-answer success and before write transaction |
| Legacy `FillAnswer` | paper belongs to legacy exam | P0 | ✅ | Staging answered two questions for each 00101/00201/00301 temporary paper |
| Legacy `HandExam` | paper belongs to competency exam | P0 | ✅ | Guard executes before write transaction and is repeated inside transaction before aggregation |
| Legacy `HandExam` | paper belongs to legacy exam | P0 | ✅ | Staging submitted and read paper-result for all three legacy types |
| Legacy tester/candidate standard-score endpoints | paper belongs to competency exam | P0 | ✅ | Both handlers guard before repo lookup and fixed formula query |
| Legacy tester/candidate standard-score endpoints | paper belongs to legacy 001/002 | P0 | ✅ | Formula unit tests plus staging stand-score response for 00101/00201; 00301 generic endpoint remained reachable |
| Legacy guard query | database lookup fails | P0 | ✅ | Any non-not-found DB error maps to controlled assessment-mode error; no legacy fallback |

### E. Phase 1A Acceptance Gate

| Gate | Priority | Coverage | Evidence required |
|------|----------|----------|-------------------|
| All new P0 dispatch branches | P0 | ✅ | `TestValidateAssessmentMode` passes |
| All exact anonymous route branches | P0 | ✅ | Direct matcher and HTTP middleware tests pass |
| All participant/paper token rejection branches | P0 | ✅ | Token tests pass with no secret output |
| All legacy API competency guards | P0 | ✅ | Runtime mode tests and handler source-order tests prove guards precede legacy reads/writes/formulas |
| Legacy 001/002/003 regression | P0 | ✅ | Go full suite plus staging temporary create-paper/detail/fill/submit/result/stand-score chains passed; cleanup=0 |
| Build and test gate | P0 | ✅ | `Go: Build` and `Go: Test All` passed on 2026-07-24 |

### F. Competency Report Audience Version

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Report audience | competency exam selects `frontline_employee` | P0 | ✅ | Staging create/detail round-trip verified |
| Report audience | competency exam selects `leader` | P0 | ✅ | Staging draft edit/detail and DB row verified |
| Report audience | competency exam omits audience or sends an unknown value | P0 | ✅ | Empty and unknown values rejected by `TestValidateCompetencyReportAudience` |
| Report audience | legacy exam has no competency report audience | P0 | ✅ | Staging migration verified 60 legacy exams remain legacy+legacy+NULL audience+published |
| Report audience | report audience is changed after competency publication | P0 | ✅ | Published exam save guard rejects audience changes; published result copies audience snapshot |
| Report audience | report layout and module list for both versions | P0 | ✅ | One `competencyReport.vue` renders both audience values |
| Report audience | overall evaluation content lookup | P0 | ✅ | temp-v1 matches exact audience + evaluation level + content version; formal customer content remains external replacement work |
| Report audience | development advice content lookup | P0 | ✅ | temp-v1 matches exact audience + dimension + level + content version; no cross-audience/version fallback |
| Report audience | result score, dimension order, and charts across versions | P0 | ✅ | SC-012 staging used identical 40/40 answers: overall/5 dimension facts and order matched; two A4 9-page PDFs had 9/9 normalized text equality and matching chart data |
| Report audience | exact text rows used by both real PDF versions | P0 | ✅ | SC-012 matched 2/2 overall and 10/10 dimension texts to exact temp-v1 audience+dimension+level rows |
| Report audience | historical report regeneration | P0 | ✅ | Report data reads `el_competency_result.report_audience` snapshot |
| Result navigation | competency exam “detail” action | P0 | ✅ | FB-075 RED→GREEN routes competency to `CompetencyResults` and preserves legacy route; deployed staging E2E queried the retained exam and clicked the primary detail action, then passed 9 operation classes and hid 9 legacy controls |
| Result navigation | competency legacy `exam/users` stale URL, bookmark or existing tab | P0 | ✅ | FB-076 component fetches exam type before loading participants and replace-routes competency to `CompetencyResults`; staging stale-URL E2E hid legacy controls 9/9 and called legacy generate-report 0 times |
| Result navigation | dashboard recent competency exam action | P0 | ✅ | FB-076 explicitly routes competency to `CompetencyResults`, preserves legacy `exam/users`; dedicated RED→GREEN source regression passed |
| Participant QR navigation | open legacy exam has a physical repo code | P0 | ✅ | Existing QR path includes the required `:repoCode` segment |
| Participant QR navigation | open competency exam has no physical repo association and `repoCode` is empty | P0 | ✅ | FB-102 resolves virtual code 00401 before building the required candidate route; exact URL unit assertion and local 8089 browser route render passed |
| Participant QR navigation | closed competency exam uses tester QR with empty optional repo code | P0 | ✅ | FB-102 uses the same resolver and URL builder; exact tester URL assertion includes `/00401` |

### F2. Competency Product / Scoring / Content / Template Versions

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Legacy save | assessment is legacy | P0 | ✅ | Four competency version fields remain empty and legacy behavior is unchanged |
| Competency draft | all four version fields are omitted | P0 | ✅ | Save resolves current product/scoring/content/template defaults before writing |
| Competency draft | a version contains an invalid identifier | P0 | ✅ | Reject before the exam transaction writes any row |
| Competency draft | a valid future content version is supplied | P0 | ✅ | Preserve the explicit content version without falling back to `temp-v1` |
| Publish | current executable product/scoring/template versions are configured | P0 | ✅ | Publish freezes all four versions with dimension/question snapshots |
| Publish | product, scoring, or template version is unsupported by the running code | P0 | ✅ | Reject before creating snapshots or changing publish status |
| Published edit | any frozen version differs from the stored version | P0 | ✅ | Reject the save while still allowing unrelated editable metadata |
| Submit | a published version set exists | P0 | ✅ | Result freezes product/scoring/content/template versions from the exam, never process constants |
| Report | result references a content and template version | P0 | ✅ | Text lookup and report instance use the exact frozen versions with no cross-version fallback |
| Compatibility migration | existing competency exams/results/reports predate version columns | P0 | ✅ | Staging executed 007 twice; 8 columns exist, all competency gaps are 0, and all 60 legacy exams retain empty versions |
| Exam API / form | competency detail or paging is loaded and switched to legacy | P1 | ✅ | Return/display all versions; switching to legacy clears all four fields |

> 2026-08-09：上述 ✅ 均为本地自动化测试/构建证据；本切片未部署 staging/production。007 迁移只有静态检查证据，必须在可用 MySQL 环境再次执行并核对回填结果。

> [纠正 - 2026-08-10] 上述 007 未部署状态已失效。`20.200.136.133` staging 已完成备份、两次幂等迁移、后端/前端部署、真实 API 拒绝测试及真实管理表单版本摘要验证；production 仍未部署。

### F3. Phase-1 Question Type / Validity / First-Level Result Structures

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Source question | existing legacy or competency row predates question type | P0 | ✅ | Nullable `competency_question_type` preserves legacy rows; existing dimension-linked competency rows backfill to `dimension` |
| Published question | snapshot predates question type | P0 | ✅ | Nullable snapshot field and compatibility backfill preserve old papers; current required dimension/direction fields remain unchanged |
| Overall result compatibility | existing result predates dimension-question counters | P0 | ✅ | Static migration contract backfills dimension counts from existing total/answered counts without recalculating scores |
| First-level group snapshot | product has grouped dimensions | P0 | ✅ | Model/migration contract stores one exam-scoped group plus nullable many-dimension links without hard-coded phase-1 names |
| First-level group result | paper has group aggregation | P0 | ✅ | Model/migration contract keeps score/level nullable and records counts/scoring version; runtime guard confirms no calculation/write was added |
| Validity result | paper has validity questions | P0 | ✅ | Model/migration contract keeps score/status nullable and records counts/scoring version; runtime guard confirms no direction/threshold/write was added |
| Historical result | old product has no group or validity data | P0 | ✅ | Migration creates no synthetic group/validity rows and scoring contract tests keep current `competency-v1` behavior |
| Full-chain delete | new group/validity rows exist | P0 | ✅ | Source-order regression test verifies result children before paper, then dimensions before referenced group snapshots, in one transaction |
| Migration rerun | 008 already applied | P0 | ✅ | Staging MySQL 8.0.46 executed 008 twice; rows, columns, indexes and foreign keys remained identical |

> 2026-08-10：上述 ✅ 已通过聚焦测试 89/89、Go 全量测试、Windows/Linux amd64 构建及 `go vet ./...`；本机无 MySQL/Docker/WSL，因此 008 尚未真实执行一次/两次，迁移重跑分支继续保持 ❌，也未部署 staging/production。

> [纠正 - 2026-08-10] 上述 008 数据库未验证状态已失效。`20.200.136.133` staging 已完成完整备份、首次迁移和第二次幂等迁移；5 个目标列、3 张表、5 个新外键、索引签名、collation、回填和孤儿检查全部通过。仅执行数据库迁移，未部署应用，production 未修改。

### F4. Phase-1 A/B Dimension Identity Reset

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Identity source | phase-1 dimension definitions are loaded | P0 | ✅ | Exactly 10 enabled identities exist in order: `A1-01` through `A1-05`, then `B1-01` through `B1-05`; names and customer core meanings match the 2026-08-07 material |
| Identity source | unresolved phase-2/3 matrix is inspected | P0 | ✅ | Migration seeds no dimension outside the confirmed phase-1 ten and makes no 34/40-dimension claim |
| Fresh installation | migration 002 initializes dimension master data | P0 | ❌ | New databases receive only the confirmed A/B phase-1 identities and no `D01-D48` row |
| Existing environment reset | old competency exams, papers, results, snapshots, reports or source questions exist | P0 | ✅ | Staging invoked the existing transactional full-chain delete for all 9 competency exams, reduced all runtime/report dependencies to zero, then migration 009 replaced 384 source questions, 392 report-text rows and 48 retired dimensions with the ten A/B identities |
| Legacy isolation | traditional 001/002/003 data exists during reset | P0 | ✅ | Ten pre-reset legacy signatures covering exams, papers, paper questions, candidates, testers, user exams, questions, answers, repository links and repositories remained byte-identical after delete, 009 and deployment |
| Report cleanup | competency report instances reference generated PDF files | P0 | ✅ | Seven referenced PDFs were backed up, removed by the application full-chain delete and individually verified absent; the competency report directory contained zero PDFs afterward |
| Temporary content | old `temp-v1` report text references the retired D identity set | P0 | ✅ | Migration 009 cleared all 392 old report-text rows and created no replacement scoring/report content |
| Reset preflight | environment authorization/write quiescence is absent, a competency exam/runtime/report dependency remains, or the database shape is unexpected | P0 | ❌ | Migration 009 requires explicit staging-only and stopped-write authorization in the same session, takes a migration lock, then aborts before deleting source/master data on any failed precondition; it never bypasses full-chain/PDF cleanup |
| Reset rerun | identity reset has already completed | P0 | ✅ | Staging first execution recorded `apply_reset=1`; the second recorded `apply_reset=0`, while the marker and full ten-row dimension signature remained unchanged |
| Candidate artifact | customer workbook is converted after identity confirmation | P0 | ✅ | Candidate JSON uses A/B dimension IDs/codes and A/B-prefixed dimension question codes, with no D identity mapping or `MAP-001` blocker |
| Dimension maintenance | administrator edits a confirmed A/B dimension | P0 | ✅ | API/UI use the two confirmed layers (`通用能力`/`心理素养`), category `基层员工`, and order range 1-10; stable ID/code remain immutable |
| Question import contract | administrator downloads or validates the current dimension-question template | P1 | ✅ | Guidance and examples use current A/B identities; validation matches a positive order to an existing dimension rather than assuming D01-D48 |

> 2026-08-10：用户确认可清除旧胜任力历史数据，并选择“本地改造 + `20.200.136.133` staging 全量重置”；传统 001/002/003 必须保留，production 不修改。本切片不导入 90 题，不实现效度方向/阈值、一级聚合或五档评分。

> 2026-08-10 本地验证：先取得Go 71通过/8失败、前端132通过/1失败和候选身份失败的RED证据；实现后聚焦Go 79/79、Go全量、前端23文件133项、Windows/Linux构建、`go vet ./...`、前端production build、候选身份与确定性检查均通过。002/009仅通过静态契约；真实MySQL首次执行、阻断分支、第二次no-op、传统摘要和PDF残留仍未验证，故对应分支保持❌。当前公网health为ok，但当前公网IP `20.239.176.250` 到staging TCP/22超时，未登录、未备份、未删除、未执行009、未部署；production未修改。

> [纠正 - 2026-08-10] 上述 staging 阻塞及009未验证状态已失效。已完成受限完整备份、Nginx停写、9个胜任力测评整链删除、7个PDF逐路径核验、009首次执行和第二次no-op、后端/前端部署、真实Nginx API及Chromium页面验收；最终为marker=1、A/B维度10、D/源题/旧文案/胜任力测评/运行依赖/PDF均0，传统十组签名不变。002全新Schema实跑和009缺授权/残留依赖的真实负向阻断尚未执行，故对应两项继续保持❌；production未修改。

### G. Competency Exam Creation Configuration

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Exam creation | assessment type is legacy | P0 | ✅ | Staging browser loaded 60 existing exams including 001/002/003 variants after migration/deploy |
| Exam creation | assessment type is competency | P0 | ✅ | Form conditionally shows report version/dimensions and hides repo controls; frontend build passes |
| Exam creation | competency selects frontline employee and one or more enabled dimensions | P0 | ✅ | Staging API created draft with 2 dimensions; detail and SQL verified |
| Exam creation | competency selects leader and one or more enabled dimensions | P0 | ✅ | Staging API edited draft to leader with 1 dimension; detail and SQL verified |
| Exam creation | competency report version is empty or unknown | P0 | ✅ | Frontend validation and backend whitelist tests pass |
| Exam creation | no dimension selected | P0 | ✅ | Frontend rules plus backend empty/nil tests pass |
| Exam creation | duplicate dimension IDs | P0 | ✅ | Covered by `TestValidateCompetencyDimensionIDs/duplicate_id` |
| Exam creation | selected dimension does not exist or is disabled | P0 | ✅ | Staging temporarily disabled D48; Save rejected and wrote zero exam rows, then D48 was restored |
| Exam creation | selected dimension has zero enabled questions | P0 | ✅ | Pure guard tests identify the first zero-count dimension by code/name before exam write |
| Exam creation | selected dimensions all have enabled questions | P0 | ✅ | Guard accepts positive counts and Save stores each count in draft association |
| Exam creation | enabled-question count query fails | P0 | ✅ | Closed-database injection test proves grouped query error propagates to the Save transaction |
| Dimension list | enabled question counts vary by dimension | P0 | ✅ | One grouped query returns each count; UI sums selected questions and disables zero-count/status-disabled dimensions |
| Exam creation | dimensions list is empty | P1 | ✅ | Selector unit test verifies explicit migration guidance empty state |
| Exam creation | edit competency draft | P0 | ✅ | Staging frontline→leader and 2→1 dimension round-trip verified |
| Exam creation | edit published competency exam | P0 | ✅ | Staging published D01 exam then rejected leader→frontline audience change; temporary exam was deleted |
| Exam creation | switch form from competency back to legacy before save | P1 | ✅ | `handleAssessmentTypeChange` clears audience/dimensions, restores legacy scoring/publish state and repo controls |
| Exam creation | save request fails | P1 | ✅ | Save uses loading with `finally`; form model is retained for retry |

### G2. Competency Exam Deletion

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Delete exam | legacy exam has participant/paper relations | P0 | ✅ | Preserve existing rejection behavior |
| Delete exam | published competency exam with full runtime chain | P0 | ✅ | Dedicated transaction deletes results, answers, paper, participants, snapshots and exam in dependency order |
| Delete exam | competency chain deletion fails midway | P0 | ✅ | Every delete error is returned to the outer transaction, causing rollback |
| Delete exam | deletion succeeds | P0 | ✅ | Staging deleted an 8-question published chain; exam/snapshots/paper/results/candidate remaining count was 0 |
| Direct delete | competency paper/candidate/tester deleted outside exam chain | P0 | ✅ | Generic physical/logical delete endpoints reject and instruct deletion through owning exam |

### H. Competency Question Metadata and Pure Scoring

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Question metadata | competency question has code, one dimension, item number, observation point, direction, and status | P0 | ✅ | Model and migration static tests expose all required fields |
| Question metadata | legacy question has NULL competency metadata | P0 | ✅ | Staging SQL verified zero existing questions with competency metadata after migration |
| Question metadata | duplicate global question code | P0 | ✅ | Staging MySQL rejected duplicate with 1062 on `uk_qu_question_code`; transaction left zero rows |
| Question metadata | duplicate dimension item number | P0 | ✅ | Staging MySQL rejected duplicate with 1062 on `uk_qu_dimension_item`; transaction left zero rows |
| Question metadata | dimension reference collation differs by environment | P0 | ✅ | Staging verified both question and dimension IDs use `utf8mb4_general_ci` |
| Question scoring | forward raw value 1 through 5 | P0 | ✅ | Final scores are 1,2,3,4,5 |
| Question scoring | reverse raw value 1 through 5 | P0 | ✅ | Final scores are 5,4,3,2,1 |
| Question scoring | raw value outside 1 through 5 | P0 | ✅ | Reject |
| Question scoring | direction empty or unknown | P0 | ✅ | Reject |
| Dimension scoring | complete dimension | P0 | ✅ | Average all question final scores |
| Dimension scoring | timeout with unanswered questions | P0 | ✅ | Average answered questions only; mark incomplete |
| Dimension scoring | zero answered questions | P0 | ✅ | Score is NULL and dimension is excluded from overall score/evaluation |
| Overall scoring | multiple valid dimensions | P0 | ✅ | Sum exact dimension averages without early rounding |
| Overall scoring | no valid dimensions | P0 | ✅ | Overall score is 0; evaluation average/level are NULL |
| Overall scoring | same answers in different display order | P0 | ✅ | Identical dimension and overall results |
| Score level | boundary values 1.00, 2.00, 3.00, 4.00, 5.00 | P0 | ✅ | Map to low/average/good/high exactly |

### I. Competency Question Import

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Import template | administrator downloads template | P0 | ✅ | HTTP test parses xlsx and verifies nine headers, four rows, and zero merged cells |
| Import template | template provides understandable sample data | P1 | ✅ | Contains two valid D01 examples covering forward and reverse scoring |
| File boundary | file missing, not xlsx, empty, or larger than 10 MiB | P0 | ✅ | HTTP tests cover missing, wrong extension, zero-byte and 10MiB+1 payloads |
| Import preview | valid rows | P0 | ✅ | Pure validation returns normalized row; source guard proves preview has no write calls |
| Import preview | header differs from the nine-column contract | P0 | ✅ | Explicit header error test passes |
| Row validation | dimension order outside 1-48 or dimension missing/disabled | P0 | ✅ | Invalid order, missing dimension, and disabled dimension tests pass |
| Row validation | dimension name differs from master data | P0 | ✅ | Exact-name mismatch test passes |
| Row validation | question code empty or duplicated in file/database | P0 | ✅ | Empty, file duplicate, and injected database duplicate tests pass |
| Row validation | dimension item number invalid or duplicated in file/database | P0 | ✅ | Invalid, file duplicate, and injected database duplicate tests pass |
| Row validation | question content or observation point empty | P0 | ✅ | Both required-field tests pass |
| Row validation | direction is not 正向/反向 | P0 | ✅ | Unknown direction test passes; values normalize to forward/reverse |
| Row validation | status is not 启用/停用 | P0 | ✅ | Unknown status test passes; values normalize to 0/1 |
| Row validation | a dimension has any question count, including outside 7-8 | P0 | ✅ | Single-row dimension validates without count warning |
| Formal import | uploaded SHA-256 differs from preview | P0 | ✅ | Real multipart HTTP test rejects changed file before DB access using constant-time comparison |
| Formal import | any validation error exists | P0 | ✅ | Validation gate precedes transaction; preview reports all row errors |
| Formal import | all rows valid | P0 | ✅ | Staging previewed and imported 384/384 rows, rerun detected complete existing set and data hash matched source |
| Formal import | database insert fails | P0 | ✅ | Staging temporary BEFORE INSERT trigger forced MySQL failure; API reported rollback, question residue=0, trigger removed |
| Question export | one or more competency questions exist | P0 | ✅ | Exports all source questions in stable dimension/item order using the same nine-column contract as import |
| Question export | no competency questions exist | P1 | ✅ | Returns a valid header-only workbook |
| Question export | database query or workbook generation fails | P1 | ✅ | Returns a controlled error before writing a partial xlsx response |

#### I1. Phase-1 90-Question Mixed Import

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Import contract | workbook uses the new ten-column contract with 题目类型 | P0 | ✅ | Template/service/export tests share ten headers; staging export returned ten headers and 90 data rows; FB-104 keeps the dialog wording on the ten-column contract |
| Row validation | 题目类型 is 维度题 | P0 | ✅ | Local service tests normalize to `dimension`; staging persisted 80 rows with dimension+type+item isolation |
| Row validation | 题目类型 is 效度题 | P0 | ✅ | Local service tests require forward validity rows; staging persisted 10 rows, one associated with each A/B dimension and excluded them from enabled dimension-question counts via FB-103 |
| Row validation | 题目类型 is empty or unknown | P0 | ✅ | Local service tests reject empty/unknown values before database writes |
| Formal import | one file contains 80 dimension and 10 validity rows | P0 | ✅ | Staging preview returned 90/0, formal import wrote 90 in one transaction, repeated preview returned 90 errors and repeated import was rejected with row count unchanged |
| Candidate conversion | all confirmed P0/P1 decisions resolve the candidate blockers | P0 | ✅ | Candidate v3 is deterministic and import-ready with zero blockers/warnings, ten forward validity rows, B1-04 all-forward and four confirmed version names |
| Candidate import workbook | generated workbook matches the resolved candidate byte-for-byte | P0 | ✅ | `--check` and identity test verify byte reproduction, 90 unique rows, 62/18 dimension directions and 10 forward validity rows; staging imported SHA-256 matched |
| Question export/UI | mixed source questions are read back after import | P1 | ✅ | Staging export content matched the imported workbook by question code; API returned 80/10; real Chromium showed both type tags, total 90 and the ten-column dialog with zero console/request errors |

#### I2. Phase-1 Fixed Product Configuration

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Draft defaults | new competency draft omits fixed audience, dimensions and four versions | P0 | ✅ | Pure profile test returns frontline employee, canonical ten A/B IDs, 20-minute default and four confirmed phase-1 versions; Save applies omitted fixed fields |
| Draft validation | client submits leader audience, another dimension set/order or another version | P0 | ✅ | Table-driven service tests reject each mutation; handler guard runs before database transaction |
| Draft inventory | all ten dimensions are enabled with exactly 8 enabled dimension rows and 1 enabled validity row each | P0 | ✅ | Grouped-query sqlmock test returns 8/1; validator accepts all ten exact inventories and draft associations retain dimension-only count 8 |
| Draft inventory | a dimension is absent/disabled or has a count/type mismatch | P0 | ✅ | Existing enabled-master length guard rejects missing/disabled dimensions; inventory matrix rejects 7/1, 8/0 and unknown-type rows before exam writes |
| Frontend profile | administrator selects competency on a new or draft form | P0 | ✅ | Vue test verifies selectors are absent, fixed profile/version text is present, fixed values are applied and duration defaults to 20 minutes |
| Frontend legacy switch | administrator switches the unsaved form back to legacy | P1 | ✅ | Existing version-form test verifies all four versions are cleared; implementation also clears audience/dimensions and restores repository data |
| Publish gate | phase-1 scoring/group/validity runtime is complete | P0 | ✅ | Focused RED proved the old service/UI gate; unified runtime tests now require backend readiness and an enabled Vue publish action |

#### I3. Phase-1 Five-Level Scoring Engine

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Question score | dimension question is forward or reverse with raw value 1–5 | P0 | ✅ | Existing exhaustive 1–5 forward/reverse table remains GREEN; invalid values and directions are rejected |
| Dimension score | each canonical A/B dimension has exactly 8 answered dimension questions | P0 | ✅ | Phase-1 pure engine requires 80 rows and computes each exact `scoreSum/8` rational without early rounding |
| Dimension score | any canonical dimension is missing, has not exactly 8 rows, contains a non-dimension type or has unanswered rows | P0 | ✅ | Missing/mixed/unknown identity inputs are rejected; trusted incomplete rows return counts but nil formal dimension/overall scores and no levels |
| Dimension level | exact score is at 1.7/2.7/3.5/4.3 boundaries or immediately above | P0 | ✅ | Boundary table covers exact and +0.01 values with upper-inclusive L1–L5 rational comparisons |
| Overall score | all ten canonical dimensions are complete | P0 | ✅ | Ten exact dimension averages sum to an exact rational in 10–50; test fixture produces 30 without evaluation-average reuse |
| Overall level | exact total crosses 25/32.5/40/45 boundaries | P0 | ✅ | Boundary table covers below/exact values for not-qualified/weak/qualified/good/excellent without percentage conversion |
| Input identity | dimensions are reordered but identities/order metadata are valid | P1 | ✅ | Reverse input order still emits canonical A/B order and deterministic total/level |
| Runtime integration | submit/publish flow uses phase-1 scoring result | P0 | ✅ | Staging真实90题链持久化十维均为3/L3，总体30/weak，重复提交不重复写入 |

#### I4. Phase-1 First-Level Group Aggregation

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Group identity | canonical ten A/B dimension results are provided in any order | P0 | ✅ | Reverse-order test emits exactly `general_ability/通用能力` then `psychological_quality/心理素养`, each with its fixed five child IDs |
| Group score | all five child dimensions are complete | P0 | ✅ | Both groups average five exact dimension rationals to 3 without early rounding and classify as L3 |
| Group score | one or more child dimensions are incomplete | P0 | ✅ | General group retains 5/4 dimensions and 40/39 question counts with nil score/level; unaffected psychological group remains complete with exact score 4 |
| Group level | exact group average is at 1.7/2.7/3.5/4.3 boundaries | P0 | ✅ | Boundary matrix reuses the exact phase-1 dimension classifier for both groups |
| Input integrity | child dimension is missing, duplicated, unknown, has invalid order or a complete row has nil score | P0 | ✅ | Malformed-input matrix rejects all five cases; score/level completeness consistency is also checked |
| Runtime integration | publish freezes two group snapshots and submit persists two group results | P0 | ✅ | Staging真实发布冻结2组并绑定10维，提交写2条3/L3一级结果，清理后零残留 |

#### I5. Phase-1 Validity Calculation

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Validity input | exactly 10 validity questions are all answered with raw values 1–5 | P0 | ✅ | Pure function sums original raw values only, requires forward metadata and returns complete integer score/status |
| Validity boundary | complete score is exactly 35 or 36 | P0 | ✅ | Explicit fixtures verify 35=`good`, 36=`questionable` with exact integer scores |
| Validity incomplete | one or more validity questions are unanswered | P0 | ✅ | 9/10 fixture preserves counts, keeps score nil and returns `incomplete` with `IsComplete=false` |
| Validity extremes | complete answers sum to 10 or 50 | P1 | ✅ | All-1 and all-5 fixtures classify good/questionable respectively |
| Input isolation | count is not 10, type is not validity, raw is outside 1–5, direction is not forward or identity/order is invalid | P0 | ✅ | Malformed matrix rejects count/type/raw low/raw high/reverse/blank/duplicate/order cases |
| Score independence | validity result is calculated alongside dimension/group results | P0 | ✅ | API accepts only `[]Phase1ValidityInput` and returns a separate value; no dimension/group object is consumed or mutated |
| Runtime integration | submit persists one validity result and management can filter good/questionable/incomplete | P0 | ✅ | Staging真实提交写10/good，good筛选1、questionable筛选0；管理详情与扩展导出已验证 |

#### I6. Phase-1 Unified 90-Question Runtime

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Publish | fixed draft has exactly 80 dimension and 10 forward validity questions | P0 | ✅ | Staging冻结2组/10维/90题，题型80/10，五个确认选项逐题匹配，重复发布幂等 |
| Publish | inventory/type/direction/group metadata is malformed or snapshot insert fails | P0 | ✅ | Staging通过临时BEFORE INSERT触发器强制题目快照写入失败；事务回滚后publish=0、group=0、question=0、group links=0，移除触发器后同一草稿可正常发布90题 |
| Answer | dimension question is forward/reverse; validity question is forward | P0 | ✅ | 本地正反向穷举保持GREEN；staging真实混合90题全部保存，效度raw=1并得到总分10 |
| Submit | all 90 questions are answered | P0 | ✅ | Staging数据库与API均验证10+2+1+1结果，total=90、dimension=80，完整性为1 |
| Submit | timeout leaves a dimension and a validity row unanswered | P0 | ✅ | Staging真实88/90：维度79/80、overall NULL、1条二级NULL、1条一级NULL、效度9/10+incomplete+NULL；默认排名0、正式报告拒绝 |
| Submit | request is repeated after successful commit | P0 | ✅ | Existing idempotent result lookup returns the frozen result without duplicate rows |
| Management | validity status is good, questionable or incomplete | P0 | ✅ | 管理筛选API真实验证good/questionable；前端139项测试覆盖状态、一级和阈值详情 |
| Statistics | score ranking sees an incomplete or validity-questionable result | P0 | ✅ | Staging分别验证timeout和40/questionable答卷均不进入默认overallScore排名；显式all和questionable筛选仍返回存疑答卷。独立常模/汇总统计端点尚未建设，继续作为产品增强 |
| Report | complete validity is good or questionable | P0 | ❌ | Formal report remains allowed for both but must display the confirmed good/questionable notice; content text is not yet configured |

#### I7. Phase-1 Formal Report Framework

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Version dispatch | frozen versions are generic or phase-1 | P0 | ✅ | Generic keeps the existing dynamic four-level renderer; phase-1 selects an isolated fixed report-data contract without reusing evaluationAverage |
| Approval gate | phase-1 content package is missing, draft, retired or lacks either approval/hash/disclaimer | P0 | ✅ | Reject before report instance creation, Chromium rendering, PDF write or success audit with a stable not-approved error |
| Content completeness | approved package lacks overall/group/dimension/validity text, contains whitespace-only text, or has inconsistent disclaimers | P0 | ✅ | Current overall, 2 group descriptions, 10 current L1-L5 dimension texts, current validity notice and one consistent final disclaimer are required; FB-107 RED reproduced three gaps and GREEN rejects them without cross-version/audience fallback |
| Report data | complete good/questionable phase-1 result has 2 groups, 10 dimensions and validity | P0 | ✅ | Build a strong `competency-phase1-report-data-v1` DTO with `/50` overall, independent `/5` groups, ten dimensions, versions and participant-visible fields |
| Ten-page layout | phase-1 report framework is selected | P0 | ✅ | Fixed pages: cover, guide, person/overall/validity, groups, 10-axis radar, then five two-dimension detail pages |
| Validity privacy | participant report renders validity | P0 | ✅ | Show good/questionable notice but never expose raw validity score or the 35/36 threshold; management detail/export remain unchanged |
| Current candidate state | formal approvals/content source are absent | P0 | ✅ | Framework exists but `reportAvailable=false`; view/generate/download/batch actions remain disabled and direct backend requests are rejected |
| Dimension level labels | report renders L1-L5 for group or dimension results | P0 | ✅ | CSV catalog exposes separate secondary labels 差/较差/合格/较优秀/优秀 and group labels 低分/较低分/中分/较高分/高分; Vue uses separate mappings |
| Overall level labels | report renders excellent/good/qualified/weak/not_qualified | P0 | ✅ | Vue resolves formal CSV labels 优秀胜任/良好胜任/合格胜任/薄弱胜任/尚未胜任 from the generated catalog |
| CSV report catalog | phase-1 template renders names, definitions and labels | P0 | ✅ | Deterministic generator validates CSV first and emits the runtime catalog; freshness check prevents stale catalog use |
| Customer sample layout | phase-1 formal template renders ten A4 pages | P0 | ✅ | Chromium and pdfinfo verify cover, guide, personal/overall/validity, groups, ten-axis radar and five two-dimension pages as exactly 10 A4 physical pages |
| Missing formal text | validity-good notice or final disclaimer is absent | P0 | ✅ | Existing dual-approval/content-completeness gate continues to reject report data before Vue/PDF rendering; template does not invent fallback text |
| Customer DOCX fixed text | cover and reading-guide fixed copy is rendered | P0 | ✅ | Seven reviewed CSV rows drive report title, English title, slogan, three reading paragraphs and special notice; staging Chromium verifies every fixed excerpt |

#### I8. Phase-1 Customer Workbook Conversion

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Dimension identity cell | V1 workbook supplies `code + newline + name` in question and level sheets | P0 | ✅ | FB-108 parses both values, requires an exact A/B code-name pair, and preserves 90 questions plus 10×5 report texts |
| Dimension identity mismatch | supplied code and name belong to different confirmed dimensions | P0 | ✅ | FB-108 rejects with a deterministic code/name mismatch error before generating candidate or import artifacts |
| Validity classification | validity rows have an empty question-type cell under the validity layer | P0 | ✅ | Continue classifying all 10 rows from the explicit validity layer and force forward scoring metadata |
| Normalized CSV export | confirmed V1 workbook is exported for manual review | P0 | ✅ | Produces six UTF-8 BOM CSV files: package, 90 questions, 10 dimensions, 50 dimension-level texts, 5 overall texts and 4 static report texts |
| CSV-only mode | reviewer requests CSV artifacts without replacing candidate JSON/XLSX | P0 | ✅ | Writes only the requested CSV package; existing candidate JSON/XLSX hashes remain separately verifiable |
| Long-term import template | CSV package is retained as the future question/report source | P0 | ✅ | Includes package metadata, stable order, review workflow fields, machine level codes, exact 1/0 score boundaries and formula-injection rejection |
| Static report content | phase-1 report requires template/group/validity text | P0 | ✅ | Exports seven customer-DOCX template rows, two group rows and two validity rows; validity-good remains an explicitly labeled AI suggestion with `pending_review` |
| Approval metadata | CSV package is reviewed but dual approval is incomplete | P0 | ✅ | Approval stays draft with explicit blank approver/time/environment/disclaimer fields; approved-import validation rejects it until all fields and content SHA match |
| Candidate database import | validated draft CSV is loaded into staging for simulation | P0 | ✅ | Imported exactly 66 inactive temporary report-text rows and one draft package using deterministic IDs in one transaction; formal report gate remains closed |
| Candidate import rerun | the same CSV package is imported again | P0 | ✅ | Transaction SQL ran twice and retained 66 unique lookup rows plus one draft package |
| Candidate source parity | staging already contains the 90 questions and 10 dimensions | P0 | ✅ | Pre-import comparison matched all question code/type/dimension/item/content/observation/direction/status and dimension id/code/name/order/status fields |

### J. Competency Publish Snapshot

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Publish | exam missing, legacy, or invalid mode | P0 | ✅ | Locked exam is strictly dispatched; invalid mode exits transaction |
| Publish | competency draft has no dimensions | P0 | ✅ | Transaction rejects empty selection |
| Publish | selected dimension disabled or has zero enabled questions | P0 | ✅ | Revalidated in publish transaction with dimension-specific zero count error |
| Publish | source question metadata invalid | P0 | ✅ | Code/content/direction/dimension metadata validated before batch insert |
| Publish | valid draft | P0 | ✅ | Staging froze 2 dimensions, 4 questions, option JSON and publish audit in one transaction |
| Publish | dimension master changes after draft save but before publish | P0 | ✅ | Publish refreshes code/name/VIRD/category/core meaning/order from current enabled master data |
| Publish | repeated request after success | P0 | ✅ | Staging repeated publish returned existing 4-question summary without rebuilding |
| Publish | source question changes after publication | P0 | ✅ | Staging mutated D01-Q01 content/observation/direction/status after publish; snapshot and historical paper content remained byte-identical, then source restored |

### K. Competency Participant Paper and Answering

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Participant auth | valid candidate/tester participant token | P0 | ✅ | Candidate staging chain issued and accepted bound participant/paper tokens; tester shares same issuer |
| Participant auth | missing, expired, wrong-purpose or mismatched token | P0 | ✅ | Token unit matrix plus Handler binding checks reject before runtime service |
| Create paper | exam not published or has no snapshots | P0 | ✅ | Runtime requires publish status and nonempty snapshot set before insert |
| Start assessment UI | competency exam is still draft/unpublished | P0 | ✅ | Online list hides drafts; direct preparation URL disables start with Chinese publish guidance; backend maps sentinel error to Chinese |
| Create paper | first entry | P0 | ✅ | Staging verified complete four-question unique set and fixed 1-N order using crypto Fisher–Yates |
| Create paper | repeated entry with in-progress paper | P0 | ✅ | Staging returned the same paper ID and a newly signed paper token |
| Create paper | participant already completed | P0 | ✅ | Locked participant/paper state returns completed instead of creating another paper |
| Create paper | secure random source fails | P0 | ✅ | Injectable random-source unit test returns error; shuffle occurs before paper insert |
| Paper detail | valid paper token | P0 | ✅ | Staging repeated detail order matched; response hid dimension, direction and final score |
| Fill answer | raw value 1-5 on owned in-progress unexpired paper | P0 | ✅ | Staging saved four values and returned answered counts only |
| Fill answer | invalid value, foreign question, finished or expired paper | P0 | ✅ | Staging rejected raw=0, foreign paper question and finished writes; expired write triggered trusted timeout submit |
| Exam timing | competency total time is zero or negative | P0 | ✅ | Save/publish reject; no hidden default duration |
| Create paper | current time is after exam end time | P0 | ✅ | New paper path rejects before snapshot read, shuffle or insert |
| Create paper | paper started before exam end time and personal duration remains | P0 | ✅ | Existing paper restore path precedes end-time start guard and preserves personal limit time |

### L. Competency Submit, Results, and Report Data

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Manual submit | unanswered question exists | P0 | ✅ | Staging rejected manual incomplete submit with first unanswered display order |
| Manual submit | all questions answered | P0 | ✅ | Staging atomically wrote one overall and two ordered dimension results |
| Timeout submit | partially answered or zero answered dimensions | P0 | ✅ | Staging Worker persisted partial 1/16 and zero 0/16 timeout results with nil zero-answer dimensions and no minimum-score fill |
| Timeout submit | participant claims timeout before personal limit | P0 | ✅ | HTTP rejects client timeout; service validates locked paper limit before timeout semantics |
| Submit | repeated/concurrent after completion | P0 | ✅ | Staging repeated submit returned `alreadySubmitted`; PK/unique indexes prevent duplicate result sets |
| Submit | successful completion | P0 | ✅ | Staging updated paper/participant and participant response contained no scores |
| Admin results | paging/detail with administrator JWT | P0 | ✅ | Staging detail and report data returned saved overall, two dimensions and four audit rows |
| Participant results | anonymous participant endpoint | P0 | ✅ | Router exposes no participant result route |
| Report data | frontline/leader audience snapshot | P0 | ✅ | Staging report used saved `leader` audience and the same saved score facts |
| Report data | formal text missing | P0 | ✅ | Report returns `reportTextReady=false` and explicit pending marker; no cross-audience fallback |
| Formal report | result is incomplete | P0 | ✅ | Formal report data validates `is_complete=1`; admin result detail remains available |

### M. Competency Result Management

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Result paging | sort by submitted time | P0 | ✅ | Staging API/browser returned High→Mixed→Low for fixed descending submission times; stable `paper_id` tie-breaker retained |
| Result paging | sort by overall score ascending/descending | P0 | ✅ | Staging API/browser ascending returned 2→6→8; whitelist rejects user SQL fragments |
| Result paging | sort by selected dimension ascending/descending | P0 | ✅ | Staging D01 descending returned 5→4→1; parameterized JOIN and measured-dimension ownership check active |
| Result paging | dimension sort requested without dimension ID | P0 | ✅ | Pure validation test rejects before query |
| Result paging | unknown sort field/direction | P0 | ✅ | Pure validation test rejects both cases before query |
| Result paging | participant is candidate or tester | P0 | ✅ | Staging returned candidate name/telephone/type for all three rows through one LEFT JOIN query without N+1 |
| Result paging | no results | P1 | ✅ | Non-nil result slice initialized; frontend renders explicit empty state |
| Result detail | open from result list | P0 | ✅ | Staging browser dialog showed overall=8, two ordered dimensions and all 16 per-question audit rows |
| Result authorization | authenticated user lacks administrator/exam:list/exam:export permission | P0 | ✅ | FB-078 returns HTTP 403 before binding/querying; administrator, global, exam:list and exam:export retain access; internal-token rendering remains independently authenticated |
| Result UI | legacy exam row | P0 | ✅ | Existing test-record/export/statistics entries remain in the non-competency branch |
| Result UI | competency exam row | P0 | ✅ | Dedicated `CompetencyResults` route and list command preserve `examId` context |
| Result UI | complete competency result opens test report | P0 | ✅ | Named route receives `params.paperId`; FB-067 RED→GREEN verifies the report page can read `$route.params.paperId` |
| Result UI | management list shows start/completion/duration and detail shows score sum | P1 | ✅ | Result paging projects `p.create_time AS started_at` and `user_time`; UI shows start/completion/duration and dimension `scoreSum`; RED→GREEN tests pass |
| Result filtering | name/telephone/completion filters are empty | P1 | ✅ | Empty/all normalization test keeps filters nil/empty and existing stable sort/pagination |
| Result filtering | name or telephone contains text | P1 | ✅ | Inputs are trimmed and applied through parameterized LIKE to frozen participant snapshots; count and rows share the helper |
| Result filtering | completion is complete/incomplete | P1 | ✅ | Pure test covers all/complete/incomplete and rejects unknown values; service maps to `is_complete=1/0` |
| Result paging | malformed JSON or wrong field type | P1 | ✅ | FB-083 checks `ShouldBindJSON` and returns “参数格式错误” before defaults or service queries |
| Result UI | query and reset buttons | P1 | ✅ | Component test verifies reset clears identity/completion/sorting, returns to page 1 and reloads |
| Result UI | overlapping list/detail requests return out of order | P1 | ✅ | FB-081 uses independent monotonic sequences for list/detail; only the latest response may update data/loading, and list requests receive a frozen query snapshot |
| Result UI | no complete row selected for batch report action | P1 | ✅ | Buttons bind disabled state to selected complete rows; `selectable` rejects incomplete rows |
| Result UI | complete phase-1 row exists after the ten-page renderer is implemented | P0 | ✅ | FB-113: complete rows are selectable; generation reaches the backend dual-approval gate, and the batch summary preserves its actionable business error |
| Phase-1 PDF | approved content package supplies a final disclaimer | P0 | ✅ | FB-114: template renders the frozen approved disclaimer and only falls back to the CSV sample notice when no approved disclaimer is present; real staging PDF text verified |
| Phase-1 PDF | report is compared with the authoritative customer DOCX/PDF | P0 | ✅ | FB-115: uses the original DOCX cover illustration and customer-style cover, matrix, overall orbit, group table/pie, centered radar, flowing dimension pages, typography and print density; final staging PDF is A4 10 pages and passed equal-DPI page review |
| Phase-1 PDF | customer uploads or adjusts the approved Word template | P0 | ✅ | FB-116/117: 49 hidden content controls and 12 native charts are customer-maintainable; local LibreOffice is enabled by default, Graph optional, Chromium fallback configured; staging real generation is GREEN |
| Phase-1 Word template | customer opens the maintainable DOCX in Microsoft Word | P0 | ✅ | UF-007/FB-117: normal sample values are visible, stable field keys remain in hidden content-control tags, and the final contract scan reports 49 unique tags with zero visible `{{...}}` tokens |
| Phase-1 template management | administrator opens the existing report-template page | P1 | ✅ | Staging `/qu/template` loads current name, size, time, SHA-256 and 49/12/0 contract above the retained MBTI table; desktop 1440×900 and mobile 390×844 browser checks pass with zero overflow |
| Phase-1 template management | administrator downloads the active template | P1 | ✅ | Real API returns configured DOCX with correct MIME and exact SHA; browser button sends the authenticated download request; component Blob/file-name test passes |
| Phase-1 template management | administrator uploads a valid DOCX | P0 | ✅ | Real API and browser upload the active DOCX, create one new 0600 backup, atomically replace it, refresh metadata and preserve exact SHA; template directory is limited to service user `liming:liming 750` |
| Phase-1 template management | upload is malformed, incomplete, duplicated, oversized or unauthorized | P0 | ✅ | Real duplicate-Tag upload is rejected while active SHA stays unchanged; unit tests cover missing/duplicate contract and administrator/global-vs-exam:list permissions; UI preserves selected file after errors |
| Phase-1 exam configuration | first save uses a custom subset of required participant fields | P0 | ✅ | UF-009/FB-118 staging GREEN: type switch initializes six defaults once; save/detail do not reapply defaults. Real create→first save→Detail→candidate page preserves and renders exactly name/gender/telephone |
| Phase-1 pre-exam page | participant reviews the assessment description before starting | P1 | ✅ | Shows the confirmed behavior/tendency guidance, two numbered answer rules, confidentiality statement and fixed red notice that all 90 questions are required; does not derive 80 from dimension-only counts |
| Phase-1 PDF | approved Word template is converted by the MBTI-style local LibreOffice path | P0 | ✅ | Staging LibreOffice 24.2 generated existing complete paper as exactly 10 A4 pages; producer, required text, unresolved tags, API/file/DB hash and cleanup all passed |
| Phase-1 PDF | local LibreOffice is unavailable, times out or returns a non-PDF file | P0 | ✅ | Unit tests reject queue cancellation, command failure and non-PDF without exposing command output, and clean temporary workspaces; Chromium fallback remains enabled and its prior real E2E is GREEN |
| Result UI | one or more complete rows selected for batch generation | P1 | ✅ | Component test verifies one generation call per selected paper, loading cleanup and summary feedback |
| Result UI | one or more complete rows selected for batch download | P1 | ✅ | Component test verifies one download/save per selected paper, loading cleanup and participant filenames |
| Result UI | selection changes while batch generation/download is running | P1 | ✅ | FB-082 captures a filtered/shallow-copied complete-result snapshot at task start; loops, progress and totals never read live selectedRows afterward |
| Result UI | row view/answer-detail/download actions | P1 | ✅ | Source/component tests verify legacy labels/order while retaining competency report, detail and download APIs |
| Result UI | mobile viewport under 768px | P2 | ✅ | FB-087 wraps heading/toolbars, makes filters full width, preserves table scrolling and opens a full-screen one-column detail dialog |

### M2. Temporary Competency Formal Report and PDF

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Report template | cover identifies exam, participant, audience and generation date | P1 | ✅ | Staging 00401 PDF cover shows exam, participant, frozen audience and actual render date |
| Report template | reading guide explains 1–5 scale, aggregation and comparison boundary | P1 | ✅ | PDF explains 1.00–5.00, dimension average, overall sum/evaluation mean and cross-combination boundary |
| Report template | participant fields follow exam requiredFields | P1 | ✅ | Template filters frozen identity fields by exam requiredFields; empty configuration shows all |
| Report template | overview and measured dimensions have printable score charts | P1 | ✅ | Staging PDF renders four-level overall scale and five measured-dimension CSS bars without external chart runtime |
| Report template | each measured dimension shows frozen core meaning and temporary interpretation | P1 | ✅ | Nine-page 00401 PDF contains all five published core meanings, exact temp-v1 interpretations and development prompts |
| Report template | six configured participant fields fit on the overview page | P1 | ✅ | FB-073 RED→GREEN; five retained staging reports stay at 9 pages without splitting overall interpretation |
| Report template | score bands 1–5 and both audiences render consistently | P1 | ✅ | Five independent configs cover averages 1/2/3/4/5, overall 5/10/15/20/25, three frontline and two leader PDFs |
| Report text | active temporary content version exists for audience, overall level and every measured dimension level | P0 | ✅ | 392 temp-v1 rows cover 2 audiences × overall/dimensions × 4 levels; exact matches are frozen into instance snapshot |
| Report text | any required text is missing or belongs to another audience/version | P0 | ✅ | Pure service test rejects missing dimension level and cross-audience fallback before rendering |
| Report text | required exact lookup key is missing | P1 | ✅ | Error identifies contentVersion, audience, dimension (or overall), and level |
| Report text | temporary content is rendered | P0 | ✅ | Staging PDF text contains “临时测试报告” and “不可作为人才决策依据” |
| Internal report authentication | token is supplied in URL/query instead of `X-Internal-Token` | P0 | ✅ | FB-079 sends only `paperId` in query, places the token in `X-Internal-Token`, and rejects even a correct query token with HTTP 401 |
| Generate report | result is incomplete | P0 | ✅ | Existing FB-047 formal report guard rejects before instance/PDF writes |
| Generate report | complete result has no prior instance | P0 | ✅ | Staging created instance, rendered Chromium PDF, persisted path/hash/size and completed status |
| Generate report | same version already completed and force=false | P0 | ✅ | Unique paper+version and existing-file guard return the same instance without duplication |
| Generate report | concurrent requests target the same paperId | P0 | ✅ | FB-080 uses a stable bounded 64-stripe lock on the singleton report handler and locks before instance lookup through final audit/read, so waiting force=false requests re-read and reuse completed output |
| Regenerate report | force=true for same content version | P0 | ✅ | Same instance is refreshed and regenerate audit action is recorded by the dedicated branch |
| Generate report | render or file write fails | P0 | ✅ | Handler marks instance failed and never sets participant pdf_flag success |
| Generate report | success audit insert fails after PDF metadata update | P0 | ✅ | FB-084 inserts success audit through the same GORM transaction as report metadata and participant PDF state; rollback removes the new file |
| Download report | completed instance path is inside configured upload root and file exists | P0 | ✅ | Staging downloaded application/pdf, SHA-256 matched instance, and audit count was 1 |
| Download report | same-name participants or spaces/plus signs in response filename | P1 | ✅ | FB-085 includes paperId in frontend names and encodes server filename* with `%20`/`%2B` RFC5987-compatible percent encoding |
| Download report | missing instance/file or path escapes upload root | P0 | ✅ | `filepath.Rel` allow-root guard and completed-instance lookup reject invalid paths |
| Download report | HTTP 200 response body is a JSON business error rather than a PDF | P0 | ✅ | FB-077 rejects every non-`application/pdf` Blob, parses the backend message with Blob.text/FileReader fallback, and prevents `saveAs`/success counting |
| Delete exam | competency report instances/audits/files exist | P0 | ✅ | Staging full-chain cleanup removed report audit/instance before paper and removed allowed-root PDF; remaining=0 |

### N. Legacy and Competency Question Management Isolation

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Legacy question paging/list/export | source question has `dimension_id IS NULL` | P0 | ✅ | Traditional endpoints retain legacy questions only |
| Legacy question paging/list/export | competency source question has non-NULL `dimension_id` | P0 | ✅ | Excluded before count/list/export by parameter-free SQL predicate |
| Legacy question detail | competency source question ID supplied | P0 | ✅ | Legacy detail query includes `dimension_id IS NULL` and returns not found |
| Legacy question save | new request contains competency metadata | P0 | ✅ | Metadata presence is rejected before validation/transaction |
| Legacy question save | existing competency source question ID supplied | P0 | ✅ | ID guard runs before transaction and preserves all metadata |
| Legacy question delete | any selected ID is a competency source question | P0 | ✅ | Whole batch rejected before transaction |
| Legacy repo batch action | any selected ID is a competency source question | P0 | ✅ | Rejected before deleting/rebuilding associations |
| Legacy question UI | row has nil/empty answer list | P0 | ✅ | Safe formatter renders `—`; direct option indexing removed |

### O. Legacy Question Form Repository Selection

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Question create | one repository selected through `repoId` | P0 | ✅ | Bound field validates and synchronizes `repoIds=[repoId]` before validation |
| Question create | no repository selected | P0 | ✅ | Bound `repoId` rule shows “必须选择一个题库！” and synchronization produces an empty array |
| Question edit | repository selection unchanged | P0 | ✅ | Synchronization preserves one matching `repoIds` value |
| Question edit | repository changed from old to new | P0 | ✅ | Stale `repoIds` is always overwritten by the current selection before validation/save |

### P. Legacy Question Import Upload Routing

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Import upload | development build | P0 | ✅ | Uses `VUE_APP_BASE_API`; development resolves to `/dev-api` |
| Import upload | staging build | P0 | ✅ | Uses `VUE_APP_BASE_API`; staging resolves to `/stage-api` |
| Import upload | production build | P0 | ✅ | Production artifact contains `/prod-api/exam/api/qu/qu/import-excel` and zero `/dev-api` matches |
| Import upload | authenticated administrator | P0 | ✅ | Upload headers include current `Authorization: Bearer <token>` |

### Q. Legacy Question Paging Performance

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Question paging | page has no records | P1 | ✅ | Returns `records=[]`; loader exits before relation queries |
| Question paging | page has one or more records | P1 | ✅ | All answers loaded by one ordered `qu_id IN ?` query |
| Question paging | question has one or more repository relations | P1 | ✅ | All ordered relations loaded once; first `(sort,id)` relation retained |
| Question paging | repository relation points to deleted repository | P1 | ✅ | `COALESCE` preserves `[已删题库:<id>]` marker behavior |
| Question paging | relation query fails | P1 | ✅ | Query error returns “查询题目关联数据失败” without partial rows |

### R. Legacy Repository Batch Association

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Batch request | duplicate/blank question or repository IDs | P0 | ✅ | IDs normalized before database work |
| Batch add | any question does not exist or is competency | P0 | ✅ | Guard plus transactional legacy query rejects entire batch before writes |
| Batch add | valid questions and repositories | P0 | ✅ | Replaces selected questions' associations atomically with `CreateInBatches` |
| Batch remove | valid question/repository pairs | P0 | ✅ | Deletes selected pairs atomically |
| Batch reorder | old and requested repositories affected | P0 | ✅ | Reorders every affected repository by `(sort,id)` |
| Batch statistics | association set changes | P0 | ✅ | Refreshes every old/requested repository in the same transaction |
| Batch operation | any read/write/reorder/stat refresh fails | P0 | ✅ | Every error returns from the transaction and rolls back the batch |

### S. Legacy Question Excel Import Atomicity

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Import repository | sheet repository exists | P0 | ✅ | Reused inside the import transaction |
| Import repository | sheet repository does not exist | P0 | ✅ | Created inside the same transaction as questions |
| Import extra repositories | all referenced IDs exist | P0 | ✅ | All IDs validated, every relation insert error checked |
| Import extra repositories | any referenced ID does not exist | P0 | ✅ | Rejected before question insertion and transaction rolled back |
| Import questions | any question/answer/relation insert fails | P0 | ✅ | Repository and all imported rows roll back together |
| Import statistics | import succeeds | P0 | ✅ | Every sheet/extra repository refreshed before commit |

### T. Legacy Question Export Performance

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Question export | no matching questions | P1 | ✅ | Generates header-only workbook; loader exits before relation queries |
| Question export | one or more matching questions | P1 | ✅ | Reuses two batch relation queries for all exported questions |
| Question export | relation query fails | P1 | ✅ | Returns controlled error before writing workbook rows |
| Question export | question has multiple repositories | P1 | ✅ | Preserves all repository IDs in `(sort,id)` order |

### U. Legacy Question Deletion Integrity

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Question delete | selected question belongs to competency | P0 | ✅ | Existing dedicated-API guard rejects the whole batch |
| Question delete | selected traditional question is referenced by a paper | P0 | ✅ | Rejects before opening delete transaction and reports reference count |
| Question delete | selected traditional questions are unreferenced | P0 | ✅ | Deletes answers/relations before questions in one transaction |
| Question delete | any child delete or statistic refresh fails | P0 | ✅ | Returns error and rolls back all selected questions |

### V. Legacy Referenced Question Immutability

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Question save | new traditional question | P0 | ✅ | No historical reference check required |
| Question save | existing unreferenced traditional question | P0 | ✅ | Reference guard passes and existing transaction remains available |
| Question save | existing question referenced by any paper | P0 | ✅ | Rejected before answer/repository replacement |
| Question save | paper reference lookup fails | P0 | ✅ | Returns controlled error and does not open write transaction |

### W. Legacy Repository Paging Boundary

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Repository paging | size is zero/negative | P1 | ✅ | Normalized by shared `capPageSize` |
| Repository paging | size exceeds global maximum | P1 | ✅ | Capped before OFFSET/LIMIT query |
| Repository paging | count query fails | P1 | ✅ | Returns “查询题库总数失败” |
| Repository paging | list query fails | P1 | ✅ | Returns “查询题库列表失败” without partial records |

### X. Legacy Question Save Consistency

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Question save | repository IDs contain blanks/duplicates | P0 | ✅ | Normalized before required validation and writes |
| Question save | any repository ID does not exist | P0 | ✅ | Rejected inside transaction before question insert/update |
| Question edit | original question lookup fails | P0 | ✅ | Returns error; never replaces create time or continues |
| Question edit | old association lookup/delete fails | P0 | ✅ | Returns error and rolls back question/answers/relations |
| Question save | all repositories and child rows valid | P0 | ✅ | Commits question, answers, relations and statistics atomically |

### Y. Legacy Repository Save Integrity

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Repository create | title is blank | P0 | ✅ | Trimmed and rejected before database write |
| Repository create | valid title | P0 | ✅ | Server initializes zero counts and both timestamps |
| Repository update | client sends count/create-time fields | P0 | ✅ | Only code/title/remark/update_time are updated; row is reloaded for response |
| Repository update | repository ID does not exist | P0 | ✅ | RowsAffected=0 returns “题库不存在” |

### Z. Legacy Question and Repository Read Reliability

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Question list | query succeeds or returns no rows | P1 | ✅ | Returns records/non-nil empty array |
| Question list | query fails | P1 | ✅ | Returns “查询题目列表失败” |
| Question detail | answer/repository/code relation query fails | P1 | ✅ | Returns specific controlled error, never partial detail |
| Repository list | query succeeds or returns no rows | P1 | ✅ | Returns records/non-nil empty array |
| Repository list | query fails | P1 | ✅ | Returns “查询题库列表失败” |

### AA. Question Bank Staging Acceptance

| Gate | Priority | Coverage | Evidence |
|------|----------|----------|----------|
| Legacy/competency source isolation | P0 | ✅ | Staging traditional paging total=855; first page competency rows=0; browser table renders without option-index error |
| Question form repository synchronization | P0 | ✅ | Staging browser validates current repoId and overwrites stale repoIds |
| Environment-aware authenticated upload | P0 | ✅ | Browser upload URL is `/prod-api/.../import-excel` with Bearer header |
| Repository batch add/remove atomicity | P0 | ✅ | Duplicate IDs normalize; invalid repository batch rolls back and preserves existing association |
| Excel import atomicity | P0 | ✅ | Valid workbook imports; invalid extra repository rolls back sheet repository and question |
| Referenced/competency question guards | P0 | ✅ | Historical edit/delete and competency delete rejected; SQL counts unchanged |
| Cleanup and integrity | P0 | ✅ | Temporary rows=0, competency orphans=0, critical logs=0 |

### AB. Competency Question Bank Entry 00401

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Repository paging | administrator opens repository list | P0 | ✅ | Adds virtual code `00401`, title “胜任力测验题库”, and live competency question count |
| Repository paging | title filter matches/does not match 00401 | P1 | ✅ | Includes virtual row only when code/title filter matches |
| Repository action | administrator opens 00401 | P0 | ✅ | Routes to dedicated read-only competency question page, never traditional editor |
| Competency question paging | no filters | P0 | ✅ | Returns only `dimension_id IS NOT NULL` rows with dimension fields |
| Competency question paging | dimension/status/code/content filters | P0 | ✅ | Applies parameterized filters and stable dimension/item ordering |
| Competency question UI | 384 rows exist | P0 | ✅ | Paginated table displays code, dimension, content, observation point, direction, status |
| Virtual repository mutation | edit/delete/batch selected | P0 | ✅ | UI disables selection/edit semantics; backend rejects traditional delete for virtual ID |

### AC. Competency Question Bank 00401 Staging Acceptance

| Gate | Priority | Coverage | Evidence |
|------|----------|----------|----------|
| Repository list visibility | P0 | ✅ | Browser first row `00401 / 胜任力测验题库 / 384`, total repositories=7 |
| Dedicated navigation | P0 | ✅ | Click opens `/#/exam/competency/questions`; virtual row checkbox disabled |
| Dedicated question data | P0 | ✅ | API/browser total=384, first page=20, first row D01-Q01 with complete metadata |
| Filter correctness | P0 | ✅ | API D01 + enabled filter returns exactly 8 rows |
| No traditional association pollution | P0 | ✅ | 00401 is virtual; existing legacy=855 and competency=384 physical source counts unchanged |
| Cleanup/health | P0 | ✅ | Temporary rows=0, service/nginx active, health OK, critical logs=0 |

### AD. Competency Question Edit and Status

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Question update | ID is empty | P0 | ✅ | Rejected before database access |
| Question update | source question does not exist or is legacy | P0 | ✅ | `id + dimension_id IS NOT NULL` RowsAffected guard rejects without fallback |
| Question update | content is blank | P0 | ✅ | Rejected with specific required-field message |
| Question update | observation point is blank | P0 | ✅ | Rejected with specific required-field message |
| Question update | direction is not `forward/reverse` | P0 | ✅ | Rejected before write |
| Question update | status is not 0/1 or frontend sends empty string | P0 | ✅ | Flexible JSON input rejects empty/nil/fractional/out-of-range values |
| Question update | valid editable fields | P0 | ✅ | Updates only content/observation/direction/status/remark/update_time |
| Question update | identity/history fields supplied by client | P0 | ✅ | Independent request omits identity/history fields; published snapshots untouched |
| Question update | database update fails | P0 | ✅ | Returns controlled error; frontend catch keeps dialog open |
| Question UI | edit opens from row | P0 | ✅ | Pre-fills editable copy and shows immutable identity context |
| Question UI | save succeeds | P0 | ✅ | Closes dialog, notifies, and refreshes current page |
| Question UI | save in progress or fails | P1 | ✅ | Loading disables duplicate submit; failure keeps dialog open |

### AE. Competency Question Edit Staging Acceptance

| Gate | Priority | Coverage | Evidence |
|------|----------|----------|----------|
| Editable field persistence | P0 | ✅ | Temporary source question persisted content/observation/direction/status/remark |
| Identity field protection | P0 | ✅ | Forged code/dimension/item fields ignored; create_time unchanged |
| Invalid and cross-flow requests | P0 | ✅ | Empty status and legacy question ID rejected without mutation |
| Browser edit flow | P0 | ✅ | Dialog opens with immutable identity and all editable fields pre-filled |
| Cleanup/health | P0 | ✅ | Temporary question=0, service/nginx active, health OK, critical logs=0 |

### AF. Competency Question Import UI

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Import template | administrator clicks download | P0 | ✅ | Downloads dedicated nine-column xlsx with fixed filename |
| File selection | no file selected | P0 | ✅ | Preview/formal import disabled with explicit guidance |
| File selection | extension is not xlsx or file exceeds 10 MiB | P0 | ✅ | Rejected client-side; backend guard remains authoritative |
| Import preview | request succeeds with valid rows | P0 | ✅ | Shows counts/normalized rows and retains SHA-256 |
| Import preview | validation has any errors | P0 | ✅ | Shows every row message and disables formal import |
| Import preview | file changes after preview | P0 | ✅ | Clears preview/hash and requires a new preview |
| Formal import | preview is valid and administrator confirms | P0 | ✅ | Re-uploads same file with expectedHash, refreshes list/dimensions |
| Formal import | request fails | P0 | ✅ | Keeps dialog, file and preview state for retry |
| Duplicate submission | preview/import request is running | P1 | ✅ | Loading state disables repeated action |

### AG. Competency Question Import UI Staging Acceptance

| Scenario | Priority | Coverage | Verified result |
|----------|----------|----------|-----------------|
| Dedicated template download | P0 | ✅ | HTTP 200, xlsx MIME, 6512 bytes |
| Template instruction row | P0 | ✅ | Initial real preview reproduced FB-064; after deployment preview skips exact instruction row |
| Valid preview | P0 | ✅ | Temporary unique row: success=1, errors=0, digest retained and confirm enabled |
| Formal import | P0 | ✅ | Browser confirmation imported one row; list refreshed from 384 to 385 |
| Cleanup | P0 | ✅ | Deleted by queried primary key; DB/list returned to 384, temporary question/relation/session/files=0 |
| Service health | P0 | ✅ | Backend/nginx active, health OK, recent panic/fatal/segmentation/import-failed counts all 0 |

### AG2. Competency Question Import/Export Round-Trip Acceptance

| Scenario | Priority | Coverage | Verified result |
|----------|----------|----------|-----------------|
| Template examples | P1 | ✅ | Four-row template contains one valid forward and one valid reverse D01 example |
| Template preview | P0 | ✅ | Real staging preview returned success=2, errors=0 and a 64-character SHA-256 |
| Template import | P0 | ✅ | Same file/hash imported two rows atomically; source question count changed 384→386 |
| Question export | P0 | ✅ | Nine import-compatible columns exported all 386 rows and both temporary examples with correct direction/status |
| Result exports | P0 | ✅ | Existing five-dimension 40/40 result exported 1 summary, 40 details and 40 dictionary rows from both endpoints; normalized content matched |
| Cleanup and health | P0 | ✅ | Temporary examples/session/files=0, source questions restored to 384, health OK, recent critical errors=0 |

### AH. Competency Dimension Master Maintenance

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Dimension update | ID is empty or dimension does not exist | P0 | ✅ | Rejects before/without mutation; no create fallback |
| Identity boundary | client supplies code/create_time | P0 | ✅ | Stable dimension code, ID and create time remain unchanged |
| Required fields | name/VIRD/category/core meaning is blank | P0 | ✅ | Rejects with field-specific message before database write |
| Fixed classifications | VIRD/category is outside supported master values | P0 | ✅ | Rejects arbitrary classification values at HTTP boundary |
| Display order | value is empty, fractional, below 1 or above 48 | P0 | ✅ | Flexible JSON input rejects invalid order |
| Display order | requested 1-48 position is occupied | P0 | ✅ | Atomically swaps the two display orders without violating unique index |
| Status | value is empty, fractional or outside 0/1 | P0 | ✅ | Flexible JSON input rejects invalid state |
| Uniqueness | name conflicts with another dimension | P0 | ✅ | Rejects with controlled conflict message; original row unchanged |
| Valid update | descriptive fields/order/status are valid | P0 | ✅ | Field whitelist update, re-read and return persisted row |
| Historical behavior | source dimension changes after an exam was published | P0 | ✅ | Existing exam snapshots/results unchanged; only future save/publish uses master data |
| Migration rerun | administrator-maintained row already exists | P0 | ✅ | Seed rerun does not overwrite maintained fields or status |
| Maintenance UI | page loads all 48 dimensions | P0 | ✅ | Stable display order, enabled question count, local filters and 20-row paging visible |
| Maintenance UI | edit succeeds | P0 | ✅ | Dialog closes, success feedback and list refresh |
| Maintenance UI | status changes | P0 | ✅ | Explicit confirmation explains impact on new exams; published exams unaffected |
| Maintenance UI | request fails | P1 | ✅ | Keeps dialog values and allows retry; loading prevents duplicate submit |

### AI. Competency Dimension Maintenance Staging Acceptance

| Gate | Priority | Coverage | Evidence |
|------|----------|----------|----------|
| Field update and identity protection | P0 | ✅ | D01 descriptive/status update persisted; forged code/create_time ignored |
| Occupied order swap | P0 | ✅ | D01 1→2 and D02 2→1 completed atomically, then restored 1/2 |
| Invalid requests | P0 | ✅ | Duplicate name and empty status rejected without partial mutation |
| Migration rerun protection | P0 | ✅ | Rerunning 002 retained temporary maintained name/order/status; restore followed |
| Browser navigation and list | P0 | ✅ | 00401 entry opens maintenance page; total 48, default page 20, D01/D02 metadata correct |
| Browser filtering and edit | P0 | ✅ | D42 filter returns one correct row; edit dialog pre-fills all fields and status-change confirmation is explicit |
| Cleanup and integrity | P0 | ✅ | 48 unique codes/names/orders, order 1-48, all enabled, temporary values/session/files=0 |
| Service health | P0 | ✅ | Backend/nginx active, health OK, recent panic/fatal/duplicate/unknown-column counts all 0 |

### AJ. Competency Dynamic Result Export

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Dispatch | assessment is legacy/001/002/003 | P0 | ✅ | Existing template and wide-export behavior remains unchanged |
| Dispatch | assessment is competency | P0 | ✅ | Both existing export endpoints call the same competency workbook builder |
| Authorization | login missing or user lacks export permission | P0 | ✅ | Rejects before any result/snapshot query |
| Exam validation | examId empty/not found or invalid competency mode | P0 | ✅ | Returns controlled error, never falls back to legacy export |
| Data source | submitted complete or timeout-incomplete result exists | P0 | ✅ | Reads persisted result, dimension result, paper answer and publish snapshot values; no score recalculation |
| Result summary | selected dimensions vary by exam | P0 | ✅ | Dynamic columns follow frozen display order and include persisted dimension/overall scores |
| Result summary | incomplete timeout result | P0 | ✅ | Completion rate/integrity retained; missing dimension score exported as blank |
| Question detail | answered/unanswered item | P0 | ✅ | Includes personal sort, snapshot code/content/dimension/observation/direction, raw value/text, final score and answered flag |
| Question dictionary | personal random orders differ | P0 | ✅ | Exports one stable publish snapshot dictionary ordered by snapshot order |
| Empty result | published exam has no saved result | P1 | ✅ | Returns workbook with three sheets and headers, not a false-success malformed file |
| Query failure | any summary/dimension/detail/dictionary query fails | P0 | ✅ | Aborts before response body and returns controlled error |
| Privacy | non-super-admin exports telephone | P0 | ✅ | Masks telephone consistently with existing export policy |
| Frontend entry | competency exam dropdown | P0 | ✅ | Shows result page plus summary/raw export actions; confirmation and download filename are explicit |

### AK. Competency Dynamic Export Staging Acceptance

| Gate | Priority | Coverage | Evidence |
|------|----------|----------|----------|
| Deployment backup and artifact identity | P0 | ✅ | Database backup created; deployed backend/frontend hashes equal local artifacts |
| Real persisted result set | P0 | ✅ | Three candidates, 16 questions each, overall 2/6/8 and D01 1/4/5 |
| Summary endpoint workbook | P0 | ✅ | HTTP xlsx, three sheets, 3 summary rows, D01/D02 dynamic columns and persisted scores verified |
| Raw-answer endpoint workbook | P0 | ✅ | Same normalized three-sheet content as summary endpoint; 48 detail rows and 16 dictionary rows |
| Response contract | P0 | ✅ | RFC 5987 competency filename and xlsx MIME/size verified for both endpoints |
| Browser entry | P0 | ✅ | Competency dropdown shows result + both export actions; confirmation describes all three sheets |
| Cleanup and integrity | P0 | ✅ | Full-chain cleanup=0, temporary exam/results/session/files=0, result orphans=0, competency source questions=384 |
| Service health | P0 | ✅ | Backend/nginx active, health OK, recent panic/fatal/export-failed/unknown-column counts all 0 |

### AL. Competency Expiry Worker Batch and Concurrency

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Startup | overdue papers exist when service starts | P0 | ✅ | Runs one scan immediately instead of waiting the first interval |
| Scheduled scan | no overdue papers | P1 | ✅ | One bounded query, no writes, no error |
| Owner resolution | candidate/tester paper | P0 | ✅ | Resolves participant type in the batch scan query, no per-paper owner query |
| Owner resolution | owner missing | P0 | ✅ | Logs paper ID, leaves state unchanged for repair/retry, continues batch |
| Batch bound | overdue count exceeds configured size | P0 | ✅ | Stable `limit_time,id` order and configured LIMIT |
| Timeout submit | partially answered paper | P0 | ✅ | Same submit service persists partial averages and marks incomplete/timeout |
| Timeout submit | zero-answer dimension/paper | P0 | ✅ | No fake minimum scores; nil dimension score and zero effective dimensions retained |
| Submit failure | one paper fails | P0 | ✅ | Logs without exposing token/data, continues other papers, failed paper retries next scan |
| Concurrent submit | frontend and Worker submit same expired paper | P0 | ✅ | Row lock/idempotency yields exactly one overall result and one result set |
| Capacity | 100 independent shuffles of 384 questions | P1 | ✅ | Every in-memory order is complete/unique and sample contains multiple distinct permutations |
| Shutdown | context cancelled | P0 | ✅ | Immediate scan/ticker goroutine exits cleanly |

### AM. Competency Expiry Worker Staging Acceptance

| Gate | Priority | Coverage | Evidence |
|------|----------|----------|----------|
| Deployment backup and identity | P0 | ✅ | Database backup created; deployed backend SHA-256 equals local Linux artifact |
| Concurrent expired submit | P0 | ✅ | Two simultaneous manual/timeout-compatible requests produced 1 overall + 2 dimension results |
| Candidate partial timeout | P0 | ✅ | 1/16 answered, effective dimensions=1, overall=3, one nil dimension, incomplete timeout |
| Tester zero-answer timeout | P0 | ✅ | 0/16 answered, effective dimensions=0, overall=0, two nil dimensions, incomplete timeout |
| Startup immediate scan | P0 | ✅ | Two newly expired papers submitted within 15 seconds after service restart, below 30-second interval |
| Cleanup and integrity | P0 | ✅ | Full-chain cleanup=0, temporary exam/results=0, result orphans=0, running expired competency papers=0, source questions=384 |
| Service health | P0 | ✅ | Backend/nginx active, health OK, recent panic/fatal/worker-failed/owner-missing counts all 0 |

### AN. Full 48-Dimension Capacity Chain

| Gate | Priority | Coverage | Planned assertion |
|------|----------|----------|-------------------|
| Publish | all 48 enabled dimensions selected | P0 | ✅ | Staging froze exactly 48 dimensions and 384 unique source questions in 0.098s |
| Paper creation | 100 participants start independently | P0 | ✅ | Created exactly 100 papers and 38,400 paper-question rows |
| Set integrity | each 384-question paper | P0 | ✅ | Every paper contained 384 unique question IDs/codes with no missing or duplicate rows |
| Random independence | compare 100 persisted orders | P0 | ✅ | Persisted orders produced 100/100 distinct SHA-256 hashes |
| Refresh stability | reload each paper detail | P0 | ✅ | All 100 orders remained byte-for-byte stable after second read |
| Runtime response | 384-question detail | P1 | ✅ | Create chain p50/p95/max=0.479/0.758/0.837s; refresh=0.063/0.105/0.125s on staging |
| Source isolation | capacity papers use published snapshot | P0 | ✅ | All 38,400 rows bound exam_question_id; snapshot count/unique source/code=384 |
| Cleanup | delete capacity exam | P0 | ✅ | Full-chain transaction removed all capacity data; cleanup_remaining=0 |
| Service health | after capacity cleanup | P0 | ✅ | Backend remained healthy; final service/log check follows stage-6 closeout |