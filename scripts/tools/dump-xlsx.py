import zipfile, re, sys
path = sys.argv[1] if len(sys.argv) > 1 else "/tmp/closed-mbti.xlsx"
z = zipfile.ZipFile(path)
ss_xml = z.read("xl/sharedStrings.xml").decode() if "xl/sharedStrings.xml" in z.namelist() else ""
ss = re.findall(r"<t[^>]*>([^<]*)</t>", ss_xml)
print(f"sharedStrings({len(ss)}):", ss[:30])
sh = z.read("xl/worksheets/sheet1.xml").decode()
for r in range(1, 16):
    m = re.search(rf'<row r="{r}"[^>]*>(.*?)</row>', sh)
    if not m: continue
    cells = re.findall(r'<c r="([A-Z]+\d+)"(?: s="\d+")?(?: t="(\w+)")?(?:[^/>]*?/>|>(.*?)</c>)', m.group(1))
    vals=[]
    for ref, t, body in cells:
        v = re.search(r"<v>([^<]*)</v>", body or "")
        val = v.group(1) if v else ""
        if t == "s" and val.isdigit():
            val = ss[int(val)] if int(val) < len(ss) else val
        vals.append((re.sub(r"\d","",ref), val))
    print(f"row{r}:", vals[:14])
