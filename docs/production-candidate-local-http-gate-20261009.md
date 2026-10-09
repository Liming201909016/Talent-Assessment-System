# Production-like candidate HTTP gate — local evidence (2026-10-09)

## Current verdict

**PASS FOR LOCAL SAME-SOURCE WINDOWS RUNTIME GATE; LINUX RUNTIME UNAVAILABLE.** The current candidate is `release-candidate-20261008T210616Z`, which supersedes the preserved `release-candidate-20261008T203459Z`. This was not a production-config startup, deployment, staging test, or production test.

The successful receipt is [runtime-http-gate.json](../scripts/test/results/release-candidate-20261008T210616Z/runtime-http-gate.json). The initial failed attempt is preserved at [runtime-http-gate-failed.json](../scripts/test/results/release-candidate-20261008T210616Z/runtime-http-gate-failed.json). The complete sanitized binding is [evidence-addendum.json](../scripts/test/results/release-candidate-20261008T210616Z/evidence-addendum.json).

## Intended environment boundary

The safe intended combination was:

- `APP_ENV=local` so the production configuration overlay and remote production resources would not be loaded.
- `REPORT_EFFECTIVE_ENV=production` and `MNG_TEST_REPORT_ENV=production` so the independent management-traits TEST runtime gate would be disabled.
- Existing isolated database copy tunnel on `127.0.0.1:23316`.
- A new owned Redis instance on `127.0.0.1:23318` with a separate private config and random password.
- Candidate HTTP on `127.0.0.1:18093`.

This would have tested only production-like feature-gate assembly while retaining local configuration. It would not have represented a full production configuration startup.

## Safety controls used

The superseding candidate added a validated loopback host setting and a local-only background-worker disable control. The successful gate used `SERVER_HOST=127.0.0.1`, `SERVER_PORT=18093`, and `LOCAL_DISABLE_BACKGROUND_WORKERS=true` with the production-valued management-traits feature gates. Redis remained isolated on `127.0.0.1:23318`.

The prior all-interface-listener and unavoidable-worker findings apply only to the preserved 203459 candidate and are not rewritten as having passed.

## Verified local evidence

- Packaged Linux server SHA-256: `331862e226d213caac6bc9b8d51385e2d7a8aa2839ed8d2aaef0d1209aa93065`.
- Same-source Windows runtime server SHA-256: `0d8589d18183ba545929570d3fbed1df61d87ee09589af360874e02b2962c47c`.
- Both receipts are bound to the same 654-file source aggregate `0e1e9d1bbf0401527d8c2fc21f6c96903f9a8c3ea1a532821216055df3263d97`.
- Successful route results: health 200, legacy 200, participant/admin/reissue/formal 404.
- Workers were disabled, schema precheck was not executed, and the DB receipt remained byte-identical before/after.
- Candidate listener residual after cleanup: 0; private runtime removed; existing services unchanged.
- All 404 archive entries were previously verified with 0 integrity errors.
- Packaged frontend: 393 files, aggregate SHA-256 `9df90ced208abb0aeef5ed99b9538f9b9fcfae6f2046033cdc32d423a997ec9b`, source maps 0.

The first attempt recorded `HEALTH_TIMEOUT_120S` even though its retained log summary contains a server listening marker. Because the logs do not prove startup was not ready before the probe, the root classification is **unattributed startup/probe timing, later success**. The later receipt is authoritative for the successful gate; the failed receipt remains immutable evidence.

## Linux runtime availability and static metadata

- The Windows host has `DISTRO_COUNT=0` under the WSL registry; no WSL distribution was available.
- The Linux/amd64 ELF was therefore not executed locally.
- Static `go version -m` evidence reports Go 1.26.2, `CGO_ENABLED=0`, `GOOS=linux`, `GOARCH=amd64`, `GOAMD64=v1`, and `-trimpath=true`.
- Static ELF/build metadata and the same-source Windows gate are separate facts; neither is presented as an actual Linux HTTP run.

## Not executed in this addendum

- No new candidate process, Redis process, HTTP call, database operation, test, or source build.
- No remote access, staging action, production action, migration, or deployment.
- No existing service was stopped, restarted, or reconfigured.
- The candidate ZIP/manifest and all old receipts remained unchanged.

## Readiness conclusion

The local disabled-assembly HTTP gate is complete for the same-source Windows runtime artifact and is hash-bound to candidate 210616. The packaged Linux ELF remains statically verified but not runtime-executed on this host. This evidence is **not production deployment approval**; candidate 210616 remains `LOCAL_ONLY_NOT_GO` pending a separate production decision.
