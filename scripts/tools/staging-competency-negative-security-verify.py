#!/usr/bin/env python3
import base64, hashlib, hmac, json, re, subprocess, time, urllib.error, urllib.request, uuid
from pathlib import Path

APP_DIR = Path('/opt/talent-assessment')
BASE = 'http://127.0.0.1:8092'


def b64url(value): return base64.urlsafe_b64encode(value).rstrip(b'=').decode()
def config_text():
    env = subprocess.run(['systemctl','show','talent-assessment','--property=Environment','--value'],check=True,capture_output=True,text=True).stdout
    match = re.search(r'(?:^|\s)APP_ENV=([^\s]+)', env); name = match.group(1) if match else 'local'
    return '\n'.join(p.read_text(encoding='utf-8') for p in [APP_DIR/'configs'/'application.yml',APP_DIR/'configs'/f'application-{name}.yml'] if p.exists())
def scalar(text, section, key, default=''):
    value=default
    for sm in re.finditer(rf'(?ms)^{re.escape(section)}:\s*\n(.*?)(?=^[A-Za-z][\w-]*:\s*$|\Z)',text):
        km=re.search(rf'(?m)^\s+{re.escape(key)}:\s*(.+?)\s*$',sm.group(1))
        if km:value=km.group(1).strip().strip('"\'')
    return value
def sign(secret, claims):
    h=b64url(json.dumps({'alg':'HS512','typ':'JWT'},separators=(',',':')).encode());p=b64url(json.dumps(claims,separators=(',',':')).encode());raw=f'{h}.{p}'
    return f'{raw}.{b64url(hmac.new(secret.encode(),raw.encode(),hashlib.sha512).digest())}'
def api_raw(path,body,token=None,headers=None):
    hs={'Content-Type':'application/json'}
    if token:hs['Authorization']='Bearer '+token
    if headers:hs.update(headers)
    req=urllib.request.Request(BASE+path,data=json.dumps(body,ensure_ascii=False).encode(),headers=hs,method='POST')
    try:
        with urllib.request.urlopen(req,timeout=60) as r:return r.status,json.loads(r.read().decode())
    except urllib.error.HTTPError as e:return e.code,json.loads(e.read().decode())
def api(path,body,token=None,headers=None):
    _,p=api_raw(path,body,token,headers)
    if p.get('code') not in (0,200):raise RuntimeError(f'{path}: {p.get("msg")}')
    return p.get('data')
def expect_error(path,body,contains,token=None,headers=None):
    _,p=api_raw(path,body,token,headers)
    if p.get('code') in (0,200) or contains not in str(p.get('msg','')):raise RuntimeError(f'{path} expected {contains!r}, got {p}')
    return p.get('msg')
def mysql(sql):return subprocess.run(['sudo','-n','mysql','element','-Nse',sql],check=True,capture_output=True,text=True).stdout.strip()


