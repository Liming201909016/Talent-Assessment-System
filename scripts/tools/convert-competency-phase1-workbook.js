const crypto = require('crypto');
const fs = require('fs');
const path = require('path');
const XLSX = require('../test/node_modules/xlsx');

const TOOL_VERSION = '1.5.1';
const SCHEMA_VERSION = 'competency-phase1-import-v3';
const CSV_CONTRACT_VERSION = 'competency-phase1-csv-v1';
const MATERIAL_DATE = '2026-08-10';
const DEFAULT_SOURCE_NAME = '260810基层员工胜任力测评题本+等级评价+总体评价V1.xlsx';
const REPORT_TEMPLATE_NAME = '胜任力测评报告样例.docx';
const DEFAULT_OUTPUT_NAME = 'competency-phase1-candidate-20260807.json';
const DEFAULT_IMPORT_OUTPUT_NAME = 'competency-phase1-import-20260810.xlsx';

const root = path.resolve(__dirname, '..', '..');
const defaultSource = path.join(root, 'docs', '260807胜任力开发资料', DEFAULT_SOURCE_NAME);
const reportTemplatePath = path.join(root, 'docs', '260807胜任力开发资料', REPORT_TEMPLATE_NAME);
const defaultOutput = path.join(root, 'scripts', 'data', DEFAULT_OUTPUT_NAME);
const defaultImportOutput = path.join(root, 'scripts', 'data', DEFAULT_IMPORT_OUTPUT_NAME);

const importHeaders = [
  '维度序号', '维度名称', '题目类型', '题目编号', '维度内题号',
  '题目内容', '考察点', '计分方向', '启用状态', '备注'
];

const confirmedVersions = {
  product: 'competency-frontline-phase1-v1',
  scoring: 'competency-phase1-scoring-v1',
  content: 'competency-phase1-content-v1',
  reportTemplate: 'competency-phase1-report-v1'
};

const dimensionMappings = [
  { phaseOrder: 1, firstLevelCode: 'general_ability', firstLevelName: '通用能力', sourceLayer: '通用能力层', sourceName: '逻辑思维', dimensionId: 'competency-a1-01', dimensionCode: 'A1-01', identityStatus: 'confirmed_phase1' },
  { phaseOrder: 2, firstLevelCode: 'general_ability', firstLevelName: '通用能力', sourceLayer: '通用能力层', sourceName: '数字应用', dimensionId: 'competency-a1-02', dimensionCode: 'A1-02', identityStatus: 'confirmed_phase1' },
  { phaseOrder: 3, firstLevelCode: 'general_ability', firstLevelName: '通用能力', sourceLayer: '通用能力层', sourceName: '计划执行', dimensionId: 'competency-a1-03', dimensionCode: 'A1-03', identityStatus: 'confirmed_phase1' },
  { phaseOrder: 4, firstLevelCode: 'general_ability', firstLevelName: '通用能力', sourceLayer: '通用能力层', sourceName: '持续学习', dimensionId: 'competency-a1-04', dimensionCode: 'A1-04', identityStatus: 'confirmed_phase1' },
  { phaseOrder: 5, firstLevelCode: 'general_ability', firstLevelName: '通用能力', sourceLayer: '通用能力层', sourceName: '沟通表达', dimensionId: 'competency-a1-05', dimensionCode: 'A1-05', identityStatus: 'confirmed_phase1' },
  { phaseOrder: 6, firstLevelCode: 'psychological_quality', firstLevelName: '心理素养', sourceLayer: '心理素养', sourceName: '敬业奉献', dimensionId: 'competency-b1-01', dimensionCode: 'B1-01', identityStatus: 'confirmed_phase1' },
  { phaseOrder: 7, firstLevelCode: 'psychological_quality', firstLevelName: '心理素养', sourceLayer: '心理素养', sourceName: '求真务实', dimensionId: 'competency-b1-02', dimensionCode: 'B1-02', identityStatus: 'confirmed_phase1' },
  { phaseOrder: 8, firstLevelCode: 'psychological_quality', firstLevelName: '心理素养', sourceLayer: '心理素养', sourceName: '自律性', dimensionId: 'competency-b1-03', dimensionCode: 'B1-03', identityStatus: 'confirmed_phase1' },
  { phaseOrder: 9, firstLevelCode: 'psychological_quality', firstLevelName: '心理素养', sourceLayer: '心理素养', sourceName: '成就导向', dimensionId: 'competency-b1-04', dimensionCode: 'B1-04', identityStatus: 'confirmed_phase1' },
  { phaseOrder: 10, firstLevelCode: 'psychological_quality', firstLevelName: '心理素养', sourceLayer: '心理素养', sourceName: '合作意识', dimensionId: 'competency-b1-05', dimensionCode: 'B1-05', identityStatus: 'confirmed_phase1' }
];

