"""
扫描 12 维度公式中每道题：
- 公式正向(+Vn) → 期望 is_right=1 的选项内容应表达该维度特征
- 公式反向(+(1-Vn)) → 期望 is_right=1 的选项内容应表达该维度的反面（即"否认特征"）

输入：
  arg1 = paper_id (用其 qu_id 来取选项内容；保证与高分这个考生用的题库版本一致)

输出：每个维度，每道题的：
  V编号 | 公式方向 | is_right=1选项 | is_right=0选项
  人工对照即可看出语义是否反了
"""
import sys, json
import xml.etree.ElementTree as ET

# 12 维公式
DIMS = [
    ("焦虑",     "1-V1+1-V36+V41+V56+1-V61+V81+1-V15"),
    ("抑郁",     "V12+V26+V38+1-V49+V70+1-V77+1-V86+V21"),
    ("心理失衡", "1-V11+1-V31+1-V47+V57+V75+V84+1-V19"),
    ("敌意",     "V3+V35+V46+V54+V63+1-V78+V14"),
    ("恐惧",     "1-V4+1-V28+1-V40+V51+1-V72+V87+1-V23"),
    ("身体不适", "1-V2+V30+1-V44+V59+V74+1-V89+V17"),
    ("认知衰退", "V7+1-V33+V42+V52+V68+1-V80+V20"),
    ("情绪化",   "V9+1-V37+V32+1-V53+1-V66+V85+1-V16"),
    ("挫折感",   "1-V10+1-V27+V48+1-V58+1-V64+1-V82+V18"),
    ("自我怀疑", "V5+1-V34+1-V43+1-V50+1-V67+1-V79+1-V24"),
    ("怀疑感",   "V6+1-V39+V25+V55+1-V73+V83+1-V90+V65+V71+1-V22"),
    ("职业倦怠", "1-V8+V29+V45+1-V60+1-V69+V76+V88+V62+V13"),
]

def parse_terms(expr):
    """ 返回 [(Vn, '+' or '-'), ...]
    +Vn → '+', +(1-Vn) → '-'
    """
    out = []
    parts = expr.split('+')
    for p in parts:
        p = p.strip()
        if p.startswith('1-V'):
            out.append((p[2:], '-'))  # Vn, 反向
        elif p.startswith('V'):
            out.append((p, '+'))  # 正向
    return out

# 输入 JSON: { "Vn": [{"text": "选项文本", "is_right": 1}, ...], ... }
data = json.load(open(sys.argv[1], encoding='utf-8'))

print(f"{'维度':<10} {'V':<5} {'方向':<8} {'is_right=1 选项':<40} {'is_right=0 选项':<40}")
print("-" * 130)
for dim, expr in DIMS:
    terms = parse_terms(expr)
    for v, sign in terms:
        opts = data.get(v, [])
        r1 = next((o['text'] for o in opts if int(o.get('is_right',0))==1), '?')
        r0 = next((o['text'] for o in opts if int(o.get('is_right',0))==0), '?')
        dir_str = "正向 +Vn" if sign=='+' else "反向 1-Vn"
        # is_right=1 的选项**应当表达**：
        #   正向 → 该维度特征 (例如：焦虑→紧张/担忧)
        #   反向 → 该维度反面 (例如：焦虑→平静/放松)
        print(f"{dim:<10} {v:<5} {dir_str:<8} {r1:<40} {r0:<40}")
    print()
