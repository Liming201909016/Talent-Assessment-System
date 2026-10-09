# Production Release Candidate — 2026-10-09

## Current verdict

**LOCAL GREEN / PRODUCTION NOT GO.** The current candidate is `release-candidate-20261008T210616Z`. It supersedes `release-candidate-20261008T203459Z`; the older directory, ZIP, manifest, and receipts are preserved and were not relabeled or modified. No remote access, deployment, database write, migration execution, or production approval occurred in this evidence-only addendum.

The hash-bound, sanitized reviewer evidence is [evidence-addendum.json](../scripts/test/results/release-candidate-20261008T210616Z/evidence-addendum.json), file SHA-256 `a8c19cbf711d7cdba83dde27cca39410227d4d28f57102758811660bb9fb7a6e`. Its verdict binding root is `c0b5d8d30c0561929854b4077aae6e39ca51504eb9c5174276cf297061139190`.

## Current candidate

- Directory: Go-based Refactored System/bin/release-candidate-20261008T210616Z/
- Archive: release-candidate-20261008T210616Z.zip
- Archive SHA-256: `ed0ed462db621fc3ce4fcddbfb7fb276fe82e4ff781be15449f385abf5a516af`
- Archive bytes: 33,909,116
- Linux server SHA-256: `331862e226d213caac6bc9b8d51385e2d7a8aa2839ed8d2aaef0d1209aa93065`
- Frontend: 393 files; aggregate SHA-256 `9df90ced208abb0aeef5ed99b9538f9b9fcfae6f2046033cdc32d423a997ec9b`; source maps 0.
- Git HEAD recorded by the immutable candidate: `5217ce6558eb7876d9e4cc37f1a617296deb2590`; the candidate remains bound to byte-level inputs rather than Git HEAD alone.

## Immutable build inputs and receipts

The original candidate recipe was recovered and replayed without writing source files. It selects every Go file under cmd/, internal/, and pkg/; go.mod/go.sum; all files under ruoyi-ui/src and ruoyi-ui/public; and four frontend build files. The result is exactly **654 files**.

- Candidate source aggregate: `0e1e9d1bbf0401527d8c2fc21f6c96903f9a8c3ea1a532821216055df3263d97`.
- Current replay: 654 files and the same aggregate. No current worktree state was substituted for an unmatched candidate snapshot.
- Compact path-set SHA-256: `06f31f26c4b2c0a6b8c2d3e100f78bf1c9a01d46a66976ae961273ff8a2530e9`.
- Compact path/SHA aggregate: `334f9a7c167da1ec452dd228e3f7e081921a8b6bc687ce92e544c85e3679128c`.
- Linux build receipt binds server SHA `331862…93065` to that source aggregate.
- Windows same-source runtime receipt binds SHA `0d8589d18183ba545929570d3fbed1df61d87ee09589af360874e02b2962c47c` to the same aggregate. The Windows artifact is not retained and is not the packaged Linux artifact.

## Runtime gate and failed first attempt

The first local gate ended with `HEALTH_TIMEOUT_120S`; its immutable receipt is [runtime-http-gate-failed.json](../scripts/test/results/release-candidate-20261008T210616Z/runtime-http-gate-failed.json), SHA-256 `169408d7d04bf787180e6f78f1610e7e31043770eb57f44c4560bcf1e7c700d2`. The retained receipt shows a listening marker but does not prove the server was not ready before the probe. Classification is therefore **unattributed startup/probe timing, later success**.

The later successful receipt is [runtime-http-gate.json](../scripts/test/results/release-candidate-20261008T210616Z/runtime-http-gate.json), SHA-256 `1a57b97fd230da444920e27886adbd4e200a11216e8c9e138af5f2d037f0cde2`: health 200, legacy 200, and participant/admin/reissue/formal routes 404 with workers disabled. This was a same-source Windows runtime gate, not execution of the Linux ELF.

## Sanitized database evidence

The pre/post gate receipt SHA is identical: `913e3c87fd3648c65e0e9b6a1de398ee06654a224bc329b716b862bb62ad81bf`. The addendum contains only counts, invariant hashes, and SHA-256 identifiers—no raw IDs, PII, identity-field hashes, credentials, or connection material.

- All 20 obsolete-table counts are 0.
- All 20 named global orphan checks are 0.
- Retained 00501 and 00502 chains each remain `1 exam / 1 profile / 1 draft / 1 candidate / 1 paper / 1 snapshot / 140 snapshot questions / 140 legacy questions / 700 answer buckets / 1 run / 13 dimensions / 4 modules / 1 receipt`.
- Each retained report chain has 0 revisions, 0 current pointers, 0 audits, 0 reissues, and 0 reissue audits at the candidate gate boundary.
- Baseline generation is `1791483691965`; only exam/run/title hashes are copied into the addendum.
- Two prior report-only receipts are bound by SHA and record zero answer-click/save/start/submit requests; their run/report identifiers remain hash-only.
- Legacy copy invariant SHA is `c27fdfb8e235ed36923e68297419a3f4f6731ec49564620f438bb57e11140b5a`; the frozen manifest/input/mapping/field-contract hashes are preserved in the addendum.

Historical limitation remains explicit: global orphan count 0 closes parent-absent residual risk, but cannot enumerate deleted obsolete child IDs or prove the absence of a separate complete non-orphan chain outside the ownership manifest.

## Package scope and integrity

Included only the Linux server, production frontend, five traditional export templates, and three previously approved competency report templates. Management-traits TEST XLSX/DOCX assets, SQL migrations, local fixtures, DPAPI material, debug binaries, scripts, results, environment files, and source maps remain excluded.

- ZIP extraction previously verified all 404 SHA256SUMS entries with 0 errors.
- Linux ELF metadata reports Go 1.26.2, CGO_ENABLED=0, GOOS=linux, GOARCH=amd64, and trimpath.
- Secret scan: PASS; 0 high-confidence findings and 0 forbidden package paths.
- Candidate ZIP, candidate manifest, original receipts, and the superseded candidate were not changed by this addendum.

## Remaining production blockers

1. Management-traits TEST assets and formal capabilities remain excluded, disabled, or not approved.
2. All migrations remain excluded and were not executed.
3. The Linux binary was not executed on this Windows host; WSL has zero registered distributions.
4. No staging or production validation was performed for this addendum.
5. A separate explicit production decision is still required before deployment.
