"""
新旧后端接口对比测试骨架。
用法：
  python api_diff.py --legacy http://127.0.0.1:18091 --refactor http://127.0.0.1:8092

逻辑：
1. 从 legacy 拉 captcha，用 redis 中的答案完成 /login，得到 legacy token
2. 用 refactor 另外走一遍 captcha+login，得到 refactor token
3. 对每个 API 用两边的 token 请求，对比 JSON 结构与关键字段
4. 差异打到 stdout + 文件

当前仅覆盖核心链路，模块接口按需扩展 cases。
"""

from __future__ import annotations

import argparse
import json
import subprocess
import sys
from pathlib import Path
from typing import Any
from urllib.parse import urljoin

import urllib.request

REDIS_CLI = r"E:\Talent-Assessment\Legacy Java System\tools\redis-portable\redis\redis-cli.exe"


def http(method: str, url: str, headers: dict[str, str] | None = None, body: Any = None) -> dict:
    data = None
    h = {"Accept": "application/json"}
    if body is not None:
        data = json.dumps(body).encode()
        h["Content-Type"] = "application/json"
    if headers:
        h.update(headers)
    req = urllib.request.Request(url, data=data, method=method, headers=h)
    with urllib.request.urlopen(req, timeout=10) as resp:
        return json.loads(resp.read().decode())


def redis_get(key: str) -> str:
    out = subprocess.check_output([REDIS_CLI, "-n", "1", "GET", key], text=True)
    return out.strip()


def login(base: str, username: str = "admin", password: str = "cp@1234") -> str:
    cap = http("GET", urljoin(base, "/captchaImage"))
    ans = redis_get(f"captcha_codes:{cap['uuid']}")
    r = http("POST", urljoin(base, "/login"), body={
        "username": username, "password": password, "code": ans, "uuid": cap["uuid"],
    })
    return r["token"]


CASES = [
    ("GET", "/getInfo", None),
    ("GET", "/getRouters", None),
    ("GET", "/system/dict/data/type/sys_normal_disable", None),
    ("POST", "/api/qu/repo/list", {}),
]


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--legacy", default="http://127.0.0.1:18091")
    ap.add_argument("--refactor", default="http://127.0.0.1:8092")
    ap.add_argument("--out", default="api_diff.txt")
    args = ap.parse_args()

    l_tok = login(args.legacy)
    r_tok = login(args.refactor)

    report = []
    for method, path, body in CASES:
        try:
            a = http(method, urljoin(args.legacy, path), {"Authorization": f"Bearer {l_tok}"}, body)
        except Exception as e:
            a = {"_err": str(e)}
        try:
            b = http(method, urljoin(args.refactor, path), {"Authorization": f"Bearer {r_tok}"}, body)
        except Exception as e:
            b = {"_err": str(e)}
        same = normalize(a) == normalize(b)
        report.append({
            "case": f"{method} {path}",
            "legacy_keys": sorted(list(a.keys()))[:15] if isinstance(a, dict) else type(a).__name__,
            "refactor_keys": sorted(list(b.keys()))[:15] if isinstance(b, dict) else type(b).__name__,
            "same_shape": same,
        })

    Path(args.out).write_text(json.dumps(report, indent=2, ensure_ascii=False), encoding="utf-8")
    print(json.dumps(report, indent=2, ensure_ascii=False))


def normalize(x: Any) -> Any:
    """结构比对：去除时间戳、随机 id 等噪声。"""
    if isinstance(x, dict):
        return {k: normalize(v) for k, v in sorted(x.items()) if k not in {"loginDate", "createTime", "updateTime", "token"}}
    if isinstance(x, list):
        return [normalize(i) for i in x]
    return type(x).__name__


if __name__ == "__main__":
    main()
