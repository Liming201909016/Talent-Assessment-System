#!/usr/bin/env python3
import base64, hashlib, hmac, json, re, subprocess, time, urllib.request, urllib.error, uuid
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
    header=b64url(json.dumps({'alg':'HS512','typ':'JWT'},separators=(',',':')).encode());payload=b64url(json.dumps({'login_user_key':token_id},separators=(',',':')).encode());raw=f'{header}.{payload}';return f'{raw}.{b64url(hmac.new(secret.encode(),raw.encode(),hashlib.sha512).digest())}'
def api(path,body,token=None,headers=None,method='POST'):
    hs={'Content-Type':'application/json'}
    if token: hs['Authorization']='Bearer '+token
    if headers: hs.update(headers)
    data=None if method=='GET' else json.dumps(body,ensure_ascii=False).encode()
    req=urllib.request.Request(BASE+path,data=data,headers=hs,method=method)
    try:
        with urllib.request.urlopen(req,timeout=30) as response: payload=json.loads(response.read().decode())
    except urllib.error.HTTPError as error: raise RuntimeError(f'{path} HTTP {error.code}: {error.read().decode(errors="replace")}')
    if payload.get('code') not in (0,200): raise RuntimeError(f'{path}: {payload.get("msg")}')
    return payload.get('data')
def expect_api_error(path,body,contains,headers=None):
    hs={'Content-Type':'application/json'}
    if headers: hs.update(headers)
    req=urllib.request.Request(BASE+path,data=json.dumps(body).encode(),headers=hs,method='POST')
    with urllib.request.urlopen(req,timeout=30) as response: payload=json.loads(response.read().decode())
    if payload.get('code')==0 or contains not in payload.get('msg',''): raise RuntimeError(f'{path} expected error containing {contains}: {payload}')
def mysql(sql): return subprocess.run(['sudo','-n','mysql','element','-Nse',sql],check=True,capture_output=True,text=True).stdout.strip()
def mysql_exec(sql): subprocess.run(['sudo','-n','mysql','element','-e',sql],check=True)