const expectedOptionLabels = ['完全不符合', '比较不符合', '不确定', '比较符合', '完全符合'];
const levelDefinitions = [
  { code: 'L1', secondaryLabel: '差', firstLevelLabel: '低分', sourceRange: '1.0—1.7分' },
  { code: 'L2', secondaryLabel: '较差', firstLevelLabel: '较低分', sourceRange: '1.7—2.7分' },
  { code: 'L3', secondaryLabel: '合格', firstLevelLabel: '中分', sourceRange: '2.7—3.5分' },
  { code: 'L4', secondaryLabel: '较优秀', firstLevelLabel: '较高分', sourceRange: '3.5—4.3分' },
  { code: 'L5', secondaryLabel: '优秀', firstLevelLabel: '高分', sourceRange: '4.3—5.0分' }
];

const dimensionBounds = [
  { code: 'L1', minScore: 1, minInclusive: 1, maxScore: 1.7, maxInclusive: 1 },
  { code: 'L2', minScore: 1.7, minInclusive: 0, maxScore: 2.7, maxInclusive: 1 },
  { code: 'L3', minScore: 2.7, minInclusive: 0, maxScore: 3.5, maxInclusive: 1 },
  { code: 'L4', minScore: 3.5, minInclusive: 0, maxScore: 4.3, maxInclusive: 1 },
  { code: 'L5', minScore: 4.3, minInclusive: 0, maxScore: 5, maxInclusive: 1 }
];

const overallDefinitions = [
  { code: 'excellent', name: '优秀胜任', minScore: 45, minInclusive: 1, maxScore: '', maxInclusive: '', sourceRange: '45分以上' },
  { code: 'good', name: '良好胜任', minScore: 40, minInclusive: 1, maxScore: 45, maxInclusive: 0, sourceRange: '40-45分' },
  { code: 'qualified', name: '合格胜任', minScore: 32.5, minInclusive: 1, maxScore: 40, maxInclusive: 0, sourceRange: '32.5-40分' },
  { code: 'weak', name: '薄弱胜任', minScore: 25, minInclusive: 1, maxScore: 32.5, maxInclusive: 0, sourceRange: '25-32.5分' },
  { code: 'not_qualified', name: '尚未胜任', minScore: 10, minInclusive: 1, maxScore: 25, maxInclusive: 0, sourceRange: '25分以下' }
];

const customerSpecialNotice = '科学测评的原则是多质多法，即对每个素质的评价采用多种测试手段，本测评作为一种测试手段，虽证明有效，其结果基于受测者自陈反应结果，受其测评状态、自我认知偏差等影响，不能作为精准判断个体岗位胜任力的唯一依据。更精准的测试建议根据实际情况，结合其他测试结果，如面试、绩效表现、情景模拟、360评估等进行综合判断。';

const reportStaticTexts = [
  { contentType: 'template', matchKey: 'cover_title', levelCode: 'fixed', content: '胜任力测评报告', reviewStatus: 'pending_review', reviewNote: '' },
  { contentType: 'template', matchKey: 'cover_english', levelCode: 'fixed', content: 'Competency Assessment Report', reviewStatus: 'pending_review', reviewNote: '' },
  { contentType: 'template', matchKey: 'cover_slogan', levelCode: 'fixed', content: '洞悉深层素质 ● 驱动精准发展', reviewStatus: 'pending_review', reviewNote: '' },
  { contentType: 'template', matchKey: 'reading_background', levelCode: 'fixed', content: '在组织人才竞争日益激烈的当下，科学评估员工的岗位胜任力水平已成为企业人力资源管理的关键环节。无数研究与实践证明，员工绩效差异往往并非仅由知识储备与基础技能的高低所决定，而是更多源于更深层的素质——包括能力倾向、价值取向、心理动力，这些深层特征构成了区分绩优者与一般者的员工胜任力，是本测评关注的焦点。', reviewStatus: 'pending_review', reviewNote: '' },
  { contentType: 'template', matchKey: 'reading_dimensions', levelCode: 'fixed', content: '本测评以麦克利兰（McClelland）的冰山模型与博亚特兹（Boyatzis）的洋葱模型为理论基础，聚焦鉴别性素质的评估，从通用能力、心理素养两个方面共计10个维度对员工进行系统考察，旨在全面评估影响工作表现的关键因素。其中，通用能力包括逻辑思维、数字应用、计划执行、持续学习、沟通表达5个子维度，心理素养包括敬业奉献、求真务实、自律性、成就导向和合作意识5个子维度。', reviewStatus: 'pending_review', reviewNote: '' },
  { contentType: 'template', matchKey: 'reading_usage', levelCode: 'fixed', content: '本测评能够有效鉴别员工的能力短板与发展优势，可以应用于组织的招聘、选拔、在岗评估、培训发展与人才盘点等，为组织客观量化人才评价与个性化人才培养提供数据支撑，辅助实现人岗匹配与人才发展战略的制定与落地。', reviewStatus: 'pending_review', reviewNote: '' },
  { contentType: 'template', matchKey: 'special_notice', levelCode: 'fixed', content: customerSpecialNotice, reviewStatus: 'pending_review', reviewNote: '' },
  { contentType: 'group', matchKey: 'general_ability', levelCode: 'description', content: '“通用能力”由逻辑思维、数字应用、计划执行、持续学习、沟通表达五个子维度构成，其得分是这五个子维度的综合得分，反映了受测者作为职业人的通用能力综合情况。', reviewStatus: 'pending_review', reviewNote: '' },
  { contentType: 'group', matchKey: 'psychological_quality', levelCode: 'description', content: '“心理素养”由敬业奉献、求真务实、自律性、成就导向、合作意识五个子维度构成，其得分是这五个子维度的综合得分，反映了受测者心理状态的综合情况。', reviewStatus: 'pending_review', reviewNote: '' },
  { contentType: 'validity', matchKey: '', levelCode: 'good', content: '本次测评作答效度良好，结果具有较好的参考价值。测评结果仍应结合实际工作表现、行为观察、访谈及其他评价信息进行综合解读，不宜作为人才决策的唯一依据。', sourceFile: 'AI建议稿-待客户确认', reviewStatus: 'pending_review', reviewNote: 'AI建议稿，待客户及心理测量负责人确认或修改' },
  { contentType: 'validity', matchKey: '', levelCode: 'questionable', content: '该受测者存在掩饰真实想法的可能性，本次测评结果请谨慎解读，必要时可重新开展评估。', reviewStatus: 'pending_review', reviewNote: '来源为Word样例，正式导入前需确认' }
];

