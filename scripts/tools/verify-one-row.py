"""单独打印某姓名行的 12 维度 + 用导出题列重算的 12 维度。"""
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
    ("自我否定", "V5+1-V34+1-V43+1-V50+1-V67+1-V79+1-V24", 3.275, 1.558),
    ("怀疑感",   "V6+1-V39+V25+V55+1-V73+V83+1-V90+V65+V71+1-V22", 4.618, 1.697),
    ("职业倦怠", "1-V8+V29+V45+1-V60+1-V69+V76+V88+V62+V13", 2.937, 2.034),
]
def tk(s):
    o,i=[],0
    while i<len(s):
        c=s[i]
        if c.isspace():i+=1
        elif c in "+-*/()":o.append(c);i+=1
        elif c=="V":
            j=i+1
            while j<len(s)and s[j].isdigit():j+=1
            o.append(s[i:j]);i=j
        elif c.isdigit():
            j=i
            while j<len(s)and s[j].isdigit():j+=1
            o.append(s[i:j]);i=j
        else:raise ValueError(c)
    return o
def rpn(t):
    p={'+':1,'-':1,'*':2,'/':2};o=[];s=[]
    for x in t:
        if x[0]=='V'or x[0].isdigit():o.append(x)
        elif x in p:
            while s and s[-1]in p and p[s[-1]]>=p[x]:o.append(s.pop())
            s.append(x)
        elif x=='(':s.append(x)
        elif x==')':
            while s and s[-1]!='(':o.append(s.pop())
            s.pop()
    while s:o.append(s.pop())
    return o
def ev(r,v):
    s=[]
    for t in r:
        if t[0]=='V':s.append(v.get(t,0.0))
        elif t[0].isdigit():s.append(float(t))
        else:
            b=s.pop();a=s.pop()
            if t=='+':s.append(a+b)
            elif t=='-':s.append(a-b)
            elif t=='*':s.append(a*b)
            elif t=='/':s.append(a/b)
    return s[0]
def E(s,v):return ev(rpn(tk(s)),v)

fn,name=sys.argv[1],sys.argv[2]
z=zipfile.ZipFile(fn)
ss=re.findall(r"<t[^>]*>([^<]*)</t>",z.read("xl/sharedStrings.xml").decode("utf-8"))
ns="{http://schemas.openxmlformats.org/spreadsheetml/2006/main}"
sn=next(n for n in z.namelist()if n.startswith("xl/worksheets/"))
root=ET.fromstring(z.read(sn).decode("utf-8"))
rows=root.find(ns+"sheetData").findall(ns+"row")
def cv(c):
    v=c.find(ns+"v")
    if v is None:return""
    if c.get("t")=="s":return ss[int(v.text)]
    return v.text
def cl(n):
    s=""
    while n:
        n,r=divmod(n-1,26);s=chr(65+r)+s
    return s

for r in rows[1:]:
    cells={c.get("r"):cv(c) for c in r.findall(ns+"c")}
    rn=r.get("r")
    if cells.get(f"B{rn}")!=name:continue
    print(f"=== 找到 {name} (行 {rn}) ===")
    print(f"\n90 题答题（D..CO）:")
    vars={}
    line=[]
    for q in range(1,91):
        col=cl(3+q)
        v=cells.get(col+rn,"0")
        v=int(float(v)) if v not in("","")else 0
        vars[f"V{q}"]=v
        line.append(f"V{q}={v}")
    print(" ".join(line))
    print(f"\n12 维度对比（导出列 vs 用题列重算）:")
    print(f"{'维度':<10} {'导出':<10} {'重算':<10} {'差异':<10}")
    for i,(d,e,avg,sd) in enumerate(DIMS):
        col=cl(94+i)
        exp_v=cells.get(col+rn,"")
        exp_v=float(exp_v) if exp_v else 0
        raw=E(e,vars)
        s=(raw-avg)/sd*2+5.5
        if s<0:s=0
        if s>10:s=10
        s=round(s+1e-9,2)
        diff="✓" if abs(s-exp_v)<0.01 else f"✗ {s-exp_v:+.2f}"
        print(f"{d:<10} {exp_v:<10.2f} {s:<10.2f} {diff}")
