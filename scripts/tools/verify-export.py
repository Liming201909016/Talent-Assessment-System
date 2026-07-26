"""
对照导出文件：用每行 90 题数据代入代码公式，与导出的 12 维列比较。
导出文件结构（修复版后）：A=ID, B=姓名, C=手机号, D..CO=1..90 题得分, CP..DA=12 维度
"""
import zipfile, re, sys, xml.etree.ElementTree as ET

DIMS = [
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

def tokenize(s):
    out, i = [], 0
    while i < len(s):
        c = s[i]
        if c.isspace(): i+=1
        elif c in "+-*/()": out.append(c); i+=1
        elif c == "V":
            j = i+1
            while j<len(s) and s[j].isdigit(): j+=1
            out.append(s[i:j]); i=j
        elif c.isdigit():
            j = i
            while j<len(s) and s[j].isdigit(): j+=1
            out.append(s[i:j]); i=j
        else: raise ValueError(c)
    return out

def to_rpn(tokens):
    prec={'+':1,'-':1,'*':2,'/':2}
    out, st = [], []
    for t in tokens:
        if t[0]=='V' or t[0].isdigit(): out.append(t)
        elif t in prec:
            while st and st[-1] in prec and prec[st[-1]] >= prec[t]: out.append(st.pop())
            st.append(t)
        elif t=='(': st.append(t)
        elif t==')':
            while st and st[-1]!='(': out.append(st.pop())
            st.pop()
    while st: out.append(st.pop())
    return out

def evalrpn(rpn, vars):
    s=[]
    for t in rpn:
        if t[0]=='V': s.append(vars.get(t,0.0))
        elif t[0].isdigit(): s.append(float(t))
        else:
            b=s.pop(); a=s.pop()
            if t=='+': s.append(a+b)
            elif t=='-': s.append(a-b)
            elif t=='*': s.append(a*b)
            elif t=='/': s.append(a/b)
    return s[0]

def expr(s, vars):
    return evalrpn(to_rpn(tokenize(s)), vars)

def score(expr_str, avg, sd, vars):
    raw = expr(expr_str, vars)
    s = (raw-avg)/sd*2 + 5.5
    if s<0: s=0
    if s>10: s=10
    return raw, round(s + 1e-9, 2)

fn = sys.argv[1]
z = zipfile.ZipFile(fn)
ss = re.findall(r"<t[^>]*>([^<]*)</t>", z.read("xl/sharedStrings.xml").decode("utf-8"))
ns = "{http://schemas.openxmlformats.org/spreadsheetml/2006/main}"
sn = next(n for n in z.namelist() if n.startswith("xl/worksheets/"))
root = ET.fromstring(z.read(sn).decode("utf-8"))
rows = root.find(ns+"sheetData").findall(ns+"row")
def cv(c):
    v=c.find(ns+"v")
    if v is None: return ""
    if c.get("t")=="s": return ss[int(v.text)]
    return v.text

def col_letter(n):
    s=""
    while n:
        n, rem = divmod(n-1, 26)
        s = chr(65+rem) + s
    return s

# 行 1 = 表头
header_row = rows[0]
header = {c.get("r"): cv(c) for c in header_row.findall(ns+"c")}

print(f"{'姓名':<14} {'维度':<10} {'代码raw':<8} {'代码分':<8} {'导出列分':<10} {'差异':<6}")
print("-" * 70)

mismatch = 0
total = 0
for r in rows[1:]:
    rownum = r.get("r")
    cells = {c.get("r"): cv(c) for c in r.findall(ns+"c")}
    name = cells.get(f"B{rownum}", "?")
    # 90 题在 D..CO（列 4..93）
    vars = {}
    for q in range(1, 91):
        col = col_letter(3 + q)  # D=4
        v = cells.get(col + rownum, "0")
        try:
            vars[f"V{q}"] = float(v) if v not in ("", None) else 0.0
        except:
            vars[f"V{q}"] = 0.0
    # 12 维在 CP..DA（列 94..105）
    for i, (dim, e, avg, sd) in enumerate(DIMS):
        col = col_letter(94 + i)
        exp_v = cells.get(col + rownum, "")
        exp_v = float(exp_v) if exp_v not in ("", None) else None
        raw, my = score(e, avg, sd, vars)
        diff = "" if exp_v is not None and abs(my-exp_v)<0.01 else f"❌ Δ={'?' if exp_v is None else f'{my-exp_v:+.2f}'}"
        if diff:
            mismatch += 1
            print(f"{name:<12} {dim:<10} {raw:<8g} {my:<8.2f} {exp_v if exp_v is not None else '?':<10} {diff}")
        total += 1
print(f"\n核对完成：{total} 个维度值，{mismatch} 处不一致")