function usage() {
  console.log([
    'Usage: node scripts/tools/convert-competency-phase1-workbook.js [options]',
    '',
    'Options:',
    '  --source <xlsx>  Customer workbook path',
    '  --output <json>  Candidate JSON output path',
    '  --import-output <xlsx>  Phase-1 ten-column import workbook path',
    '  --csv-output-dir <dir>  Write four normalized UTF-8 BOM CSV files for manual review',
    '  --csv-only       Write only the normalized CSV files',
    '  --check          Verify the existing JSON and xlsx outputs are exactly reproducible',
    '  --stdout         Print candidate JSON instead of writing a file',
    '  --help           Show this help'
  ].join('\n'));
}

function parseArgs(argv) {
  const options = { source: defaultSource, output: defaultOutput, importOutput: defaultImportOutput, csvOutputDir: '', csvOnly: false, check: false, stdout: false };
  for (let index = 0; index < argv.length; index++) {
    const arg = argv[index];
    if (arg === '--help') {
      usage();
      process.exit(0);
    }
    if (arg === '--check') {
      options.check = true;
      continue;
    }
    if (arg === '--stdout') {
      options.stdout = true;
      continue;
    }
    if (arg === '--csv-only') {
      options.csvOnly = true;
      continue;
    }
    if (arg === '--source' || arg === '--output' || arg === '--import-output' || arg === '--csv-output-dir') {
      const value = argv[++index];
      if (!value) throw new Error(`${arg} requires a path`);
      const key = arg === '--import-output' ? 'importOutput' : arg === '--csv-output-dir' ? 'csvOutputDir' : arg.slice(2);
      options[key] = path.resolve(value);
      continue;
    }
    throw new Error(`unknown argument: ${arg}`);
  }
  if (options.check && options.stdout) throw new Error('--check and --stdout cannot be used together');
  if (options.csvOnly && !options.csvOutputDir) throw new Error('--csv-only requires --csv-output-dir');
  if (options.csvOnly && (options.check || options.stdout)) throw new Error('--csv-only cannot be combined with --check or --stdout');
  return options;
}

function text(value) {
  return value === null || value === undefined ? '' : String(value).trim();
}

function integer(value, context) {
  const normalized = text(value);
  if (!/^\d+$/.test(normalized)) throw new Error(`${context} must be an integer, got ${JSON.stringify(value)}`);
  return Number(normalized);
}

function requireValue(value, context) {
  const normalized = text(value);
  if (!normalized) throw new Error(`${context} is required`);
  return normalized;
}

function requireEqual(actual, expected, context) {
  if (actual !== expected) throw new Error(`${context}=${JSON.stringify(actual)}, want ${JSON.stringify(expected)}`);
}

function requireArrayEqual(actual, expected, context) {
  if (JSON.stringify(actual) !== JSON.stringify(expected)) {
    throw new Error(`${context}=${JSON.stringify(actual)}, want ${JSON.stringify(expected)}`);
  }
}

function rowsFor(workbook, sheetName) {
  const sheet = workbook.Sheets[sheetName];
  if (!sheet) throw new Error(`missing worksheet: ${sheetName}`);
  return {
    ref: sheet['!ref'] || '',
    merges: (sheet['!merges'] || []).length,
    rows: XLSX.utils.sheet_to_json(sheet, { header: 1, defval: null, raw: false })
  };
}

