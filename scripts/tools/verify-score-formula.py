"""
直接从客户 DB 取一份 001 (李沫) 和一份 002 (李林) 的实际答题，
按代码公式计算 12/13 维度，与已有的 Excel 模板公式对照。
"""
import sys

# Java/Go 实现的 12 维心理（001）公式
PSY_EXPR = [
    ("焦虑",     "1-V1+1-V36+V41+V56+1-V61+V81+1-V15", 2.295, 1.367),
    ("抑郁",     "V12+V26+V38+1-V49+V70+1-V77+1-V86+V21", 4.513, 1.479),
    ("心理失衡", "1-V11+1-V31+1-V47+V57+V75+V84+1-V19", 3.266, 1.448),
    ("敌意",     "V3+V35+V46+V54+V63+1-V78+V14", 3.274, 1.287),
    ("恐惧",     "1-V4+1-V28+1-V40+V51+1-V72+V87+1-V23", 3.417, 1.549),
    ("身体不适", "1-V2+V30+1-V44+V59+V74+1-V89+V17", 1.933, 1.746),
    ("认知衰退", "V7+1-V33+V42+V52+V68+1-V80+V20", 4.268, 1.395),
    ("情绪化",   "V9+1-V37+V32+1-V53+1-V66+V85+1-V16", 3.352, 1.523),
    ("挫折感",   "1-V10+1-V27+V48+1-V58+1-V64+1-V82+V18", 3.504, 1.310),
    ("自我怀疑", "V5+1-V34+1-V43+1-V50+1-V67+1-V79+1-V24", 3.275, 1.558),
    ("怀疑感",   "V6+1-V39+V25+V55+1-V73+V83+1-V90+V65+V71+1-V22", 4.618, 1.697),
    ("职业倦怠", "1-V8+V29+V45+1-V60+1-V69+V76+V88+V62+V13", 2.937, 2.034),
]

# 13 维管理（002）公式（标准化分=raw无需 std formula，直接 4 位小数）
MNG_EXPR = [
    ("社会性",   "(V1+V14+6-V28+6-V42+6-V56+6-V72+6-V87+6-V102+6-V115+6-V128)/10.0"),
    ("进取性",   "(V2+V15+V29+V43+6-V57+V73+V88+V103+6-V116+V129)/10.0"),
    ("领导性",   "(V3+V16+V31+V44+V58+V74+V89+V104+6-V117+6-V130)/10.0"),
    ("计划性",   "(V4+V17+V21+V30+V45+V59+V75+V81+V90+V105+V118+6-V131)/12.0"),
    ("人际敏感性","(V5+V18+V32+V46+V50+V61+V65+V76+V83+V91+6-V106+6-V119+6-V132)/13.0"),
    ("自信心",   "(V6+V19+V33+V47+V60+V66+V78+V92+V101+V107+6-V120+6-V133)/12.0"),
    ("责任心",   "(V7+V20+V34+V48+V63+V77+6-V94+6-V108+6-V121+6-V134)/10.0"),
    ("学习力",   "(6-V8+V22+V36+V49+6-V64+V79+V93+6-V100+V109+V122+6-V135)/11.0"),
    ("创新性",   "(6-V9+V23+V37+6-V51+V67+6-V80+V95+V110+V123+6-V136)/10.0"),
    ("情绪稳定性","(V10+V24+V38+V52+V68+V82+V96+6-V111+6-V124+6-V137)/10.0"),
    ("自律性",   "(V11+6-V25+V39+V53+6-V69+V84+6-V97+V112+6-V125+V138)/10.0"),
    ("决断性",   "(6-V12+V26+V35+V40+V54+V62+V70+V85+V98+V113+V126+V139)/12.0"),
    ("合作性",   "(V13+V27+V41+V55+V71+V86+V99+6-V114+V127+6-V140)/10.0"),
]

def tokenize(s):
    out, i = [], 0
    while i < len(s):
        c = s[i]
        if c.isspace(): i += 1
        elif c in "+-*/()": out.append(c); i += 1
        elif c == "V":
            j = i + 1
            while j < len(s) and s[j].isdigit(): j += 1
            out.append(s[i:j]); i = j
        elif c.isdigit():
            j = i
            while j < len(s) and (s[j].isdigit() or s[j] == "."): j += 1
            out.append(s[i:j]); i = j
        else: raise ValueError(c)
    return out

def to_rpn(tokens):
    prec = {'+':1,'-':1,'*':2,'/':2}
    out, st = [], []
    for t in tokens:
        if t[0]=='V' or t[0].isdigit():
            out.append(t)
        elif t in prec:
            while st and st[-1] in prec and prec[st[-1]] >= prec[t]: out.append(st.pop())
            st.append(t)
        elif t == '(': st.append(t)
        elif t == ')':
            while st and st[-1] != '(': out.append(st.pop())
            st.pop()
    while st: out.append(st.pop())
    return out

def evalrpn(rpn, vars):
    s = []
    for t in rpn:
        if t[0]=='V': s.append(vars.get(t, 0.0))
        elif t[0].isdigit(): s.append(float(t))
        else:
            b = s.pop(); a = s.pop()
            if t=='+': s.append(a+b)
            elif t=='-': s.append(a-b)
            elif t=='*': s.append(a*b)
            elif t=='/': s.append(a/b)
    return s[0]

def expr(s, vars):
    return evalrpn(to_rpn(tokenize(s)), vars)

# 模式：argv[1] = 'psy'/'mng'，argv[2..] = 答题"V1=1 V2=0 ..."
mode = sys.argv[1]
vars = {}
for arg in sys.argv[2:]:
    k, v = arg.split('=')
    vars[k] = float(v)

if mode == 'psy':
    print(f"{'维度':<10} {'raw':>5} {'标准分':>8}")
    for name, e, avg, sd in PSY_EXPR:
        raw = expr(e, vars)
        s = (raw - avg) / sd * 2 + 5.5
        if s < 0: s = 0
        if s > 10: s = 10
        s = round(s + 1e-9, 2)
        print(f"{name:<10} {raw:>5g} {s:>8.2f}")
elif mode == 'mng':
    print(f"{'维度':<12} {'分':>8}")
    for name, e in MNG_EXPR:
        v = expr(e, vars)
        v = round(v + 1e-9, 4)
        print(f"{name:<12} {v:>8.4f}")