def main():
    suffix=uuid.uuid4().hex[:12]; exam_id=None; paper_id=None; candidate_id=None
    qids=[f'rt-{suffix}-q{i}' for i in range(1,5)]; codes=[f'RT-{suffix}-Q{i}' for i in range(1,5)]
    text=config_text();secret=scalar(text,'jwt','secret');redis_db=scalar(text,'redis','db','1')
    token_id='competency-runtime-'+suffix;now_ms=int(time.time()*1000);redis_key='login_tokens:'+token_id
    login_user={'userId':1,'token':token_id,'loginTime':now_ms,'expireTime':now_ms+1800000,'permissions':['*:*:*'],'roles':['admin'],'user':None}
    subprocess.run(['redis-cli','-n',redis_db,'SET',redis_key,json.dumps(login_user,separators=(',',':')),'EX','1800'],check=True,stdout=subprocess.DEVNULL)
    admin=sign(secret,token_id)
    try:
        values=[]
        for i,(qid,code,dim,item,direction) in enumerate(zip(qids,codes,['competency-d01','competency-d01','competency-d02','competency-d02'],[910001,910002,910001,910002],['forward','forward','reverse','reverse'])):
            values.append(f"('{qid}',1,0,'runtime verification {i+1}','temporary','{code}','{dim}',{item},'runtime verification','{direction}',0,NOW(),NOW())")
        mysql_exec("INSERT INTO el_qu (id,qu_type,level,content,remark,question_code,dimension_id,dimension_item_no,observation_point,scoring_direction,question_status,create_time,update_time) VALUES "+','.join(values)+';')
        baseline_counts={row.split('|')[0]:int(row.split('|')[1]) for row in mysql("SELECT CONCAT(d.code,'|',COUNT(q.id)) FROM el_competency_dimension d LEFT JOIN el_qu q ON q.dimension_id=d.id AND q.question_status=0 WHERE d.code IN ('D01','D02') GROUP BY d.code;").splitlines()}
        dims=api('/exam/api/competency/dimensions/list',{},admin); selected=[d['id'] for d in dims if d['code'] in ('D01','D02')]
        if len(selected)!=2: raise RuntimeError('D01/D02 not found')
        for d in dims:
            if d['id'] in selected and d.get('questionCount')!=baseline_counts[d['code']]: raise RuntimeError(f'question count mismatch {d}')
        exam=api('/exam/api/exam/exam/save',{'title':'RUNTIME-'+suffix,'content':'temporary','assessmentType':'competency','scoringMode':'competency_average','competencyReportAudience':'leader','dimensionIds':selected,'joinType':1,'openType':1,'isOpen':1,'answerType':1,'state':0,'totalTime':30,'repoList':[],'departIds':[]},admin);exam_id=exam['id']
        published=api('/exam/api/competency/exams/publish',{'examId':exam_id},admin)
        expected_questions=sum(baseline_counts.values())
        if published['questionCount']!=expected_questions or published['dimensionCount']!=2: raise RuntimeError(f'publish mismatch {published}')
        repeated=api('/exam/api/competency/exams/publish',{'examId':exam_id},admin)
        if not repeated['alreadyPublished'] or repeated['questionCount']!=expected_questions: raise RuntimeError('publish not idempotent')
        telephone='139'+suffix[:8]
        candidate=api('/exam/api/candidate/save',{'examId':exam_id,'name':'Runtime Verify','telephone':telephone,'gender':'0'},admin);candidate_id=candidate['id'];participant_token=candidate.get('participantToken')
        if not participant_token: raise RuntimeError('participant token missing')
        access=api('/exam/api/competency/participant/create-paper',{'examId':exam_id,'participantId':candidate_id,'participantType':'candidate','participantToken':participant_token});paper_id=access['paperId'];paper_token=access['paperToken']
        restored=api('/exam/api/competency/participant/create-paper',{'examId':exam_id,'participantId':candidate_id,'participantType':'candidate','participantToken':participant_token})
        if restored['paperId']!=paper_id: raise RuntimeError('paper restore created another paper')
        detail=api('/exam/api/competency/participant/paper-detail',{'paperId':paper_id},headers={'X-Competency-Token':paper_token})
        order=[q['id'] for q in detail['questions']]
        detail2=api('/exam/api/competency/participant/paper-detail',{'paperId':paper_id},headers={'X-Competency-Token':paper_token})
        if order!=[q['id'] for q in detail2['questions']] or len(set(order))!=expected_questions: raise RuntimeError('paper order not fixed/complete')
        leaked=set(detail['questions'][0]).intersection({'dimensionId','scoringDirection','finalScore'})
        if leaked: raise RuntimeError(f'participant response leaked {leaked}')
        expect_api_error('/exam/api/competency/participant/submit',{'paperId':paper_id,'submitType':'manual'},'尚未作答',{'X-Competency-Token':paper_token})
        for question in detail['questions']: api('/exam/api/competency/participant/fill-answer',{'paperId':paper_id,'paperQuestionId':question['id'],'rawValue':3},headers={'X-Competency-Token':paper_token})
        submitted=api('/exam/api/competency/participant/submit',{'paperId':paper_id,'submitType':'manual'},headers={'X-Competency-Token':paper_token})
        if not submitted['isComplete']: raise RuntimeError('submit incomplete')
        repeated_submit=api('/exam/api/competency/participant/submit',{'paperId':paper_id,'submitType':'manual'},headers={'X-Competency-Token':paper_token})
        if not repeated_submit['alreadySubmitted']: raise RuntimeError('submit not idempotent')
        result=api('/exam/api/competency/results/detail',{'paperId':paper_id},admin)
        overall=result['result'];dims_result=result['dimensions']
        if float(overall['overallScore'])!=6.0 or float(overall['evaluationAverage'])!=3.0 or overall['evaluationLevel']!='good' or overall['reportAudience']!='leader': raise RuntimeError(f'score mismatch {overall}')
        if len(dims_result)!=2 or [float(x['dimensionScore']) for x in dims_result]!=[3.0,3.0]: raise RuntimeError(f'dimension score mismatch {dims_result}')
        report=api('/exam/api/competency/admin/report-data?paperId='+paper_id,None,admin,method='GET')
        if report['result']['reportAudience']!='leader' or report['reportTextReady'] is not False: raise RuntimeError('report snapshot/text state mismatch')
        db=mysql(f"SELECT CONCAT((SELECT COUNT(*) FROM el_exam_competency_question WHERE exam_id='{exam_id}'),'|',(SELECT COUNT(*) FROM el_paper_qu WHERE paper_id='{paper_id}'),'|',(SELECT COUNT(*) FROM el_competency_dimension_result WHERE paper_id='{paper_id}'),'|',(SELECT COUNT(*) FROM el_competency_result WHERE paper_id='{paper_id}'))")
        expected_db=f'{expected_questions}|{expected_questions}|2|1'
        if db!=expected_db: raise RuntimeError('DB counts '+db)
        print('COMPETENCY_RUNTIME_VERIFY_OK');print(f'publish={expected_questions}_questions_2_dimensions');print('paper_restore_same=1');print('fixed_complete_order=1');print('manual_incomplete_rejected=1');print('scores=3.000000,3.000000 overall=6.000000 average=3.000000 good');print('audience=leader');print('db_counts='+expected_db)
    finally:
        if paper_id: mysql_exec(f"DELETE FROM el_competency_dimension_result WHERE paper_id='{paper_id}'; DELETE FROM el_competency_result WHERE paper_id='{paper_id}'; DELETE FROM el_paper_qu_answer WHERE paper_id='{paper_id}'; DELETE FROM el_paper_qu WHERE paper_id='{paper_id}'; UPDATE el_candidate SET paper_id=NULL,end_time=NULL WHERE id='{candidate_id}'; DELETE FROM el_paper WHERE id='{paper_id}';")
        if candidate_id: mysql_exec(f"DELETE FROM el_candidate WHERE id='{candidate_id}';")
        if exam_id: mysql_exec(f"DELETE FROM el_exam_competency_question WHERE exam_id='{exam_id}'; DELETE FROM el_exam_competency_dimension WHERE exam_id='{exam_id}'; DELETE FROM el_exam WHERE id='{exam_id}';")
        mysql_exec("DELETE FROM el_qu WHERE id IN ("+','.join("'"+x+"'" for x in qids)+');')
        subprocess.run(['redis-cli','-n',redis_db,'DEL',redis_key],check=False,stdout=subprocess.DEVNULL)
        remaining=mysql("SELECT (SELECT COUNT(*) FROM el_qu WHERE id IN ("+','.join("'"+x+"'" for x in qids)+f"))+(SELECT COUNT(*) FROM el_exam WHERE id='{exam_id or ''}')+(SELECT COUNT(*) FROM el_paper WHERE id='{paper_id or ''}');")
        print('cleanup_remaining='+remaining)
if __name__=='__main__': main()