function mappingByName(name) {
  const mapping = dimensionMappings.find(item => item.sourceName === name);
  if (!mapping) throw new Error(`unknown phase-1 dimension: ${name}`);
  return mapping;
}

function mappingFromCell(value, context) {
  const parts = requireValue(value, context).split(/\r?\n/).map(text).filter(Boolean);
  if (parts.length === 1) return mappingByName(parts[0]);
  if (parts.length !== 2) throw new Error(`${context} must contain a dimension name or code + newline + name`);
  const [code, name] = parts;
  const mapping = dimensionMappings.find(item => item.dimensionCode === code);
  if (!mapping) throw new Error(`unknown phase-1 dimension code: ${code}`);
  if (mapping.sourceName !== name) {
    throw new Error(`${context} dimension code/name mismatch: ${code}=${JSON.stringify(name)}, want ${JSON.stringify(mapping.sourceName)}`);
  }
  return mapping;
}

function phaseQuestionCode(mapping, itemNo) {
  return `${mapping.dimensionCode}-Q${String(itemNo).padStart(2, '0')}`;
}

function validityQuestionCode(itemNo) {
  return `P1-VAL-Q${String(itemNo).padStart(2, '0')}`;
}

function parseQuestionSheet(sheetData) {
  const { rows } = sheetData;
  requireEqual(text(rows[0] && rows[0][0]), '基层员工胜任力测评题本', '题本!A1');
  requireArrayEqual((rows[2] || []).slice(4, 9).map(text), expectedOptionLabels, '题本 option labels');

  const questions = [];
  let currentLayer = '';
  let currentDimension = '';
  for (let index = 3; index < rows.length; index++) {
    const source = rows[index] || [];
    const rowNumber = index + 1;
    if (text(source[0])) currentLayer = text(source[0]);
    if (text(source[1])) currentDimension = text(source[1]);
    const content = text(source[3]);
    if (!content) continue;

    const mapping = mappingFromCell(currentDimension, `题本 row ${rowNumber} dimension`);
    const itemNo = integer(source[2], `题本 row ${rowNumber} item number`);
    const observationPoint = requireValue(source[9], `题本 row ${rowNumber} observation point`);
    const sourceDirection = text(source[10]);
    const sourceQuestionType = text(source[11]);
    const isValidity = currentLayer === '效度量表';

    if (isValidity) {
      questions.push({
        candidateQuestionCode: validityQuestionCode(itemNo),
        questionType: 'validity',
        sourceRow: rowNumber,
        sourceLayer: currentLayer,
        associatedDimensionName: mapping.sourceName,
        associatedPhaseDimensionOrder: mapping.phaseOrder,
        dimensionId: mapping.dimensionId,
        dimensionCode: mapping.dimensionCode,
        itemNo,
        content,
        observationPoint,
        direction: 'forward',
        directionSource: 'customer_decision_2026-08-10_all_validity_forward',
        sourceQuestionType: sourceQuestionType || null,
        proposedStatus: 'enabled',
        importStatus: 'ready_for_staging_candidate_import'
      });
      continue;
    }

    if (currentLayer !== mapping.sourceLayer) {
      throw new Error(`题本 row ${rowNumber} layer=${currentLayer}, want ${mapping.sourceLayer}`);
    }
    if (sourceQuestionType !== '维度题') {
      throw new Error(`题本 row ${rowNumber} type=${sourceQuestionType}, want 维度题`);
    }
    if (sourceDirection !== '' && sourceDirection !== '反向') {
      throw new Error(`题本 row ${rowNumber} direction=${sourceDirection}, want blank or 反向`);
    }
    const direction = sourceDirection === '反向' ? 'reverse' : 'forward';
    questions.push({
      candidateQuestionCode: phaseQuestionCode(mapping, itemNo),
      questionType: 'dimension',
      sourceRow: rowNumber,
      sourceLayer: currentLayer,
      firstLevelCode: mapping.firstLevelCode,
      firstLevelName: mapping.firstLevelName,
      phaseDimensionOrder: mapping.phaseOrder,
      sourceDimensionName: mapping.sourceName,
      dimensionId: mapping.dimensionId,
      dimensionCode: mapping.dimensionCode,
      itemNo,
      content,
      observationPoint,
      direction,
      directionSource: sourceDirection === '反向' ? 'explicit_reverse' : 'blank_interpreted_as_forward',
      sourceQuestionType,
      proposedStatus: 'enabled',
      importStatus: 'ready_for_staging_candidate_import'
    });
  }
  return questions;
}

