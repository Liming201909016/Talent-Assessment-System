"""把 002 Excel 公式中的列字母换成 V 编号，与 Go 公式对比。
Excel 列结构：A=姓名, B-E=元数据, F=V1, G=V2, ..., EO=V140。
公式形如 (F2+S2+6-AG2+...)/10
列字母 -> 题号：F=6 → V1, 即 V_n 在第 (n+5) 列
"""
import re

# Excel 公式（来自 sheet1 R2 的 EP2..FB2）
EXCEL = {
    "社会性":   "(F+S+6-AG+6-AU+6-BI+6-BY+6-CN+6-DC+6-DP+6-EC)/10",
    "进取性":   "(G+T+AH+AV+6-BJ+BZ+CO+DD+6-DQ+ED)/10",
    "领导性":   "(H+U+AJ+AW+BK+CA+CP+DE+6-DR+6-EE)/10",
    "计划性":   "(I+V+Z+AI+AX+BL+CB+CH+CQ+DF+DS+6-EF)/12",
    "人际敏感性":"(J+W+AK+AY+BC+BN+BR+CC+CJ+CR+6-DG+6-DT+6-EG)/13",
    "自信心":   "(K+X+AL+AZ+BM+BS+CE+CS+DB+DH+6-DU+6-EH)/12",
    "责任心":   "(L+Y+AM+BA+BP+CD+6-CU+6-DI+6-DV+6-EI)/10",
    "学习力":   "(6-M+AA+AO+BB+6-BQ+CF+CT+6-DA+DJ+DW+6-EJ)/11",
    "创新性":   "(6-N+AB+AP+6-BD+BT+6-CG+CV+DK+DX+6-EK)/10",
    "情绪稳定性":"(O+AC+AQ+BE+BU+CI+CW+6-DL+6-DY+6-EL)/10",
    "自律性":   "(P+6-AD+AR+BF+6-BV+CK+6-CX+6-DM+DZ+EM)/10",
    "决断性":   "(6-Q+AE+AN+AS+BG+BO+BW+CL+CY+DN+EA+EN)/12",
    "合作性":   "(R+AF+AT+BH+BX+CM+CZ+6-DO+EB+6-EO)/10",
}

# Go 代码公式（来自 tester_score.go standScore2）
GO = {
    "社会性":   "(V1+V14+6-V28+6-V42+6-V56+6-V72+6-V87+6-V102+6-V115+6-V128)/10.0",
    "进取性":   "(V2+V15+V29+V43+6-V57+V73+V88+V103+6-V116+V129)/10.0",
    "领导性":   "(V3+V16+V31+V44+V58+V74+V89+V104+6-V117+6-V130)/10.0",
    "计划性":   "(V4+V17+V21+V30+V45+V59+V75+V81+V90+V105+V118+6-V131)/12.0",
    "人际敏感性":"(V5+V18+V32+V46+V50+V61+V65+V76+V83+V91+6-V106+6-V119+6-V132)/13.0",
    "自信心":   "(V6+V19+V33+V47+V60+V66+V78+V92+V101+V107+6-V120+6-V133)/12.0",
    "责任心":   "(V7+V20+V34+V48+V63+V77+6-V94+6-V108+6-V121+6-V134)/10.0",
    "学习力":   "(6-V8+V22+V36+V49+6-V64+V79+V93+6-V100+V109+V122+6-V135)/11.0",
    "创新性":   "(6-V9+V23+V37+6-V51+V67+6-V80+V95+V110+V123+6-V136)/10.0",
    "情绪稳定性":"(V10+V24+V38+V52+V68+V82+V96+6-V111+6-V124+6-V137)/10.0",
    "自律性":   "(V11+6-V25+V39+V53+6-V69+V84+6-V97+V112+6-V125+V138)/10.0",
    "决断性":   "(6-V12+V26+V35+V40+V54+V62+V70+V85+V98+V113+V126+V139)/12.0",
    "合作性":   "(V13+V27+V41+V55+V71+V86+V99+6-V114+V127+6-V140)/10.0",
}

def col_to_n(s):
    n = 0
    for c in s: n = n*26 + (ord(c.upper())-64)
    return n

def excel_to_v(formula):
    """把 Excel 列字母换成 V编号。F→V1, G→V2 ... 列号 - 5 = V编号"""
    def rep(m):
        col = m.group(1)
        n = col_to_n(col) - 5
        return f"V{n}"
    return re.sub(r"\b([A-Z]{1,3})\b(?!\d)", rep, formula)

def parse_terms(expr):
    """提取每个 +/- 项及其方向"""
    inner = re.sub(r"^\(|\)/[\d.]+$", "", expr.strip())
    terms = []
    for p in re.split(r'(?<![A-Z])\+', inner):
        p = p.strip()
        if p.startswith("6-"):
            terms.append(("6-" + p[2:], p[2:]))
        elif p.startswith("V"):
            terms.append((p, p))
        else:
            terms.append(("?", p))
    return terms

def divisor(expr):
    m = re.search(r"/(\d+(?:\.\d+)?)", expr)
    return float(m.group(1)) if m else 1.0

print(f"{'维度':<12} {'Go 题数':<8} {'Excel 题数':<10} {'分母 Go':<8} {'分母 Excel':<10} {'方向匹配?'}")
print("-" * 90)
diff_dims = []
for dim in EXCEL:
    e_v = excel_to_v(EXCEL[dim])
    e_terms = parse_terms(e_v)
    g_terms = parse_terms(GO[dim])
    e_div = divisor(EXCEL[dim])
    g_div = divisor(GO[dim])
    e_set = sorted(e_terms)
    g_set = sorted(g_terms)
    match = "✓" if e_set == g_set and e_div == g_div else "✗"
    if match == "✗": diff_dims.append(dim)
    print(f"{dim:<10} {len(g_terms):<8} {len(e_terms):<10} {g_div:<8g} {e_div:<10g} {match}")
    if match == "✗":
        e_only = set(e_terms) - set(g_terms)
        g_only = set(g_terms) - set(e_terms)
        if e_only: print(f"   只在 Excel: {sorted(e_only)}")
        if g_only: print(f"   只在 Go:    {sorted(g_only)}")

print()
if diff_dims:
    print(f"⚠️ 不一致维度: {diff_dims}")
else:
    print("✅ 13 个维度公式（题号集合 + 方向 + 分母）完全一致")
