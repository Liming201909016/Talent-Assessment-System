-- DISABLED 2026-10-10: DO NOT USE FOR DELETION.
--
-- Production read-only closure proved that the 409 papers whose parent exam is
-- absent are approved historical retention exceptions. They include 210
-- completed papers, existing sys_user ownership and answered history. The
-- earliest available full backup does not contain the 14 missing parent exams.
--
-- The former script was unsafe for this data set because it:
--   1. deleted retained business history without an ownership manifest;
--   2. used NOT IN, which has unsafe NULL semantics;
--   3. did not close el_exam_repo, el_user_exam or el_user_book references;
--   4. deleted parents before validating every child/reference path;
--   5. had no backup SHA, restore rehearsal, exact-ID allowlist or rollback gate.
--
-- Authoritative evidence and retention decision:
--   docs/production-paper-orphan-governance-20261010.md
--
-- This file is intentionally read-only. Any future deletion requires a new,
-- exact primary-key allowlist script under a separately approved change.

SET SESSION TRANSACTION READ ONLY;

SELECT 'RETAINED_ORPHAN_PAPER_SUMMARY' AS step;
SELECT p.state, COUNT(*) AS paper_count
FROM el_paper p
WHERE NOT EXISTS (SELECT 1 FROM el_exam e WHERE e.id = p.exam_id)
GROUP BY p.state
ORDER BY p.state;

SELECT 'RETAINED_ORPHAN_PAPER_REFERENCES' AS step;
SELECT
	(SELECT COUNT(*)
	 FROM el_paper p
	 WHERE NOT EXISTS (SELECT 1 FROM el_exam e WHERE e.id = p.exam_id)) AS papers,
	(SELECT COUNT(*)
	 FROM el_paper_qu pq
	 JOIN el_paper p ON p.id = pq.paper_id
	 WHERE NOT EXISTS (SELECT 1 FROM el_exam e WHERE e.id = p.exam_id)) AS paper_questions,
	(SELECT COUNT(*)
	 FROM el_paper_qu_answer pqa
	 JOIN el_paper p ON p.id = pqa.paper_id
	 WHERE NOT EXISTS (SELECT 1 FROM el_exam e WHERE e.id = p.exam_id)) AS paper_answers;