function parseDimensionContentSheet(sheetData) {
  const { rows } = sheetData;
  requireEqual(text(rows[0] && rows[0][0]), '基层员工胜任力测评维度评价', '等级评价!A1');
  const contents = [];
  let currentLayer = '';
  for (let index = 3; index < rows.length; index++) {
    const source = rows[index] || [];
    if (text(source[0])) currentLayer = text(source[0]);
    const dimensionCell = text(source[2]);
    if (!dimensionCell) continue;
    const rowNumber = index + 1;
    const mapping = mappingFromCell(dimensionCell, `等级评价 row ${rowNumber} dimension`);
    const sourceOrder = integer(source[1], `等级评价 row ${rowNumber} order`);
    if (sourceOrder !== mapping.phaseOrder) {
      throw new Error(`等级评价 row ${rowNumber} order=${sourceOrder}, want ${mapping.phaseOrder}`);
    }
    if (currentLayer !== mapping.sourceLayer) {
      throw new Error(`等级评价 row ${rowNumber} layer=${currentLayer}, want ${mapping.sourceLayer}`);
    }
    const levels = levelDefinitions.map((level, levelIndex) => ({
      ...level,
      content: requireValue(source[5 + levelIndex], `等级评价 row ${rowNumber} ${level.code} content`)
    }));
    contents.push({
      phaseOrder: mapping.phaseOrder,
      firstLevelCode: mapping.firstLevelCode,
      firstLevelName: mapping.firstLevelName,
      sourceLayer: currentLayer,
      sourceDimensionName: mapping.sourceName,
      dimensionId: mapping.dimensionId,
      dimensionCode: mapping.dimensionCode,
      identityStatus: mapping.identityStatus,
      coreMeaning: requireValue(source[3], `等级评价 row ${rowNumber} core meaning`),
      definition: requireValue(source[4], `等级评价 row ${rowNumber} definition`),
      levels
    });
  }
  return contents;
}

function parseOverallContentSheet(sheetData) {
  const { rows } = sheetData;
  requireEqual(text(rows[0] && rows[0][0]), '基层员工胜任力测评总体评价', '总体评价!A1');
  const contents = [];
  for (let index = 2; index < rows.length; index++) {
    const source = rows[index] || [];
    const levelName = text(source[0]);
    if (!levelName) continue;
    contents.push({
      rank: contents.length + 1,
      levelName,
      sourceRange: requireValue(source[1], `总体评价 row ${index + 1} range`),
      content: requireValue(source[2], `总体评价 row ${index + 1} content`)
    });
  }
  return contents;
}

function validateCandidate(candidate) {
  const questions = candidate.questions;
  const dimensionQuestions = questions.filter(item => item.questionType === 'dimension');
  const validityQuestions = questions.filter(item => item.questionType === 'validity');
  requireEqual(questions.length, 90, 'total question count');
  requireEqual(dimensionQuestions.length, 80, 'dimension question count');
  requireEqual(validityQuestions.length, 10, 'validity question count');
  requireEqual(new Set(questions.map(item => item.candidateQuestionCode)).size, 90, 'unique candidate question codes');
  requireEqual(new Set(questions.map(item => item.content)).size, 90, 'unique question contents');
  requireEqual(dimensionQuestions.filter(item => item.direction === 'forward').length, 62, 'forward question count');
  requireEqual(dimensionQuestions.filter(item => item.direction === 'reverse').length, 18, 'reverse question count');
  requireEqual(validityQuestions.filter(item => item.direction === null).length, 0, 'unresolved validity direction count');
  requireEqual(validityQuestions.filter(item => item.direction === 'forward').length, 10, 'forward validity question count');
  requireEqual(candidate.dimensionContent.length, 10, 'dimension content count');
  requireEqual(candidate.dimensionContent.reduce((count, item) => count + item.levels.length, 0), 50, 'dimension level content count');
  requireEqual(candidate.overallContent.length, 5, 'overall content count');
  requireEqual(candidate.dimensions.filter(item => item.identityStatus === 'confirmed_phase1').length, 10, 'confirmed dimension identity count');

  for (const mapping of dimensionMappings) {
    const rows = dimensionQuestions.filter(item => item.phaseDimensionOrder === mapping.phaseOrder);
    requireEqual(rows.length, 8, `${mapping.sourceName} question count`);
    requireArrayEqual(rows.map(item => item.itemNo), [1, 2, 3, 4, 5, 6, 7, 8], `${mapping.sourceName} item numbers`);
    const reverse = rows.filter(item => item.direction === 'reverse').map(item => item.itemNo);
    if (mapping.sourceName === '成就导向') {
      requireArrayEqual(reverse, [], `${mapping.sourceName} reverse items`);
    } else {
      requireArrayEqual(reverse, [7, 8], `${mapping.sourceName} reverse items`);
    }
  }
}

