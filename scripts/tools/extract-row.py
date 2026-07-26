"""从客户 260428测算 Excel 提取 R4 董的 90 题答题，输出 V1=x V2=x ... 格式"""
import zipfile, re, sys, xml.etree.ElementTree as ET
fn = sys.argv[1]; rownum = sys.argv[2]
z = zipfile.ZipFile(fn)
ns = "{http://schemas.openxmlformats.org/spreadsheetml/2006/main}"
root = ET.fromstring(z.read("xl/worksheets/sheet1.xml").decode("utf-8"))
rows = root.find(ns+"sheetData").findall(ns+"row")
def cell_val(c):
    v = c.find(ns+"v")
    return v.text if v is not None else None
def col2num(s):
    n = 0
    for c in s: n = n*26 + (ord(c)-64)
    return n
target = None
for r in rows:
    if r.get("r") == rownum:
        target = r; break
cells = {c.get("r"): cell_val(c) for c in target.findall(ns+"c")}
out = []
# 90 题：B=1, C=2, ..., CM=90（新版加了 A=姓名）
for q in range(1, 91):
    n = q + 1  # 列号 = 题号 + 1
    s = ""
    while n:
        n, rem = divmod(n-1, 26)
        s = chr(65+rem) + s
    v = cells.get(s+rownum, "0")
    out.append(f"V{q}={v}")
print(" ".join(out))
# 同时打印 Excel 计算结果（CO..CZ 列）
print("--- Excel 输出 ---")
for col, dim in [("CO","焦虑"),("CP","抑郁"),("CQ","心理失衡"),("CR","敌意"),("CS","恐惧"),("CT","身体不适"),("CU","认知衰退"),("CV","情绪化"),("CW","挫折感"),("CX","自我怀疑"),("CY","怀疑感"),("CZ","职业倦怠")]:
    v = cells.get(col+rownum)
    if v: print(f"{dim:<10} {float(v):.4f}")
