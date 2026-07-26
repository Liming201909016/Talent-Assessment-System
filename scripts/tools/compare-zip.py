import zipfile, sys

t = zipfile.ZipFile(sys.argv[1])
g = zipfile.ZipFile(sys.argv[2])
t_names = set(e.filename for e in t.infolist())
g_names = set(e.filename for e in g.infolist())
print('Only in template:', t_names - g_names)
print('Only in generated:', g_names - t_names)
print(f'Template entries: {len(t_names)}, Generated entries: {len(g_names)}')
for n in sorted(g_names):
    te = t.getinfo(n) if n in t_names else None
    ge = g.getinfo(n)
    if te:
        if te.compress_type != ge.compress_type:
            print(f'  METHOD_DIFF {n}: template={te.compress_type} gen={ge.compress_type}')