function buildCandidate(sourcePath) {
  if (!fs.existsSync(sourcePath)) throw new Error(`source workbook not found: ${sourcePath}`);
  if (path.extname(sourcePath).toLowerCase() !== '.xlsx') throw new Error('source workbook must be .xlsx');
  const sourceBytes = fs.readFileSync(sourcePath);
  const sourceSha256 = crypto.createHash('sha256').update(sourceBytes).digest('hex');
  const workbook = XLSX.read(sourceBytes, { type: 'buffer', raw: false });
  requireArrayEqual(workbook.SheetNames, ['题本', '等级评价', '总体评价'], 'worksheet names');

  const questionSheet = rowsFor(workbook, '题本');
  const dimensionContentSheet = rowsFor(workbook, '等级评价');
  const overallContentSheet = rowsFor(workbook, '总体评价');
  const questions = parseQuestionSheet(questionSheet);
  const dimensionContent = parseDimensionContentSheet(dimensionContentSheet);
  const overallContent = parseOverallContentSheet(overallContentSheet);

  const candidate = {
    schemaVersion: SCHEMA_VERSION,
    materialDate: MATERIAL_DATE,
    converter: {
      name: path.basename(__filename),
      version: TOOL_VERSION,
      deterministicOutput: true
    },
    source: {
      fileName: path.basename(sourcePath),
      sha256: sourceSha256,
      worksheets: [
        { name: '题本', ref: questionSheet.ref, mergedRangeCount: questionSheet.merges },
        { name: '等级评价', ref: dimensionContentSheet.ref, mergedRangeCount: dimensionContentSheet.merges },
        { name: '总体评价', ref: overallContentSheet.ref, mergedRangeCount: overallContentSheet.merges }
      ]
    },
    readiness: {
      candidateGenerated: true,
      importReady: true,
      databaseWritten: false,
      blockingIssueIds: [],
      warningIssueIds: []
    },
    fixedProfile: {
      profileCode: 'frontline-phase1-v1',
      profileName: '基层员工胜任力测评第一期',
      reportAudience: 'frontline_employee',
      fixedDimensionCount: 10,
      dimensionQuestionCount: 80,
      validityQuestionCount: 10,
      totalQuestionCount: 90,
      optionScale: expectedOptionLabels.map((label, index) => ({ rawValue: index + 1, label }))
    },
    versions: confirmedVersions,
    approval: {
      decisionDate: '2026-08-10',
      decisionDocument: 'docs/00401-phase1-customer-decisions-20260810.md',
      stagingCandidateImportApproved: true,
      productionContentApproved: false
    },
    firstLevelGroups: [
      { code: 'general_ability', name: '通用能力', phaseOrder: 1, childPhaseOrders: [1, 2, 3, 4, 5] },
      { code: 'psychological_quality', name: '心理素养', phaseOrder: 2, childPhaseOrders: [6, 7, 8, 9, 10] }
    ],
    dimensions: dimensionMappings,
    questionSummary: {
      total: questions.length,
      dimension: questions.filter(item => item.questionType === 'dimension').length,
      validity: questions.filter(item => item.questionType === 'validity').length,
      forward: questions.filter(item => item.questionType === 'dimension' && item.direction === 'forward').length,
      reverse: questions.filter(item => item.questionType === 'dimension' && item.direction === 'reverse').length,
      validityForward: questions.filter(item => item.questionType === 'validity' && item.direction === 'forward').length,
      unresolvedValidityDirections: questions.filter(item => item.questionType === 'validity' && item.direction === null).length
    },
    questions,
    dimensionContent,
    overallContent,
    issues: [],
    resolvedIssues: [
      { id: 'VAL-001', resolution: '10道效度题正式标记为validity' },
      { id: 'VAL-002', resolution: '10道效度题全部按原始1至5分正向累计' },
      { id: 'VAL-003', resolution: '不创建伪装维度；A/B关联仅用于来源追踪' },
      { id: 'STATUS-001', resolution: '90道题全部启用' },
      { id: 'VERSION-001', resolution: '使用已确认的四类一期v1版本' },
      { id: 'DIR-001', resolution: 'B1-04成就导向8题保持全部正向' },
      { id: 'LABEL-001', resolution: '一级维度L3统一为中分' },
      { id: 'BOUNDARY-001', resolution: '采用客户决策基线中的精确边界' }
    ]
  };
  validateCandidate(candidate);
  return candidate;
}

function serialize(candidate) {
  return `${JSON.stringify(candidate, null, 2)}\n`;
}

function buildImportWorkbook(candidate) {
  const rows = [importHeaders];
  for (const question of candidate.questions) {
    const order = question.questionType === 'dimension'
      ? question.phaseDimensionOrder
      : question.associatedPhaseDimensionOrder;
    const dimension = candidate.dimensions.find(item => item.phaseOrder === order);
    rows.push([
      order,
      dimension.sourceName,
      question.questionType === 'validity' ? '效度题' : '维度题',
      question.candidateQuestionCode,
      question.itemNo,
      question.content,
      question.observationPoint,
      question.direction === 'reverse' ? '反向' : '正向',
      '启用',
      '一期客户题本-20260807-已确认staging候选'
    ]);
  }
  const workbook = XLSX.utils.book_new();
  workbook.Props = {
    Title: '00401基层员工胜任力测评第一期导入题本',
    Author: 'Talent Assessment',
    CreatedDate: new Date('2026-08-10T00:00:00.000Z'),
    ModifiedDate: new Date('2026-08-10T00:00:00.000Z')
  };
  const sheet = XLSX.utils.aoa_to_sheet(rows);
  sheet['!cols'] = [12, 14, 10, 18, 12, 48, 24, 10, 10, 36].map(width => ({ wch: width }));
  XLSX.utils.book_append_sheet(workbook, sheet, '胜任力题目');
  return XLSX.write(workbook, { type: 'buffer', bookType: 'xlsx', compression: true });
}

