"""
从 260428 Excel sheet1 提取每个维度公式（CN4..CY4），解析每个题号的方向：
- 引用 Xn 形式 → 正向
- 1-Xn 形式 → 反向
"""
import zipfile, re, sys, xml.etree.ElementTree as ET

fn = sys.argv[1]
z = zipfile.ZipFile(fn)
ss = re.findall(r"<t[^>]*>([^<]*)</t>", z.read("xl/sharedStrings.xml").decode("utf-8"))
ns = "{http://schemas.openxmlformats.org/spreadsheetml/2006/main}"

# 列字母 → 数字
def colnum(s):
    n = 0
    for c in s: n = n*26 + (ord(c)-64)
    return n

# 列字母 ↔ 题号（新版 Excel 加了 A=姓名，B=V1, ..., CM=V90）
def col2qn(letters):
    return colnum(letters) - 1

DIM_COLS = [("CO","焦虑"),("CP","抑郁"),("CQ","心理失衡"),("CR","敌意"),
            ("CS","恐惧"),("CT","身体不适"),("CU","认知衰退"),("CV","情绪化"),
            ("CW","挫折感"),("CX","自我怀疑"),("CY","怀疑感"),("CZ","职业倦怠")]

root = ET.fromstring(z.read("xl/worksheets/sheet1.xml").decode("utf-8"))
rows = root.find(ns+"sheetData").findall(ns+"row")
r4 = next(r for r in rows if r.get("r") == "4")
cells = {c.get("r"): c for c in r4.findall(ns+"c")}

print(f"{'维度':<10} {'题号':<6} {'方向':<6}")
print("-" * 30)
for col, dim in DIM_COLS:
    cell = cells.get(col + "4")
    if cell is None: continue
    f = cell.find(ns+"f")
    if f is None: continue
    formula = f.text
    # 提取 raw 部分（在 -avg 之前）
    m = re.match(r"\(\((.+?)\)-[\d.]+\)", formula)
    if not m:
        print(f"{dim} 解析失败: {formula}")
        continue
    raw_expr = m.group(1)
    # 拆 + 项
    parts = re.split(r'(?<![A-Z])\+', raw_expr)  # 不在字母后的 +
    print(f"\n=== {dim} ===  公式: {raw_expr}")
    for p in parts:
        p = p.strip()
        m2 = re.match(r"^1-([A-Z]+)4$", p)
        if m2:
            qn = col2qn(m2.group(1))
            print(f"  V{qn:<4} 反向 (1-V{qn})")
            continue
        m2 = re.match(r"^([A-Z]+)4$", p)
        if m2:
            qn = col2qn(m2.group(1))
            print(f"  V{qn:<4} 正向 (V{qn})")
            continue
        print(f"  ?? unparsed: {p}")
