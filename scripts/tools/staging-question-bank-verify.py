#!/usr/bin/env python3
import base64, hashlib, hmac, io, json, re, subprocess, time, urllib.error, urllib.request, uuid, zipfile
from pathlib import Path
APP_DIR=Path('/opt/talent-assessment'); BASE='http://127.0.0.1:8092'

def b64url(v): return base64.urlsafe_b64encode(v).rstrip(b'=').decode()
def config_text():
    env=subprocess.run(['systemctl','show','talent-assessment','--property=Environment','--value'],check=True,capture_output=True,text=True).stdout
    m=re.search(r'(?:^|\s)APP_ENV=([^\s]+)',env); name=m.group(1) if m else 'local'
    files=[APP_DIR/'configs'/'application.yml',APP_DIR/'configs'/f'application-{name}.yml']
    return '\n'.join(p.read_text(encoding='utf-8') for p in files if p.exists())
def scalar(text,section,key,default=''):
    value=default
    for sm in re.finditer(rf'(?ms)^{re.escape(section)}:\s*\n(.*?)(?=^[A-Za-z][\w-]*:\s*$|\Z)',text):
        km=re.search(rf'(?m)^\s+{re.escape(key)}:\s*(.+?)\s*$',sm.group(1))
        if km:value=km.group(1).strip().strip('"\'')
    return value
def sign(secret,token_id):
    header=b64url(json.dumps({'alg':'HS512','typ':'JWT'},separators=(',',':')).encode()); payload=b64url(json.dumps({'login_user_key':token_id},separators=(',',':')).encode()); raw=f'{header}.{payload}'
    return f'{raw}.{b64url(hmac.new(secret.encode(),raw.encode(),hashlib.sha512).digest())}'
def api(path,body,token,expect_error=None):
    req=urllib.request.Request(BASE+path,data=json.dumps(body,ensure_ascii=False).encode(),headers={'Content-Type':'application/json','Authorization':'Bearer '+token},method='POST')
    with urllib.request.urlopen(req,timeout=60) as response: payload=json.loads(response.read().decode())
    if expect_error:
        if payload.get('code') in (0,200) or expect_error not in payload.get('msg',''): raise RuntimeError(f'{path} expected {expect_error}: {payload}')
        return payload
    if payload.get('code') not in (0,200): raise RuntimeError(f'{path}: {payload}')
    return payload.get('data')
def multipart_upload(path,file_name,data,token):
    boundary='----verify'+uuid.uuid4().hex
    body=(f'--{boundary}\r\nContent-Disposition: form-data; name="file"; filename="{file_name}"\r\nContent-Type: application/vnd.openxmlformats-officedocument.spreadsheetml.sheet\r\n\r\n').encode()+data+f'\r\n--{boundary}--\r\n'.encode()
    req=urllib.request.Request(BASE+path,data=body,headers={'Content-Type':'multipart/form-data; boundary='+boundary,'Authorization':'Bearer '+token},method='POST')
    with urllib.request.urlopen(req,timeout=60) as response: payload=json.loads(response.read().decode())
    if payload.get('code') not in (0,200): raise RuntimeError(f'{path}: {payload}')
    return payload.get('data')
def mysql(sql): return subprocess.run(['sudo','-n','mysql','element','-Nse',sql],check=True,capture_output=True,text=True).stdout.strip()
def xlsx(sheet, rows):
    escaped=lambda s:str(s).replace('&','&amp;').replace('<','&lt;').replace('>','&gt;')
    cols='ABCDEFGHIJKLMNOPQRSTUVWXYZ'
    xml=['<?xml version="1.0" encoding="UTF-8" standalone="yes"?><worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData>']
    for ri,row in enumerate(rows,1):
        xml.append(f'<row r="{ri}">')
        for ci,value in enumerate(row): xml.append(f'<c r="{cols[ci]}{ri}" t="inlineStr"><is><t>{escaped(value)}</t></is></c>')
        xml.append('</row>')
    xml.append('</sheetData></worksheet>')
    out=io.BytesIO()
    with zipfile.ZipFile(out,'w',zipfile.ZIP_DEFLATED) as z:
        z.writestr('[Content_Types].xml','<?xml version="1.0"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/><Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/></Types>')
        z.writestr('_rels/.rels','<?xml version="1.0"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/></Relationships>')
        z.writestr('xl/workbook.xml',f'<?xml version="1.0"?><workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets><sheet name="{escaped(sheet)}" sheetId="1" r:id="rId1"/></sheets></workbook>')
        z.writestr('xl/_rels/workbook.xml.rels','<?xml version="1.0"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/></Relationships>')
        z.writestr('xl/worksheets/sheet1.xml',''.join(xml))
    return out.getvalue()