function csvPayload(rows) {
  const content = rows.map(row => row.map(value => `"${String(value === null || value === undefined ? '' : value).replace(/"/g, '""')}"`).join(',')).join('\r\n');
  return Buffer.from(`\ufeff${content}\r\n`, 'utf8');
}

function buildNormalizedCSVs(candidate) {
  const byOrder = new Map(candidate.dimensions.map(item => [item.phaseOrder, item]));
  const contentByOrder = new Map(candidate.dimensionContent.map(item => [item.phaseOrder, item]));
  const reportTemplateSHA256 = fs.existsSync(reportTemplatePath)
    ? crypto.createHash('sha256').update(fs.readFileSync(reportTemplatePath)).digest('hex')
    : '';
  const packageRows = [['契约版本', '产品版本', '评分版本', '内容版本', '报告模板版本', '报告对象', '原始题本文件', '原始题本SHA256', '正式内容源SHA256', '报告模板文件', '报告模板SHA256', '批准状态', '内容批准人', '内容批准时间', '测量批准人', '测量批准时间', '生效环境', '最终免责声明'], [
    CSV_CONTRACT_VERSION, candidate.versions.product, candidate.versions.scoring,
    candidate.versions.content, candidate.versions.reportTemplate,
    candidate.fixedProfile.reportAudience, candidate.source.fileName, candidate.source.sha256,
    '', REPORT_TEMPLATE_NAME, reportTemplateSHA256, 'draft', '', '', '', '', '', customerSpecialNotice
  ]];

  const questionRows = [['顺序', '来源行', '题目编号', '题目类型', '一级维度编号', '一级维度名称', '维度ID', '维度编号', '维度名称', '维度内题号', '题目内容', '考察点', '计分方向', '启用状态', '审核状态', '审核备注']];
  candidate.questions.forEach((question, index) => {
    const order = question.questionType === 'dimension' ? question.phaseDimensionOrder : question.associatedPhaseDimensionOrder;
    const dimension = byOrder.get(order);
    questionRows.push([
      index + 1, question.sourceRow, question.candidateQuestionCode, question.questionType,
      dimension.firstLevelCode, dimension.firstLevelName, dimension.dimensionId,
      dimension.dimensionCode, dimension.sourceName, question.itemNo, question.content,
      question.observationPoint, question.direction, question.proposedStatus, 'pending_review', ''
    ]);
  });

  const dimensionRows = [['顺序', '一级维度编号', '一级维度名称', '维度ID', '维度编号', '维度名称', '核心含义', '定义', '启用状态', '审核状态', '审核备注']];
  const levelRows = [['顺序', '维度ID', '维度编号', '维度名称', '等级编号', '二级标签', '一级标签', '最小分', '包含最小分', '最大分', '包含最大分', '显示区间', '诊断与建议', '审核状态', '审核备注']];
  for (const dimension of candidate.dimensions) {
    const content = contentByOrder.get(dimension.phaseOrder);
    dimensionRows.push([
      dimension.phaseOrder, dimension.firstLevelCode, dimension.firstLevelName,
      dimension.dimensionId, dimension.dimensionCode, dimension.sourceName,
      content.coreMeaning, content.definition, 'enabled', 'pending_review', ''
    ]);
    for (const level of content.levels) {
      const boundary = dimensionBounds.find(item => item.code === level.code);
      levelRows.push([
        dimension.phaseOrder, dimension.dimensionId, dimension.dimensionCode,
        dimension.sourceName, level.code, level.secondaryLabel, level.firstLevelLabel,
        boundary.minScore, boundary.minInclusive, boundary.maxScore, boundary.maxInclusive,
        level.sourceRange, level.content, 'pending_review', ''
      ]);
    }
  }

  const overallRows = [['排名', '总体等级编号', '总体等级', '最小分', '包含最小分', '最大分', '包含最大分', '显示区间', '诊断', '审核状态', '审核备注']];
  for (const definition of overallDefinitions) {
    const item = candidate.overallContent.find(content => content.levelName === definition.name);
    if (!item || item.sourceRange !== definition.sourceRange) {
      throw new Error(`overall definition mismatch: ${definition.code}`);
    }
    overallRows.push([
      item.rank, definition.code, item.levelName, definition.minScore, definition.minInclusive,
      definition.maxScore, definition.maxInclusive, item.sourceRange, item.content, 'pending_review', ''
    ]);
  }

  const staticTextRows = [['内容类型', '匹配键', '等级编号', '内容', '来源文件', '审核状态', '审核备注']];
  for (const item of reportStaticTexts) {
    staticTextRows.push([item.contentType, item.matchKey, item.levelCode, item.content, item.sourceFile || REPORT_TEMPLATE_NAME, item.reviewStatus, item.reviewNote]);
  }
  return new Map([
    ['phase1-package.csv', csvPayload(packageRows)],
    ['phase1-questions.csv', csvPayload(questionRows)],
    ['phase1-dimensions.csv', csvPayload(dimensionRows)],
    ['phase1-dimension-levels.csv', csvPayload(levelRows)],
    ['phase1-overall-levels.csv', csvPayload(overallRows)],
    ['phase1-report-static-texts.csv', csvPayload(staticTextRows)]
  ]);
}

