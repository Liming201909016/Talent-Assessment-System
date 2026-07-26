#!/usr/bin/env python3
import base64, hashlib, hmac, json, re, subprocess, sys, time, urllib.error, urllib.request, uuid
from pathlib import Path

APP_DIR=Path('/opt/talent-assessment'); BASE='http://127.0.0.1:8092'

def b64url(v): return base64.urlsafe_b64encode(v).rstrip(b'=').decode()
def config_text():
    env=subprocess.run(['systemctl','show','talent-assessment','--property=Environment','--value'],check=True,capture_output=True,text=True).stdout
    match=re.search(r'(?:^|\s)APP_ENV=([^\s]+)',env); name=match.group(1) if match else 'local'
    files=[APP_DIR/'configs'/'application.yml',APP_DIR/'configs'/f'application-{name}.yml']
    return '\n'.join(path.read_text(encoding='utf-8') for path in files if path.exists())
def scalar(text,section,key,default=''):
    value=default
    for sm in re.finditer(rf'(?ms)^{re.escape(section)}:\s*\n(.*?)(?=^[A-Za-z][\w-]*:\s*$|\Z)',text):
        km=re.search(rf'(?m)^\s+{re.escape(key)}:\s*(.+?)\s*$',sm.group(1))
        if km:value=km.group(1).strip().strip('"\'')
    return value
def sign(secret,token_id):
    header=b64url(json.dumps({'alg':'HS512','typ':'JWT'},separators=(',',':')).encode()); payload=b64url(json.dumps({'login_user_key':token_id},separators=(',',':')).encode()); raw=f'{header}.{payload}'
    return f'{raw}.{b64url(hmac.new(secret.encode(),raw.encode(),hashlib.sha512).digest())}'
def request(path,body=None,token=None,headers=None,method='POST',expect_error=None):
    hs={'Content-Type':'application/json'}
    if token: hs['Authorization']='Bearer '+token
    if headers: hs.update(headers)
    data=None if method=='GET' else json.dumps(body or {},ensure_ascii=False).encode()
    req=urllib.request.Request(BASE+path,data=data,headers=hs,method=method)
    try:
        with urllib.request.urlopen(req,timeout=40) as response: payload=json.loads(response.read().decode())
    except urllib.error.HTTPError as error: raise RuntimeError(f'{path} HTTP {error.code}: {error.read().decode(errors="replace")}')
    if expect_error:
        if payload.get('code') in (0,200) or expect_error not in payload.get('msg',''): raise RuntimeError(f'{path} expected {expect_error}: {payload}')
        return payload
    if payload.get('code') not in (0,200): raise RuntimeError(f'{path}: {payload.get("msg")}')
    return payload.get('data')
def mysql(sql): return subprocess.run(['sudo','-n','mysql','element','-Nse',sql],check=True,capture_output=True,text=True).stdout.strip()
def mysql_exec(sql): subprocess.run(['sudo','-n','mysql','element','-e',sql],check=True)
def sql_quote(value): return value.replace("'","''")