def main():
    suffix=uuid.uuid4().hex[:10]; title='QB-VERIFY-'+suffix; text=config_text(); secret=scalar(text,'jwt','secret'); db=scalar(text,'redis','db','1')
    token_id='qb-verify-'+suffix; key='login_tokens:'+token_id; now=int(time.time()*1000)
    login={'userId':1,'token':token_id,'loginTime':now,'expireTime':now+1800000,'permissions':['*:*:*'],'roles':['admin'],'user':{'userId':1,'userName':'admin','avatar':''}}
    subprocess.run(['redis-cli','-n',db,'SET',key,json.dumps(login,separators=(',',':')),'EX','1800'],check=True,stdout=subprocess.DEVNULL); token=sign(secret,token_id)
    repo_id=question_id=None
    try:
        before_legacy=int(mysql('SELECT COUNT(*) FROM el_qu WHERE dimension_id IS NULL;')); competency=int(mysql('SELECT COUNT(*) FROM el_qu WHERE dimension_id IS NOT NULL;'))
        page=api('/exam/api/qu/qu/paging',{'current':1,'size':20,'params':{}},token)
        if page['total']!=before_legacy or any(row.get('dimensionId') for row in page['records']): raise RuntimeError('legacy isolation failed')
        repo=api('/exam/api/repo/save',{'title':title,'code':'QB'+suffix[:4],'radioCount':999},token); repo_id=repo['id']
        if repo['radioCount']!=0: raise RuntimeError('protected repo counts overwritten')
        question=api('/exam/api/qu/qu/save',{'content':'V990001','title':'verify','quType':1,'level':1,'repoIds':[repo_id], 'answerList':[{'content':'A','isRight':True},{'content':'B','isRight':False}]},token); question_id=question['id']
        api('/exam/api/repo/batch-action',{'quIds':[question_id,question_id,''],'repoIds':[repo_id,repo_id],'remove':False},token)
        linked=mysql(f"SELECT CONCAT(COUNT(*),'|',MIN(sort),'|',MAX(sort)) FROM el_qu_repo WHERE qu_id='{question_id}' AND repo_id='{repo_id}';")
        if linked!='1|1|1': raise RuntimeError('batch add normalization failed '+linked)
        api('/exam/api/repo/batch-action',{'quIds':[question_id],'repoIds':['missing-repo'],'remove':False},token,expect_error='部分题库不存在')
        if mysql(f"SELECT COUNT(*) FROM el_qu_repo WHERE qu_id='{question_id}' AND repo_id='{repo_id}';")!='1': raise RuntimeError('failed batch changed existing association')
        api('/exam/api/repo/batch-action',{'quIds':[question_id],'repoIds':[repo_id],'remove':True},token)
        if mysql(f"SELECT COUNT(*) FROM el_qu_repo WHERE qu_id='{question_id}' AND repo_id='{repo_id}';")!='0': raise RuntimeError('batch remove failed')
        api('/exam/api/repo/batch-action',{'quIds':[question_id],'repoIds':[repo_id],'remove':False},token)
        rows=[['题目序号','题目类型','题目内容','整体解析','题目图片','题目视频','所属题库','是否正确项','选项内容','选项解析','选项图片','题目标题'],['1','1','V990002','','','',repo_id,'1','是','','','导入验证'],['1','','','','','','','0','否','','','']]
        imported=multipart_upload('/exam/api/qu/qu/import-excel','verify.xlsx',xlsx(title,rows),token)
        if '导入 1 题' not in imported['message']: raise RuntimeError('import failed '+str(imported))
        imported_id=mysql(f"SELECT id FROM el_qu WHERE content='V990002' AND dimension_id IS NULL ORDER BY create_time DESC LIMIT 1;")
        # invalid extra repo must roll back sheet repository and question
        bad_title='QB-BAD-'+suffix
        bad_rows=[rows[0],['1','1','V990003','','','','missing-repo','1','是','','','坏导入']]
        try: multipart_upload('/exam/api/qu/qu/import-excel','bad.xlsx',xlsx(bad_title,bad_rows),token); raise RuntimeError('bad import unexpectedly succeeded')
        except RuntimeError as exc:
            if '不存在的题库' not in str(exc): raise
        if mysql(f"SELECT (SELECT COUNT(*) FROM el_repo WHERE title='{bad_title}')+(SELECT COUNT(*) FROM el_qu WHERE content='V990003');")!='0': raise RuntimeError('bad import rollback failed')
        api('/exam/api/qu/qu/delete',{'ids':[question_id,imported_id]},token)
        question_id=None
        api('/exam/api/repo/delete',{'ids':[repo_id]},token); repo_id=None
        after_legacy=int(mysql('SELECT COUNT(*) FROM el_qu WHERE dimension_id IS NULL;'))
        if after_legacy!=before_legacy: raise RuntimeError(f'legacy count changed {before_legacy}->{after_legacy}')
        print('QUESTION_BANK_VERIFY_OK'); print(f'legacy_total={before_legacy}'); print(f'competency_total={competency}'); print('page_competency_rows=0'); print('batch_add_remove_and_rollback=ok'); print('import_success_and_rollback=ok'); print('protected_repo_fields=ok')
    finally:
        if question_id: mysql(f"DELETE FROM el_qu_answer WHERE qu_id='{question_id}'; DELETE FROM el_qu_repo WHERE qu_id='{question_id}'; DELETE FROM el_qu WHERE id='{question_id}';")
        mysql(f"DELETE qa FROM el_qu_answer qa JOIN el_qu q ON q.id=qa.qu_id WHERE q.content IN ('V990002','V990003'); DELETE qr FROM el_qu_repo qr JOIN el_qu q ON q.id=qr.qu_id WHERE q.content IN ('V990002','V990003'); DELETE FROM el_qu WHERE content IN ('V990002','V990003'); DELETE FROM el_repo WHERE title IN ('{title}','QB-BAD-{suffix}');")
        subprocess.run(['redis-cli','-n',db,'DEL',key],check=False,stdout=subprocess.DEVNULL)
        remaining=mysql(f"SELECT (SELECT COUNT(*) FROM el_repo WHERE title IN ('{title}','QB-BAD-{suffix}'))+(SELECT COUNT(*) FROM el_qu WHERE content IN ('V990001','V990002','V990003')); ")
        print('cleanup_remaining='+remaining)
if __name__=='__main__': main()