function writeNormalizedCSVs(outputDirectory, payloads) {
  fs.mkdirSync(outputDirectory, { recursive: true });
  for (const [fileName, payload] of payloads) {
    fs.writeFileSync(path.join(outputDirectory, fileName), payload);
  }
}

function report(candidate, outputPath, importOutputPath, importPayload, action, csvOutputDir = '') {
  const payload = serialize(candidate);
  const outputSha256 = crypto.createHash('sha256').update(payload).digest('hex');
  const importSha256 = crypto.createHash('sha256').update(importPayload).digest('hex');
  console.log('COMPETENCY_PHASE1_CANDIDATE_OK');
  console.log(`action=${action}`);
  console.log(`source_sha256=${candidate.source.sha256}`);
  console.log(`candidate_sha256=${outputSha256}`);
  console.log(`questions=${candidate.questionSummary.total}`);
  console.log(`dimension_questions=${candidate.questionSummary.dimension}`);
  console.log(`validity_questions=${candidate.questionSummary.validity}`);
  console.log(`forward=${candidate.questionSummary.forward}`);
  console.log(`reverse=${candidate.questionSummary.reverse}`);
  console.log(`unresolved_validity_directions=${candidate.questionSummary.unresolvedValidityDirections}`);
  console.log(`validity_forward=${candidate.questionSummary.validityForward}`);
  console.log(`dimension_level_texts=${candidate.dimensionContent.reduce((count, item) => count + item.levels.length, 0)}`);
  console.log(`overall_texts=${candidate.overallContent.length}`);
  console.log(`blockers=${candidate.readiness.blockingIssueIds.length}`);
  console.log(`warnings=${candidate.readiness.warningIssueIds.length}`);
  if (outputPath) console.log(`output=${outputPath}`);
  console.log(`import_sha256=${importSha256}`);
  if (importOutputPath) console.log(`import_output=${importOutputPath}`);
  if (csvOutputDir) console.log(`csv_output_dir=${csvOutputDir}`);
}

function main() {
  const options = parseArgs(process.argv.slice(2));
  const candidate = buildCandidate(options.source);
  const payload = serialize(candidate);
  const importPayload = buildImportWorkbook(candidate);
  const csvPayloads = options.csvOutputDir ? buildNormalizedCSVs(candidate) : new Map();

  if (options.stdout) {
    process.stdout.write(payload);
    return;
  }
  if (options.check) {
    if (!fs.existsSync(options.output)) throw new Error(`candidate output not found: ${options.output}`);
    const existing = fs.readFileSync(options.output, 'utf8');
    if (existing !== payload) throw new Error('candidate output is stale or was modified; regenerate it');
    if (!fs.existsSync(options.importOutput)) throw new Error(`import workbook not found: ${options.importOutput}`);
    const existingImport = fs.readFileSync(options.importOutput);
    if (!existingImport.equals(importPayload)) throw new Error('import workbook is stale or was modified; regenerate it');
    report(candidate, options.output, options.importOutput, importPayload, 'check');
    return;
  }

  if (options.csvOnly) {
    writeNormalizedCSVs(options.csvOutputDir, csvPayloads);
    report(candidate, '', '', importPayload, 'csv-only', options.csvOutputDir);
    return;
  }

  fs.mkdirSync(path.dirname(options.output), { recursive: true });
  fs.mkdirSync(path.dirname(options.importOutput), { recursive: true });
  fs.writeFileSync(options.output, payload, 'utf8');
  fs.writeFileSync(options.importOutput, importPayload);
  if (options.csvOutputDir) writeNormalizedCSVs(options.csvOutputDir, csvPayloads);
  report(candidate, options.output, options.importOutput, importPayload, 'write', options.csvOutputDir);
}

try {
  main();
} catch (error) {
  console.error(`COMPETENCY_PHASE1_CANDIDATE_FAILED: ${error.message}`);
  process.exitCode = 1;
}
