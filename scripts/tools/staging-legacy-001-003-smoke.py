#!/usr/bin/env python3
import base64,hashlib,hmac,json,re,subprocess,time,urllib.error,urllib.request,uuid
from pathlib import Path
APP=Path('/opt/talent-assessment');BASE='http://127.0.0.1:8092'
def b64(v):return base64.urlsafe_b64encode(v).rstrip(b'=').decode()
def config():
 e=subprocess.run(['systemctl','show','talent-assessment','--property=Environment','--value'],capture_output=True,text=True,check=True).stdout;m=re.search(r'APP_ENV=([^\s]+)',e);n=m.group(1) if m else 'local';return '\n'.join(p.read_text() for p in [APP/'configs'/'application.yml',APP/'configs'/f'application-{n}.yml'] if p.exists())
def scalar(t,s,k,d=''):
 v=d
 for x in re.finditer(rf'(?ms)^{s}:\s*\n(.*?)(?=^[A-Za-z][\w-]*:\s*$|\Z)',t):
  m=re.search(rf'(?m)^\s+{k}:\s*(.+?)\s*$',x.group(1));v=m.group(1).strip().strip('"\'') if m else v
 return v
def token(secret,tid):
 h=b64(json.dumps({'alg':'HS512','typ':'JWT'},separators=(',',':')).encode());p=b64(json.dumps({'login_user_key':tid},separators=(',',':')).encode());r=f'{h}.{p}';return f'{r}.{b64(hmac.new(secret.encode(),r.encode(),hashlib.sha512).digest())}'
def api(path,body,tok):
 req=urllib.request.Request(BASE+path,data=json.dumps(body).encode(),headers={'Content-Type':'application/json','Authorization':'Bearer '+tok},method='POST')
 with urllib.request.urlopen(req,timeout=60) as r:p=json.loads(r.read())
 if p.get('code') not in (0,200):raise RuntimeError(f'{path}: {p}')
 return p.get('data')
def mysql(sql):return subprocess.run(['sudo','-n','mysql','element','-Nse',sql],capture_output=True,text=True,check=True).stdout.strip()
def main():
 suf=uuid.uuid4().hex[:8];txt=config();db=scalar(txt,'redis','db','1');tid='legacy-smoke-'+suf;key='login_tokens:'+tid;ms=int(time.time()*1000);lu={'userId':1,'token':tid,'loginTime':ms,'expireTime':ms+1800000,'permissions':['*:*:*'],'roles':['admin'],'user':None};subprocess.run(['redis-cli','-n',db,'SET',key,json.dumps(lu,separators=(',',':')),'EX','1800'],check=True,stdout=subprocess.DEVNULL);tok=token(scalar(txt,'jwt','secret'),tid);created=[]
 try:
  repos=mysql("SELECT id,code FROM el_repo WHERE code IN ('00101','00201','00301') ORDER BY code").splitlines()
  if len(repos)!=3:raise RuntimeError('missing legacy repos: '+str(repos))
  for line in repos:
   rid,code=line.split('\t');title=f'LEGACY-SMOKE-{code}-{suf}'
   exam=api('/exam/api/exam/exam/save',{'title':title,'content':'temporary legacy smoke','assessmentType':'legacy','scoringMode':'legacy','joinType':1,'openType':1,'isOpen':1,'answerType':1,'state':0,'totalTime':30,'repoList':[{'repoId':rid,'radioCount':2,'radioScore':5,'multiCount':0,'multiScore':0,'judgeCount':0,'judgeScore':0}],'departIds':[]},tok);eid=exam['id']
   paper=api('/exam/api/paper/paper/create-paper',{'examId':eid},tok);pid=paper['id'];created.append((eid,pid,code))
   detail=api('/exam/api/paper/paper/paper-detail',{'id':pid},tok);qs=detail.get('radioList',[])
   if len(qs)!=2:raise RuntimeError(f'{code} paper count {len(qs)}')
   for q in qs:
    qd=api('/exam/api/paper/paper/qu-detail',{'paperId':pid,'quId':q['quId']},tok);answers=qd.get('answerList') or []
    if not answers:raise RuntimeError(code+' answer missing')
    aid=answers[0]['answerId'];api('/exam/api/paper/paper/fill-answer',{'paperId':pid,'quId':q['quId'],'answers':[aid],'answer':aid},tok)
   api('/exam/api/paper/paper/hand-exam',{'id':pid},tok);result=api('/exam/api/paper/paper/paper-result',{'id':pid},tok);score=api('/exam/api/paper/paper/stand-score',{'id':pid},tok)
   if result is None or score is None:raise RuntimeError(code+' result/score missing')
   print(f'{code}=paper2|submit|result|stand-score')
  print('LEGACY_001_003_SMOKE_OK')
 finally:
  for eid,pid,code in reversed(created):
   try:api('/exam/api/paper/paper/delete',{'ids':[pid]},tok)
   except Exception as e:print('paper_cleanup_'+code+'='+str(e))
   try:api('/exam/api/exam/exam/delete',{'ids':[eid]},tok)
   except Exception as e:print('exam_cleanup_'+code+'='+str(e))
  subprocess.run(['redis-cli','-n',db,'DEL',key],stdout=subprocess.DEVNULL)
  if created:
   ids=','.join("'%s'"%x[0] for x in created);rem=mysql(f"SELECT (SELECT COUNT(*) FROM el_exam WHERE id IN ({ids}))+(SELECT COUNT(*) FROM el_paper WHERE exam_id IN ({ids}))")
   print('cleanup_remaining='+rem)
   if rem!='0':raise RuntimeError('cleanup='+rem)
if __name__=='__main__':main()
