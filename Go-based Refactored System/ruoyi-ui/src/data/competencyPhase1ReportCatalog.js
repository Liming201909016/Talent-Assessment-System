// Generated from scripts/data/competency-phase1-csv by generate-competency-phase1-report-catalog.js.
// Do not edit manually; update the validated CSV package and regenerate.
export default {
  contractVersion: 'competency-phase1-csv-v1',
  productVersion: 'competency-frontline-phase1-v1',
  scoringVersion: 'competency-phase1-scoring-v1',
  contentVersion: 'competency-phase1-content-v1',
  reportTemplateVersion: 'competency-phase1-report-v1',
  audience: 'frontline_employee',
  fixedTexts: {
    coverTitle: '胜任力测评报告',
    coverEnglish: 'Competency Assessment Report',
    coverSlogan: '洞悉深层素质 ● 驱动精准发展',
    readingBackground: '在组织人才竞争日益激烈的当下，科学评估员工的岗位胜任力水平已成为企业人力资源管理的关键环节。无数研究与实践证明，员工绩效差异往往并非仅由知识储备与基础技能的高低所决定，而是更多源于更深层的素质——包括能力倾向、价值取向、心理动力，这些深层特征构成了区分绩优者与一般者的员工胜任力，是本测评关注的焦点。',
    readingDimensions: '本测评以麦克利兰（McClelland）的冰山模型与博亚特兹（Boyatzis）的洋葱模型为理论基础，聚焦鉴别性素质的评估，从通用能力、心理素养两个方面共计10个维度对员工进行系统考察，旨在全面评估影响工作表现的关键因素。其中，通用能力包括逻辑思维、数字应用、计划执行、持续学习、沟通表达5个子维度，心理素养包括敬业奉献、求真务实、自律性、成就导向和合作意识5个子维度。',
    readingUsage: '本测评能够有效鉴别员工的能力短板与发展优势，可以应用于组织的招聘、选拔、在岗评估、培训发展与人才盘点等，为组织客观量化人才评价与个性化人才培养提供数据支撑，辅助实现人岗匹配与人才发展战略的制定与落地。',
    specialNotice: '科学测评的原则是多质多法，即对每个素质的评价采用多种测试手段，本测评作为一种测试手段，虽证明有效，其结果基于受测者自陈反应结果，受其测评状态、自我认知偏差等影响，不能作为精准判断个体岗位胜任力的唯一依据。更精准的测试建议根据实际情况，结合其他测试结果，如面试、绩效表现、情景模拟、360评估等进行综合判断。'
  },
  groups: [
    {
      code: 'general_ability',
      name: '通用能力',
      description: '“通用能力”由逻辑思维、数字应用、计划执行、持续学习、沟通表达五个子维度构成，其得分是这五个子维度的综合得分，反映了受测者作为职业人的通用能力综合情况。'
    },
    {
      code: 'psychological_quality',
      name: '心理素养',
      description: '“心理素养”由敬业奉献、求真务实、自律性、成就导向、合作意识五个子维度构成，其得分是这五个子维度的综合得分，反映了受测者心理状态的综合情况。'
    }
  ],
  dimensions: [
    {
      order: 1,
      groupCode: 'general_ability',
      groupName: '通用能力',
      id: 'competency-a1-01',
      code: 'A1-01',
      name: '逻辑思维',
      coreMeaning: '逻辑分析严谨，推理判断有据',
      definition: '运用归纳、演绎等逻辑方法分析问题，能对信息进行分类整合和概念辨析，推理过程条理清晰、环环相扣，能构建有充分依据的判断和结论。',
      levels: [
        {
          code: 'L1',
          secondaryLabel: '差',
          groupLabel: '低分',
          minScore: 1,
          minInclusive: true,
          maxScore: 1.7,
          maxInclusive: true,
          displayRange: '1.0—1.7分'
        },
        {
          code: 'L2',
          secondaryLabel: '较差',
          groupLabel: '较低分',
          minScore: 1.7,
          minInclusive: false,
          maxScore: 2.7,
          maxInclusive: true,
          displayRange: '1.7—2.7分'
        },
        {
          code: 'L3',
          secondaryLabel: '合格',
          groupLabel: '中分',
          minScore: 2.7,
          minInclusive: false,
          maxScore: 3.5,
          maxInclusive: true,
          displayRange: '2.7—3.5分'
        },
        {
          code: 'L4',
          secondaryLabel: '较优秀',
          groupLabel: '较高分',
          minScore: 3.5,
          minInclusive: false,
          maxScore: 4.3,
          maxInclusive: true,
          displayRange: '3.5—4.3分'
        },
        {
          code: 'L5',
          secondaryLabel: '优秀',
          groupLabel: '高分',
          minScore: 4.3,
          minInclusive: false,
          maxScore: 5,
          maxInclusive: true,
          displayRange: '4.3—5.0分'
        }
      ]
    },
    {
      order: 2,
      groupCode: 'general_ability',
      groupName: '通用能力',
      id: 'competency-a1-02',
      code: 'A1-02',
      name: '数字应用',
      coreMeaning: '善用数字化工具与AI技术，具备数据思维',
      definition: '熟练运用各类数字化工具和AI技术处理工作事务，具备数据思维与人机协作意识，能够借助新技术手段提升工作效率和决策质量。',
      levels: [
        {
          code: 'L1',
          secondaryLabel: '差',
          groupLabel: '低分',
          minScore: 1,
          minInclusive: true,
          maxScore: 1.7,
          maxInclusive: true,
          displayRange: '1.0—1.7分'
        },
        {
          code: 'L2',
          secondaryLabel: '较差',
          groupLabel: '较低分',
          minScore: 1.7,
          minInclusive: false,
          maxScore: 2.7,
          maxInclusive: true,
          displayRange: '1.7—2.7分'
        },
        {
          code: 'L3',
          secondaryLabel: '合格',
          groupLabel: '中分',
          minScore: 2.7,
          minInclusive: false,
          maxScore: 3.5,
          maxInclusive: true,
          displayRange: '2.7—3.5分'
        },
        {
          code: 'L4',
          secondaryLabel: '较优秀',
          groupLabel: '较高分',
          minScore: 3.5,
          minInclusive: false,
          maxScore: 4.3,
          maxInclusive: true,
          displayRange: '3.5—4.3分'
        },
        {
          code: 'L5',
          secondaryLabel: '优秀',
          groupLabel: '高分',
          minScore: 4.3,
          minInclusive: false,
          maxScore: 5,
          maxInclusive: true,
          displayRange: '4.3—5.0分'
        }
      ]
    },
    {
      order: 3,
      groupCode: 'general_ability',
      groupName: '通用能力',
      id: 'competency-a1-03',
      code: 'A1-03',
      name: '计划执行',
      coreMeaning: '高效推进计划并达成预期结果',
      definition: '接到任务后迅速响应、立即行动，紧盯时间节点，在保证质量的前提下追求效率，确保按时交付成果。',
      levels: [
        {
          code: 'L1',
          secondaryLabel: '差',
          groupLabel: '低分',
          minScore: 1,
          minInclusive: true,
          maxScore: 1.7,
          maxInclusive: true,
          displayRange: '1.0—1.7分'
        },
        {
          code: 'L2',
          secondaryLabel: '较差',
          groupLabel: '较低分',
          minScore: 1.7,
          minInclusive: false,
          maxScore: 2.7,
          maxInclusive: true,
          displayRange: '1.7—2.7分'
        },
        {
          code: 'L3',
          secondaryLabel: '合格',
          groupLabel: '中分',
          minScore: 2.7,
          minInclusive: false,
          maxScore: 3.5,
          maxInclusive: true,
          displayRange: '2.7—3.5分'
        },
        {
          code: 'L4',
          secondaryLabel: '较优秀',
          groupLabel: '较高分',
          minScore: 3.5,
          minInclusive: false,
          maxScore: 4.3,
          maxInclusive: true,
          displayRange: '3.5—4.3分'
        },
        {
          code: 'L5',
          secondaryLabel: '优秀',
          groupLabel: '高分',
          minScore: 4.3,
          minInclusive: false,
          maxScore: 5,
          maxInclusive: true,
          displayRange: '4.3—5.0分'
        }
      ]
    },
    {
      order: 4,
      groupCode: 'general_ability',
      groupName: '通用能力',
      id: 'competency-a1-04',
      code: 'A1-04',
      name: '持续学习',
      coreMeaning: '主动学习，多渠道获取知识并学以致用',
      definition: '保持学习新知识、新技能的意愿，善于通过观察、请教和实践等多种途径获取知识，并将所学有效运用到实际工作中。',
      levels: [
        {
          code: 'L1',
          secondaryLabel: '差',
          groupLabel: '低分',
          minScore: 1,
          minInclusive: true,
          maxScore: 1.7,
          maxInclusive: true,
          displayRange: '1.0—1.7分'
        },
        {
          code: 'L2',
          secondaryLabel: '较差',
          groupLabel: '较低分',
          minScore: 1.7,
          minInclusive: false,
          maxScore: 2.7,
          maxInclusive: true,
          displayRange: '1.7—2.7分'
        },
        {
          code: 'L3',
          secondaryLabel: '合格',
          groupLabel: '中分',
          minScore: 2.7,
          minInclusive: false,
          maxScore: 3.5,
          maxInclusive: true,
          displayRange: '2.7—3.5分'
        },
        {
          code: 'L4',
          secondaryLabel: '较优秀',
          groupLabel: '较高分',
          minScore: 3.5,
          minInclusive: false,
          maxScore: 4.3,
          maxInclusive: true,
          displayRange: '3.5—4.3分'
        },
        {
          code: 'L5',
          secondaryLabel: '优秀',
          groupLabel: '高分',
          minScore: 4.3,
          minInclusive: false,
          maxScore: 5,
          maxInclusive: true,
          displayRange: '4.3—5.0分'
        }
      ]
    },
    {
      order: 5,
      groupCode: 'general_ability',
      groupName: '通用能力',
      id: 'competency-a1-05',
      code: 'A1-05',
      name: '沟通表达',
      coreMeaning: '清晰传递信息，重视倾听与反馈',
      definition: '用准确恰当的语言传递信息，表达观点时条理清晰、重点突出，同时重视倾听与反馈，保证信息传达到位、理解准确。',
      levels: [
        {
          code: 'L1',
          secondaryLabel: '差',
          groupLabel: '低分',
          minScore: 1,
          minInclusive: true,
          maxScore: 1.7,
          maxInclusive: true,
          displayRange: '1.0—1.7分'
        },
        {
          code: 'L2',
          secondaryLabel: '较差',
          groupLabel: '较低分',
          minScore: 1.7,
          minInclusive: false,
          maxScore: 2.7,
          maxInclusive: true,
          displayRange: '1.7—2.7分'
        },
        {
          code: 'L3',
          secondaryLabel: '合格',
          groupLabel: '中分',
          minScore: 2.7,
          minInclusive: false,
          maxScore: 3.5,
          maxInclusive: true,
          displayRange: '2.7—3.5分'
        },
        {
          code: 'L4',
          secondaryLabel: '较优秀',
          groupLabel: '较高分',
          minScore: 3.5,
          minInclusive: false,
          maxScore: 4.3,
          maxInclusive: true,
          displayRange: '3.5—4.3分'
        },
        {
          code: 'L5',
          secondaryLabel: '优秀',
          groupLabel: '高分',
          minScore: 4.3,
          minInclusive: false,
          maxScore: 5,
          maxInclusive: true,
          displayRange: '4.3—5.0分'
        }
      ]
    },
    {
      order: 6,
      groupCode: 'psychological_quality',
      groupName: '心理素养',
      id: 'competency-b1-01',
      code: 'B1-01',
      name: '敬业奉献',
      coreMeaning: '视工作为使命，全心投入，甘于奉献',
      definition: '将工作视为使命，以高度的责任感和奉献精神对待岗位职责，不计较分内分外，愿意为团队和组织目标付出额外努力。',
      levels: [
        {
          code: 'L1',
          secondaryLabel: '差',
          groupLabel: '低分',
          minScore: 1,
          minInclusive: true,
          maxScore: 1.7,
          maxInclusive: true,
          displayRange: '1.0—1.7分'
        },
        {
          code: 'L2',
          secondaryLabel: '较差',
          groupLabel: '较低分',
          minScore: 1.7,
          minInclusive: false,
          maxScore: 2.7,
          maxInclusive: true,
          displayRange: '1.7—2.7分'
        },
        {
          code: 'L3',
          secondaryLabel: '合格',
          groupLabel: '中分',
          minScore: 2.7,
          minInclusive: false,
          maxScore: 3.5,
          maxInclusive: true,
          displayRange: '2.7—3.5分'
        },
        {
          code: 'L4',
          secondaryLabel: '较优秀',
          groupLabel: '较高分',
          minScore: 3.5,
          minInclusive: false,
          maxScore: 4.3,
          maxInclusive: true,
          displayRange: '3.5—4.3分'
        },
        {
          code: 'L5',
          secondaryLabel: '优秀',
          groupLabel: '高分',
          minScore: 4.3,
          minInclusive: false,
          maxScore: 5,
          maxInclusive: true,
          displayRange: '4.3—5.0分'
        }
      ]
    },
    {
      order: 7,
      groupCode: 'psychological_quality',
      groupName: '心理素养',
      id: 'competency-b1-02',
      code: 'B1-02',
      name: '求真务实',
      coreMeaning: '追求真理，尊重事实，注重实效',
      definition: '坚持追求事物的客观真相，尊重事实和数据，注重实际效果，以解决问题为根本导向，反对形式主义和空谈。',
      levels: [
        {
          code: 'L1',
          secondaryLabel: '差',
          groupLabel: '低分',
          minScore: 1,
          minInclusive: true,
          maxScore: 1.7,
          maxInclusive: true,
          displayRange: '1.0—1.7分'
        },
        {
          code: 'L2',
          secondaryLabel: '较差',
          groupLabel: '较低分',
          minScore: 1.7,
          minInclusive: false,
          maxScore: 2.7,
          maxInclusive: true,
          displayRange: '1.7—2.7分'
        },
        {
          code: 'L3',
          secondaryLabel: '合格',
          groupLabel: '中分',
          minScore: 2.7,
          minInclusive: false,
          maxScore: 3.5,
          maxInclusive: true,
          displayRange: '2.7—3.5分'
        },
        {
          code: 'L4',
          secondaryLabel: '较优秀',
          groupLabel: '较高分',
          minScore: 3.5,
          minInclusive: false,
          maxScore: 4.3,
          maxInclusive: true,
          displayRange: '3.5—4.3分'
        },
        {
          code: 'L5',
          secondaryLabel: '优秀',
          groupLabel: '高分',
          minScore: 4.3,
          minInclusive: false,
          maxScore: 5,
          maxInclusive: true,
          displayRange: '4.3—5.0分'
        }
      ]
    },
    {
      order: 8,
      groupCode: 'psychological_quality',
      groupName: '心理素养',
      id: 'competency-b1-03',
      code: 'B1-03',
      name: '自律性',
      coreMeaning: '自我约束，规划在先，言行一致',
      definition: '能够自我约束、严格要求自己，做事有事先的规划和控制，有清晰的个人标准并以此规范行为，即使在无人监督时也能坚持按计划行事。',
      levels: [
        {
          code: 'L1',
          secondaryLabel: '差',
          groupLabel: '低分',
          minScore: 1,
          minInclusive: true,
          maxScore: 1.7,
          maxInclusive: true,
          displayRange: '1.0—1.7分'
        },
        {
          code: 'L2',
          secondaryLabel: '较差',
          groupLabel: '较低分',
          minScore: 1.7,
          minInclusive: false,
          maxScore: 2.7,
          maxInclusive: true,
          displayRange: '1.7—2.7分'
        },
        {
          code: 'L3',
          secondaryLabel: '合格',
          groupLabel: '中分',
          minScore: 2.7,
          minInclusive: false,
          maxScore: 3.5,
          maxInclusive: true,
          displayRange: '2.7—3.5分'
        },
        {
          code: 'L4',
          secondaryLabel: '较优秀',
          groupLabel: '较高分',
          minScore: 3.5,
          minInclusive: false,
          maxScore: 4.3,
          maxInclusive: true,
          displayRange: '3.5—4.3分'
        },
        {
          code: 'L5',
          secondaryLabel: '优秀',
          groupLabel: '高分',
          minScore: 4.3,
          minInclusive: false,
          maxScore: 5,
          maxInclusive: true,
          displayRange: '4.3—5.0分'
        }
      ]
    },
    {
      order: 9,
      groupCode: 'psychological_quality',
      groupName: '心理素养',
      id: 'competency-b1-04',
      code: 'B1-04',
      name: '成就导向',
      coreMeaning: '追求工作成功，不断挑战更高目标',
      definition: '有强烈的追求工作成功的愿望，不断设定挑战性的目标挑战自我，关注自身职业生涯的发展，追求事业的成功和卓越。',
      levels: [
        {
          code: 'L1',
          secondaryLabel: '差',
          groupLabel: '低分',
          minScore: 1,
          minInclusive: true,
          maxScore: 1.7,
          maxInclusive: true,
          displayRange: '1.0—1.7分'
        },
        {
          code: 'L2',
          secondaryLabel: '较差',
          groupLabel: '较低分',
          minScore: 1.7,
          minInclusive: false,
          maxScore: 2.7,
          maxInclusive: true,
          displayRange: '1.7—2.7分'
        },
        {
          code: 'L3',
          secondaryLabel: '合格',
          groupLabel: '中分',
          minScore: 2.7,
          minInclusive: false,
          maxScore: 3.5,
          maxInclusive: true,
          displayRange: '2.7—3.5分'
        },
        {
          code: 'L4',
          secondaryLabel: '较优秀',
          groupLabel: '较高分',
          minScore: 3.5,
          minInclusive: false,
          maxScore: 4.3,
          maxInclusive: true,
          displayRange: '3.5—4.3分'
        },
        {
          code: 'L5',
          secondaryLabel: '优秀',
          groupLabel: '高分',
          minScore: 4.3,
          minInclusive: false,
          maxScore: 5,
          maxInclusive: true,
          displayRange: '4.3—5.0分'
        }
      ]
    },
    {
      order: 10,
      groupCode: 'psychological_quality',
      groupName: '心理素养',
      id: 'competency-b1-05',
      code: 'B1-05',
      name: '合作意识',
      coreMeaning: '主动协作，乐于分享，促成共赢',
      definition: '能主动配合他人开展工作，乐于分享信息和资源，妥善处理分歧，及时补位，与团队成员共同推动目标实现。',
      levels: [
        {
          code: 'L1',
          secondaryLabel: '差',
          groupLabel: '低分',
          minScore: 1,
          minInclusive: true,
          maxScore: 1.7,
          maxInclusive: true,
          displayRange: '1.0—1.7分'
        },
        {
          code: 'L2',
          secondaryLabel: '较差',
          groupLabel: '较低分',
          minScore: 1.7,
          minInclusive: false,
          maxScore: 2.7,
          maxInclusive: true,
          displayRange: '1.7—2.7分'
        },
        {
          code: 'L3',
          secondaryLabel: '合格',
          groupLabel: '中分',
          minScore: 2.7,
          minInclusive: false,
          maxScore: 3.5,
          maxInclusive: true,
          displayRange: '2.7—3.5分'
        },
        {
          code: 'L4',
          secondaryLabel: '较优秀',
          groupLabel: '较高分',
          minScore: 3.5,
          minInclusive: false,
          maxScore: 4.3,
          maxInclusive: true,
          displayRange: '3.5—4.3分'
        },
        {
          code: 'L5',
          secondaryLabel: '优秀',
          groupLabel: '高分',
          minScore: 4.3,
          minInclusive: false,
          maxScore: 5,
          maxInclusive: true,
          displayRange: '4.3—5.0分'
        }
      ]
    }
  ],
  overallLevels: [
    {
      rank: 1,
      code: 'excellent',
      name: '优秀胜任',
      minScore: 45,
      minInclusive: true,
      maxScore: null,
      maxInclusive: null,
      displayRange: '45分以上'
    },
    {
      rank: 2,
      code: 'good',
      name: '良好胜任',
      minScore: 40,
      minInclusive: true,
      maxScore: 45,
      maxInclusive: false,
      displayRange: '40-45分'
    },
    {
      rank: 3,
      code: 'qualified',
      name: '合格胜任',
      minScore: 32.5,
      minInclusive: true,
      maxScore: 40,
      maxInclusive: false,
      displayRange: '32.5-40分'
    },
    {
      rank: 4,
      code: 'weak',
      name: '薄弱胜任',
      minScore: 25,
      minInclusive: true,
      maxScore: 32.5,
      maxInclusive: false,
      displayRange: '25-32.5分'
    },
    {
      rank: 5,
      code: 'not_qualified',
      name: '尚未胜任',
      minScore: 10,
      minInclusive: true,
      maxScore: 25,
      maxInclusive: false,
      displayRange: '25分以下'
    }
  ],
  validity: {
    good: '本次测评作答效度良好，结果具有较好的参考价值。测评结果仍应结合实际工作表现、行为观察、访谈及其他评价信息进行综合解读，不宜作为人才决策的唯一依据。',
    questionable: '该受测者存在掩饰真实想法的可能性，本次测评结果请谨慎解读，必要时可重新开展评估。'
  }
}
