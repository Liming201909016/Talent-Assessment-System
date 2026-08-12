const assert = require('assert');
const fs = require('fs');
const path = require('path');
const XLSX = require('./node_modules/xlsx');

const root = path.resolve(__dirname, '..', '..');
const candidatePath = path.join(root, 'scripts', 'data', 'competency-phase1-candidate-20260807.json');
const importPath = path.join(root, 'scripts', 'data', 'competency-phase1-import-20260810.xlsx');
const candidate = JSON.parse(fs.readFileSync(candidatePath, 'utf8'));

assert.strictEqual(candidate.schemaVersion, 'competency-phase1-import-v3', 'resolved phase-1 schema version');
assert.strictEqual(candidate.readiness.importReady, true, 'confirmed P0/P1 decisions must make the candidate import-ready');
assert.deepStrictEqual(candidate.readiness.blockingIssueIds, [], 'resolved candidate must have no blockers');
assert.deepStrictEqual(candidate.readiness.warningIssueIds, [], 'resolved candidate must have no warnings');

const expectedDimensions = [
  ['competency-a1-01', 'A1-01', '逻辑思维', 'general_ability', 1],
  ['competency-a1-02', 'A1-02', '数字应用', 'general_ability', 2],
  ['competency-a1-03', 'A1-03', '计划执行', 'general_ability', 3],
  ['competency-a1-04', 'A1-04', '持续学习', 'general_ability', 4],
  ['competency-a1-05', 'A1-05', '沟通表达', 'general_ability', 5],
  ['competency-b1-01', 'B1-01', '敬业奉献', 'psychological_quality', 6],
  ['competency-b1-02', 'B1-02', '求真务实', 'psychological_quality', 7],
  ['competency-b1-03', 'B1-03', '自律性', 'psychological_quality', 8],
  ['competency-b1-04', 'B1-04', '成就导向', 'psychological_quality', 9],
  ['competency-b1-05', 'B1-05', '合作意识', 'psychological_quality', 10]
];

assert.strictEqual(candidate.dimensions.length, expectedDimensions.length, 'phase-1 must contain exactly ten confirmed dimensions');
for (let index = 0; index < expectedDimensions.length; index++) {
  const [id, code, name, firstLevelCode, phaseOrder] = expectedDimensions[index];
  const dimension = candidate.dimensions[index];
  assert.deepStrictEqual(
    [dimension.dimensionId, dimension.dimensionCode, dimension.sourceName, dimension.firstLevelCode, dimension.phaseOrder],
    [id, code, name, firstLevelCode, phaseOrder],
    `dimension ${phaseOrder} identity mismatch`
  );
  assert.ok(!Object.prototype.hasOwnProperty.call(dimension, 'proposedExistingDimensionId'), `dimension ${code} retains retired D mapping`);
  assert.ok(!Object.prototype.hasOwnProperty.call(dimension, 'proposedExistingDimensionCode'), `dimension ${code} retains retired D code`);
}

const byOrder = new Map(candidate.dimensions.map(item => [item.phaseOrder, item]));
for (const question of candidate.questions) {
  const order = question.questionType === 'dimension'
    ? question.phaseDimensionOrder
    : question.associatedPhaseDimensionOrder;
  const dimension = byOrder.get(order);
  assert.ok(dimension, `question ${question.candidateQuestionCode} has unknown phase dimension order ${order}`);
  assert.strictEqual(question.dimensionId, dimension.dimensionId, `question ${question.candidateQuestionCode} dimension ID mismatch`);
  assert.strictEqual(question.dimensionCode, dimension.dimensionCode, `question ${question.candidateQuestionCode} dimension code mismatch`);
  assert.ok(!Object.prototype.hasOwnProperty.call(question, 'proposedExistingDimensionId'), `question ${question.candidateQuestionCode} retains retired D mapping`);
  assert.ok(!Object.prototype.hasOwnProperty.call(question, 'proposedExistingDimensionCode'), `question ${question.candidateQuestionCode} retains retired D code`);
  if (question.questionType === 'dimension') {
    assert.strictEqual(
      question.candidateQuestionCode,
      `${dimension.dimensionCode}-Q${String(question.itemNo).padStart(2, '0')}`,
      `dimension question ${question.candidateQuestionCode} must use the stable A/B prefix`
    );
  } else {
    assert.strictEqual(question.questionType, 'validity', `unexpected question type ${question.questionType}`);
    assert.strictEqual(question.direction, 'forward', `validity question ${question.candidateQuestionCode} must be forward`);
  }
  assert.strictEqual(question.proposedStatus, 'enabled', `question ${question.candidateQuestionCode} must be enabled`);
  assert.strictEqual(question.importStatus, 'ready_for_staging_candidate_import', `question ${question.candidateQuestionCode} import status`);
}

