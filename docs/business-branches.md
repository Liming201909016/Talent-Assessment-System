# Business Branches

| Area | Branch | Status | Notes |
|------|--------|--------|-------|
| MBTI full report generation | document.xml contains static body runs with w14:textFill / w14:props3d | ✅ | Triggered by production tofu-box issue; now covered by FB-042 fallback |
| MBTI full report generation | document.xml contains risky static body font families such as HYYakuHei / 汉仪雅酷黑 | ✅ | Covered by FB-043 font-family normalization fallback |
| MBTI full report generation | document.xml contains East Asian static body runs with only w:hint and no explicit font family | ✅ | Triggered by production ESTP "功利型/凭借" tofu-box issue; covered by FB-044 |
| MBTI full report generation | styles.xml / fontTable.xml declare unstable CJK fonts used by body fallback | ✅ | Covered by FB-044 style/font-table normalization |

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
| Report audience | result score, dimension order, and charts across versions | P0 | ✅ | Shared result tables and one report component; audience does not enter scoring |
| Report audience | historical report regeneration | P0 | ✅ | Report data reads `el_competency_result.report_audience` snapshot |

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
| Import template | administrator downloads template | P0 | ✅ | HTTP test parses xlsx and verifies nine headers, three rows, and zero merged cells |
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
| Result UI | legacy exam row | P0 | ✅ | Existing test-record/export/statistics entries remain in the non-competency branch |
| Result UI | competency exam row | P0 | ✅ | Dedicated `CompetencyResults` route and list command preserve `examId` context |
| Result UI | complete competency result opens test report | P0 | ✅ | Named route receives `params.paperId`; FB-067 RED→GREEN verifies the report page can read `$route.params.paperId` |

### M2. Temporary Competency Formal Report and PDF

| Area | Branch | Priority | Coverage | Planned assertion |
|------|--------|----------|----------|-------------------|
| Report text | active temporary content version exists for audience, overall level and every measured dimension level | P0 | ✅ | 392 temp-v1 rows cover 2 audiences × overall/dimensions × 4 levels; exact matches are frozen into instance snapshot |
| Report text | any required text is missing or belongs to another audience/version | P0 | ✅ | Pure service test rejects missing dimension level and cross-audience fallback before rendering |
| Report text | temporary content is rendered | P0 | ✅ | Staging PDF text contains “临时测试报告” and “不可作为人才决策依据” |
| Generate report | result is incomplete | P0 | ✅ | Existing FB-047 formal report guard rejects before instance/PDF writes |
| Generate report | complete result has no prior instance | P0 | ✅ | Staging created instance, rendered Chromium PDF, persisted path/hash/size and completed status |
| Generate report | same version already completed and force=false | P0 | ✅ | Unique paper+version and existing-file guard return the same instance without duplication |
| Regenerate report | force=true for same content version | P0 | ✅ | Same instance is refreshed and regenerate audit action is recorded by the dedicated branch |
| Generate report | render or file write fails | P0 | ✅ | Handler marks instance failed and never sets participant pdf_flag success |
| Download report | completed instance path is inside configured upload root and file exists | P0 | ✅ | Staging downloaded application/pdf, SHA-256 matched instance, and audit count was 1 |
| Download report | missing instance/file or path escapes upload root | P0 | ✅ | `filepath.Rel` allow-root guard and completed-instance lookup reject invalid paths |
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