def main():
    suffix=uuid.uuid4().hex[:10];text=config_text();secret=scalar(text,'jwt','secret');redis_db=scalar(text,'redis','db','1');tid='negative-'+suffix;key='login_tokens:'+tid
    now_ms=int(time.time()*1000);login={'userId':1,'token':tid,'loginTime':now_ms,'expireTime':now_ms+1800000,'permissions':['*:*:*'],'roles':['admin'],'user':None}
    subprocess.run(['redis-cli','-n',redis_db,'SET',key,json.dumps(login,separators=(',',':')),'EX','1800'],check=True,stdout=subprocess.DEVNULL);admin=sign(secret,{'login_user_key':tid})
    exam_id=None;source_original=None
    try:
        dims=api('/exam/api/competency/dimensions/list',{},admin);selected=[x['id'] for x in dims if x['code'] in ('D01','D02')]
        exam=api('/exam/api/exam/exam/save',{'title':'NEGATIVE-'+suffix,'content':'negative/security/snapshot','assessmentType':'competency','scoringMode':'competency_average','competencyReportAudience':'leader','dimensionIds':selected,'joinType':1,'openType':1,'isOpen':1,'answerType':1,'state':0,'totalTime':30,'repoList':[],'departIds':[]},admin);exam_id=exam['id'];api('/exam/api/competency/exams/publish',{'examId':exam_id},admin)
        def create(name,tel):
            c=api('/exam/api/candidate/save',{'examId':exam_id,'name':name,'telephone':tel,'gender':'0'},admin);a=api('/exam/api/competency/participant/create-paper',{'examId':exam_id,'participantId':c['id'],'participantType':'candidate','participantToken':c['participantToken']});d=api('/exam/api/competency/participant/paper-detail',{'paperId':a['paperId']},headers={'X-Competency-Token':a['paperToken']});return c,a,d
        c1,a1,d1=create('Negative A','13710000001');c2,a2,d2=create('Negative B','13710000002')
        expect_error('/exam/api/competency/participant/fill-answer',{'paperId':a1['paperId'],'paperQuestionId':d1['questions'][0]['id'],'rawValue':0},'1到5',headers={'X-Competency-Token':a1['paperToken']})
        expect_error('/exam/api/competency/participant/fill-answer',{'paperId':a1['paperId'],'paperQuestionId':d2['questions'][0]['id'],'rawValue':3},'不属于此试卷',headers={'X-Competency-Token':a1['paperToken']})
        for q in d1['questions']:api('/exam/api/competency/participant/fill-answer',{'paperId':a1['paperId'],'paperQuestionId':q['id'],'rawValue':3},headers={'X-Competency-Token':a1['paperToken']})
        api('/exam/api/competency/participant/submit',{'paperId':a1['paperId'],'submitType':'manual'},headers={'X-Competency-Token':a1['paperToken']})
        expect_error('/exam/api/competency/participant/fill-answer',{'paperId':a1['paperId'],'paperQuestionId':d1['questions'][0]['id'],'rawValue':4},'already completed',headers={'X-Competency-Token':a1['paperToken']})
        mysql(f"UPDATE el_paper SET limit_time=DATE_SUB(NOW(),INTERVAL 1 MINUTE) WHERE id='{a2['paperId']}'")
        expired=api('/exam/api/competency/participant/fill-answer',{'paperId':a2['paperId'],'paperQuestionId':d2['questions'][0]['id'],'rawValue':3},headers={'X-Competency-Token':a2['paperToken']})
        if not expired.get('expired'):raise RuntimeError('expired save did not trigger trusted timeout submit')

        expect_error('/exam/api/competency/participant/create-paper',{'examId':exam_id,'participantId':c1['id'],'participantType':'candidate'},'认证失败')
        expect_error('/exam/api/competency/participant/create-paper',{'examId':exam_id,'participantId':c1['id'],'participantType':'candidate','participantToken':a1['paperToken']},'认证失败')
        expect_error('/exam/api/competency/participant/paper-detail',{'paperId':a2['paperId']},'试卷认证失败',headers={'X-Competency-Token':a1['paperToken']})
        expect_error('/exam/api/competency/participant/create-paper',{'examId':exam_id,'participantId':c2['id'],'participantType':'candidate','participantToken':c1['participantToken']},'不匹配')
        _,unauth=api_raw('/exam/api/competency/results/paging',{'examId':exam_id})
        if unauth.get('code') != 401:raise RuntimeError(f'unauth management expected 401, got {unauth}')

        source_id=mysql("SELECT source_qu_id FROM el_exam_competency_question WHERE exam_id='%s' AND question_code='D01-Q01'"%exam_id)
        source_original=json.loads(json.dumps(api('/exam/api/competency/questions/paging',{'current':1,'size':10,'questionCode':'D01-Q01'},admin)['records'][0]))
        snapshot_before=mysql("SELECT CONCAT(question_content,'|',observation_point,'|',scoring_direction,'|',source_update_time) FROM el_exam_competency_question WHERE exam_id='%s' AND question_code='D01-Q01'"%exam_id)
        api('/exam/api/competency/questions/update',{'id':source_id,'content':source_original['content']+'-MUTATED','observationPoint':source_original['observationPoint']+'-MUTATED','scoringDirection':'reverse' if source_original['scoringDirection']=='forward' else 'forward','questionStatus':1,'remark':'snapshot mutation verification'},admin)
        snapshot_after=mysql("SELECT CONCAT(question_content,'|',observation_point,'|',scoring_direction,'|',source_update_time) FROM el_exam_competency_question WHERE exam_id='%s' AND question_code='D01-Q01'"%exam_id)
        paper_content=mysql("SELECT q.question_content FROM el_paper_qu pq JOIN el_exam_competency_question q ON q.id=pq.exam_question_id WHERE pq.paper_id='%s' AND q.question_code='D01-Q01'"%a1['paperId'])
        if snapshot_before!=snapshot_after or paper_content!=source_original['content']:raise RuntimeError('published snapshot/history changed after source mutation')
        print('COMPETENCY_NEGATIVE_SECURITY_VERIFY_OK')
        print('participant_negatives=invalid_raw|foreign_question|finished_write|expired_write')
        print('security=missing_token|wrong_purpose|cross_paper|cross_participant|management_unauth')
        print('snapshot_immutable=D01-Q01')
    finally:
        if source_original:
            try:api('/exam/api/competency/questions/update',{'id':source_original['id'],'content':source_original['content'],'observationPoint':source_original['observationPoint'],'scoringDirection':source_original['scoringDirection'],'questionStatus':source_original['questionStatus'],'remark':source_original.get('remark','')},admin)
            except Exception as e:print('source_restore_error='+str(e))
        if exam_id:
            try:api('/exam/api/exam/exam/delete',{'ids':[exam_id]},admin)
            except Exception as e:print('cleanup_error='+str(e))
        subprocess.run(['redis-cli','-n',redis_db,'DEL',key],check=False,stdout=subprocess.DEVNULL)
        if exam_id:
            remaining=mysql(f"SELECT (SELECT COUNT(*) FROM el_exam WHERE id='{exam_id}')+(SELECT COUNT(*) FROM el_paper WHERE exam_id='{exam_id}')+(SELECT COUNT(*) FROM el_candidate WHERE exam_id='{exam_id}')+(SELECT COUNT(*) FROM el_competency_result WHERE exam_id='{exam_id}')")
            print('cleanup_remaining='+remaining)
            if remaining!='0':raise RuntimeError('cleanup remaining='+remaining)

if __name__=='__main__':main()