for (const content of candidate.dimensionContent) {
  const dimension = byOrder.get(content.phaseOrder);
  assert.strictEqual(content.dimensionId, dimension.dimensionId, `content ${content.sourceDimensionName} dimension ID mismatch`);
  assert.strictEqual(content.dimensionCode, dimension.dimensionCode, `content ${content.sourceDimensionName} dimension code mismatch`);
  assert.ok(!Object.prototype.hasOwnProperty.call(content, 'proposedExistingDimensionId'), `content ${content.sourceDimensionName} retains retired D mapping`);
}

assert.ok(!candidate.readiness.blockingIssueIds.includes('MAP-001'), 'resolved A/B identity must remove MAP-001 blocker');
assert.ok(!candidate.issues.some(issue => issue.id === 'MAP-001'), 'resolved A/B identity must remove MAP-001 issue');
assert.ok(!JSON.stringify(candidate).includes('competency-d'), 'candidate JSON contains retired D identity');
assert.ok(!JSON.stringify(candidate).match(/"D\d{2}"/), 'candidate JSON contains retired D code');
assert.strictEqual(candidate.readiness.databaseWritten, false, 'identity conversion must not claim database writes');
assert.deepStrictEqual(candidate.versions, {
  product: 'competency-frontline-phase1-v1',
  scoring: 'competency-phase1-scoring-v1',
  content: 'competency-phase1-content-v1',
  reportTemplate: 'competency-phase1-report-v1'
}, 'confirmed phase-1 versions');
assert.strictEqual(candidate.questionSummary.unresolvedValidityDirections, 0, 'validity directions must be resolved');
assert.strictEqual(candidate.questionSummary.validityForward, 10, 'all ten validity questions must be forward');

const importWorkbook = XLSX.read(fs.readFileSync(importPath), { type: 'buffer', raw: false });
assert.deepStrictEqual(importWorkbook.SheetNames, ['胜任力题目'], 'import workbook sheet names');
const importRows = XLSX.utils.sheet_to_json(importWorkbook.Sheets['胜任力题目'], { header: 1, defval: '', raw: false });
const importHeaders = ['维度序号', '维度名称', '题目类型', '题目编号', '维度内题号', '题目内容', '考察点', '计分方向', '启用状态', '备注'];
assert.deepStrictEqual(importRows[0], importHeaders, 'phase-1 import headers');
assert.strictEqual(importRows.length, 91, 'header plus 90 import rows');
const dataRows = importRows.slice(1);
assert.strictEqual(dataRows.filter(row => row[2] === '维度题').length, 80, 'dimension import rows');
assert.strictEqual(dataRows.filter(row => row[2] === '效度题').length, 10, 'validity import rows');
assert.strictEqual(dataRows.filter(row => row[2] === '维度题' && row[7] === '正向').length, 62, 'forward dimension import rows');
assert.strictEqual(dataRows.filter(row => row[2] === '维度题' && row[7] === '反向').length, 18, 'reverse dimension import rows');
assert.strictEqual(dataRows.filter(row => row[2] === '效度题' && row[7] === '正向').length, 10, 'forward validity import rows');
assert.strictEqual(dataRows.filter(row => row[8] === '启用').length, 90, 'enabled import rows');
assert.strictEqual(new Set(dataRows.map(row => row[3])).size, 90, 'unique import question codes');
for (const question of candidate.questions) {
  const row = dataRows.find(item => item[3] === question.candidateQuestionCode);
  assert.ok(row, `import row missing: ${question.candidateQuestionCode}`);
  assert.strictEqual(row[2], question.questionType === 'validity' ? '效度题' : '维度题', `import type mismatch: ${question.candidateQuestionCode}`);
  assert.strictEqual(row[7], question.direction === 'reverse' ? '反向' : '正向', `import direction mismatch: ${question.candidateQuestionCode}`);
}

console.log('COMPETENCY_PHASE1_IDENTITY_TEST_PASS');