def main():
    suffix=uuid.uuid4().hex[:12]; prefix='BOUNDARY-'+suffix
    text=config_text(); secret=scalar(text,'jwt','secret'); redis_db=scalar(text,'redis','db','1')
    token_id='competency-boundary-'+suffix; now_ms=int(time.time()*1000); redis_key='login_tokens:'+token_id
    login={'userId':1,'token':token_id,'loginTime':now_ms,'expireTime':now_ms+1800000,'permissions':['*:*:*'],'roles':['admin'],'user':None}
    subprocess.run(['redis-cli','-n',redis_db,'SET',redis_key,json.dumps(login,separators=(',',':')),'EX','1800'],check=True,stdout=subprocess.DEVNULL)
    admin=sign(secret,token_id); exam_id=None; candidate_id=None; paper_id=None
    try:
        dimensions=request('/exam/api/competency/dimensions/list',{},admin)
        selected=[d['id'] for d in dimensions if d['code']=='D01']
        if len(selected)!=1 or next(d for d in dimensions if d['code']=='D01')['questionCount']!=8: raise RuntimeError('D01 readiness mismatch')
        base={'content':'temporary','assessmentType':'competency','scoringMode':'competency_average','competencyReportAudience':'frontline_employee','dimensionIds':selected,'joinType':1,'openType':1,'isOpen':1,'answerType':1,'state':0,'repoList':[],'departIds':[]}

        zero={**base,'title':prefix+'-ZERO','totalTime':0}
        request('/exam/api/exam/exam/save',zero,admin,expect_error='大于0')
        if mysql(f"SELECT COUNT(*) FROM el_exam WHERE title='{sql_quote(zero['title'])}'")!='0': raise RuntimeError('zero-duration exam was written')

        ended={**base,'title':prefix+'-ENDED','totalTime':30,'endTime':'2026-01-01 00:00:00'}
        ended_exam=request('/exam/api/exam/exam/save',ended,admin); ended_id=ended_exam['id']
        request('/exam/api/competency/exams/publish',{'examId':ended_id},admin)
        ended_candidate=request('/exam/api/candidate/save',{'examId':ended_id,'name':'Ended Verify','telephone':'138'+suffix[:8],'gender':'0'},admin)
        request('/exam/api/competency/participant/create-paper',{'examId':ended_id,'participantId':ended_candidate['id'],'participantType':'candidate','participantToken':ended_candidate['participantToken']},expect_error='测评已结束')
        request('/exam/api/exam/exam/delete',{'ids':[ended_id]},admin)
        if mysql(f"SELECT (SELECT COUNT(*) FROM el_exam WHERE id='{ended_id}')+(SELECT COUNT(*) FROM el_candidate WHERE id='{ended_candidate['id']}')")!='0': raise RuntimeError('ended exam cleanup failed')

        active={**base,'title':prefix+'-ACTIVE','totalTime':30,'endTime':'2027-12-31 23:59:59'}
        exam=request('/exam/api/exam/exam/save',active,admin); exam_id=exam['id']
        request('/exam/api/competency/exams/publish',{'examId':exam_id},admin)
        candidate=request('/exam/api/candidate/save',{'examId':exam_id,'name':'Boundary Verify','telephone':'137'+suffix[:8],'gender':'0'},admin); candidate_id=candidate['id']
        access=request('/exam/api/competency/participant/create-paper',{'examId':exam_id,'participantId':candidate_id,'participantType':'candidate','participantToken':candidate['participantToken']}); paper_id=access['paperId']; headers={'X-Competency-Token':access['paperToken']}

        request('/exam/api/competency/participant/submit',{'paperId':paper_id,'submitType':'timeout'},headers=headers,expect_error='超时由系统判定')
        request(f'/exam/api/candidate/{candidate_id}',{},admin,method='DELETE',expect_error='请删除所属胜任力测评')
        request('/exam/api/paper/paper/delete',{'ids':[paper_id]},admin,expect_error='请删除所属胜任力测评')

        mysql_exec(f"UPDATE el_paper SET limit_time=DATE_SUB(NOW(),INTERVAL 1 SECOND) WHERE id='{paper_id}';")
        submitted=request('/exam/api/competency/participant/submit',{'paperId':paper_id,'submitType':'manual'},headers=headers)
        if submitted['isComplete'] is not False: raise RuntimeError('timeout submit should be incomplete')
        request('/exam/api/competency/results/detail',{'paperId':paper_id},admin)
        request('/exam/api/competency/admin/report-data?paperId='+paper_id,None,admin,method='GET',expect_error='未完整作答')

        counts_before=mysql(f"SELECT CONCAT((SELECT COUNT(*) FROM el_exam_competency_question WHERE exam_id='{exam_id}'),'|',(SELECT COUNT(*) FROM el_paper WHERE exam_id='{exam_id}'),'|',(SELECT COUNT(*) FROM el_paper_qu WHERE paper_id='{paper_id}'),'|',(SELECT COUNT(*) FROM el_competency_dimension_result WHERE paper_id='{paper_id}'),'|',(SELECT COUNT(*) FROM el_competency_result WHERE paper_id='{paper_id}'),'|',(SELECT COUNT(*) FROM el_candidate WHERE exam_id='{exam_id}'))")
        if counts_before!='8|1|8|1|1|1': raise RuntimeError('pre-delete chain mismatch '+counts_before)
        request('/exam/api/exam/exam/delete',{'ids':[exam_id]},admin)
        remaining=mysql(f"SELECT (SELECT COUNT(*) FROM el_exam WHERE id='{exam_id}')+(SELECT COUNT(*) FROM el_exam_competency_dimension WHERE exam_id='{exam_id}')+(SELECT COUNT(*) FROM el_exam_competency_question WHERE exam_id='{exam_id}')+(SELECT COUNT(*) FROM el_paper WHERE exam_id='{exam_id}')+(SELECT COUNT(*) FROM el_paper_qu WHERE paper_id='{paper_id}')+(SELECT COUNT(*) FROM el_competency_dimension_result WHERE paper_id='{paper_id}')+(SELECT COUNT(*) FROM el_competency_result WHERE paper_id='{paper_id}')+(SELECT COUNT(*) FROM el_candidate WHERE exam_id='{exam_id}')+(SELECT COUNT(*) FROM el_tester WHERE exam_id='{exam_id}')")
        if remaining!='0': raise RuntimeError('full-chain delete remaining='+remaining)
        exam_id=candidate_id=paper_id=None
        print('COMPETENCY_BOUNDARY_DELETE_VERIFY_OK')
        print('zero_duration_rejected=1')
        print('ended_exam_start_rejected=1')
        print('client_timeout_rejected=1')
        print('direct_entity_delete_rejected=1')
        print('incomplete_report_rejected=1')
        print('pre_delete_counts=8|1|8|1|1|1')
        print('full_chain_remaining=0')
    finally:
        if exam_id:
            subprocess.run(['sudo','-n','mysql','element','-e',f"DELETE FROM el_competency_dimension_result WHERE paper_id IN (SELECT id FROM el_paper WHERE exam_id='{exam_id}'); DELETE FROM el_competency_result WHERE exam_id='{exam_id}'; DELETE pq FROM el_paper_qu pq JOIN el_paper p ON p.id=pq.paper_id WHERE p.exam_id='{exam_id}'; DELETE FROM el_candidate WHERE exam_id='{exam_id}'; DELETE FROM el_tester WHERE exam_id='{exam_id}'; DELETE FROM el_paper WHERE exam_id='{exam_id}'; DELETE FROM el_exam_competency_question WHERE exam_id='{exam_id}'; DELETE FROM el_exam_competency_dimension WHERE exam_id='{exam_id}'; DELETE FROM el_exam WHERE id='{exam_id}';"],check=False,stdout=subprocess.DEVNULL,stderr=subprocess.DEVNULL)
        subprocess.run(['redis-cli','-n',redis_db,'DEL',redis_key],check=False,stdout=subprocess.DEVNULL)
        remaining=mysql(f"SELECT COUNT(*) FROM el_exam WHERE title LIKE '{prefix}%';")
        print('cleanup_remaining='+remaining)

if __name__=='__main__':
    try: main()
    except Exception as error:
        print('COMPETENCY_BOUNDARY_DELETE_VERIFY_FAILED: '+str(error),file=sys.stderr); sys.exit(1)
