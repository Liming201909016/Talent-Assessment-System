import zipfile, re, sys
path = sys.argv[1]
z = zipfile.ZipFile(path)
c = z.read('word/document.xml').decode()
for label in ['姓名：','年龄：','性别：','单位：','联系方式：','报告日期：','测评时长：']:
    idx = c.find(label + '</w:t>')
    if idx < 0:
        # cross-run: find first char of label after a <w:t
        runes = list(label)
        firstChar = runes[0]
        pos = 0
        while True:
            p = c.find(firstChar, pos)
            if p < 0:
                print(f'[{label}] NOT FOUND')
                idx = -2
                break
            lb = max(0, p - 50)
            if '<w:t' not in c[lb:p]:
                pos = p + 1
                continue
            # try matching label across tags
            stripped = re.sub(r'<[^>]+>', '', c[p:p+1000])
            if stripped.startswith(label):
                idx = p
                break
            pos = p + 1
        if idx < 0:
            continue
    seg = c[idx:idx+700]
    print(f'--- {label} (idx={idx}) ---')
    # Pretty print
    print(seg[:700].replace('><', '>\n<'))
    print()
