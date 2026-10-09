#!/usr/bin/env python3
"""Validate a real phase-1 v2 three-sheet export using only the standard library."""

from __future__ import annotations

import argparse
import hashlib
import re
import zipfile
import xml.etree.ElementTree as ET
from pathlib import Path

SHEET_NS = "{http://schemas.openxmlformats.org/spreadsheetml/2006/main}"
DOC_REL_NS = "{http://schemas.openxmlformats.org/officeDocument/2006/relationships}"
PKG_REL_NS = "{http://schemas.openxmlformats.org/package/2006/relationships}"


def read_workbook(path: Path) -> dict[str, list[list[object]]]:
    with zipfile.ZipFile(path) as archive:
        shared: list[str] = []
        if "xl/sharedStrings.xml" in archive.namelist():
            root = ET.fromstring(archive.read("xl/sharedStrings.xml"))
            shared = ["".join(item.text or "" for item in value.iter(SHEET_NS + "t")) for value in root.findall(SHEET_NS + "si")]
        workbook = ET.fromstring(archive.read("xl/workbook.xml"))
        relationships = ET.fromstring(archive.read("xl/_rels/workbook.xml.rels"))
        targets = {item.get("Id"): item.get("Target", "") for item in relationships.findall(PKG_REL_NS + "Relationship")}
        result: dict[str, list[list[object]]] = {}
        sheets = workbook.find(SHEET_NS + "sheets")
        assert sheets is not None
        for sheet in sheets:
            name = sheet.get("name", "")
            target = targets[sheet.get(DOC_REL_NS + "id", "")].lstrip("/")
            if not target.startswith("xl/"):
                target = "xl/" + target
            xml = ET.fromstring(archive.read(target))
            rows: list[list[object]] = []
            for row in xml.iter(SHEET_NS + "row"):
                values: list[object] = []
                for cell in row.findall(SHEET_NS + "c"):
                    column = re.match(r"[A-Z]+", cell.get("r", "")).group()
                    index = 0
                    for character in column:
                        index = index * 26 + ord(character) - 64
                    while len(values) < index:
                        values.append(None)
                    kind = cell.get("t")
                    raw = cell.find(SHEET_NS + "v")
                    inline = cell.find(SHEET_NS + "is")
                    value: object = ""
                    if kind == "s" and raw is not None:
                        value = shared[int(raw.text or "0")]
                    elif kind == "inlineStr" and inline is not None:
                        value = "".join(item.text or "" for item in inline.iter(SHEET_NS + "t"))
                    elif raw is not None:
                        value = raw.text or ""
                        if value and re.fullmatch(r"-?\d+(?:\.\d+)?", str(value)):
                            value = float(value) if "." in str(value) else int(value)
                    values[index - 1] = value
                rows.append(values)
            result[name] = rows
    return result


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("summary", type=Path)
    parser.add_argument("answers", type=Path)
    parser.add_argument("facts", type=Path)
    args = parser.parse_args()
    assert args.summary.read_bytes() == args.answers.read_bytes(), "two export endpoints returned different workbooks"
    workbook = read_workbook(args.summary)
    assert list(workbook) == ["结果汇总", "逐题明细", "题目字典"]

    summary_headers = workbook["结果汇总"][0]
    summary_rows = workbook["结果汇总"][1:]
    assert len(summary_headers) == 75 and len(summary_rows) == 3
    required_summary = [
        "v2结果RunID", "产品版本", "计分版本", "内容版本", "报告模板版本", "报告对象",
        "总体得分", "总体等级", "总体常模", "总体常模比较", "任务管理类-得分", "人际管理类-得分",
        "自我管理类-得分", "效度原始分", "效度状态", "A1-01 逻辑思维-得分", "C1-03 敬业奉献-常模比较",
    ]
    assert all(value in summary_headers for value in required_summary)
    indexes = {value: index for index, value in enumerate(summary_headers)}
    for row in summary_rows:
        assert row[indexes["产品版本"]] == "competency-frontline-phase1-v2"
        assert row[indexes["计分版本"]] == "competency-phase1-scoring-v2"
        assert row[indexes["报告对象"]] == "基层员工"
        assert 0 <= float(row[indexes["总体得分"]]) <= 100
        assert row[indexes["总体等级"]] in ("优秀", "良好", "合格", "薄弱", "不足")
        assert row[indexes["效度状态"]] in ("有效", "存疑")

    summary_by_run = {str(row[indexes["v2结果RunID"]]): row for row in summary_rows}
    level_labels = {"excellent": "优秀", "good": "良好", "qualified": "合格", "weak": "薄弱", "insufficient": "不足"}
    overall_comparisons = {"superior": "优势突出", "above_norm": "优于常模分", "at_norm": "与常模分持平", "below_norm": "低于常模分"}
    module_comparisons = {"standout": "优势突出", "above_norm": "略高于常模分", "at_norm": "与常模分持平", "below_norm": "低于常模分"}
    module_names = {"task_management": "任务管理类", "interpersonal_management": "人际管理类", "self_management": "自我管理类"}
    dimension_prefixes = {
        "competency-logical-reasoning": "A1-01 逻辑思维", "competency-plan-execution": "A1-02 计划执行",
        "competency-digital-application": "A1-03 数字应用", "competency-achievement-orientation": "A1-04 成就导向",
        "competency-continuous-learning": "A1-05 持续学习", "competency-communication": "B1-01 沟通表达",
        "competency-cooperation": "B1-02 合作意识", "competency-truth-pragmatism": "C1-01 求真务实",
        "competency-self-discipline": "C1-02 自律性", "competency-dedication": "C1-03 敬业奉献",
    }
    fact_count = 0
    for line in args.facts.read_text(encoding="utf-8").splitlines():
        values = line.split("\t")
        kind, run_id = values[:2]
        row = summary_by_run[run_id]
        if kind == "RUN":
            _, _, score, level, norm, comparison, validity_score, validity_status = values
            assert abs(float(row[indexes["总体得分"]]) - float(score)) < 0.011
            assert row[indexes["总体等级"]] == level_labels[level]
            assert abs(float(row[indexes["总体常模"]]) - float(norm)) < 0.011
            assert row[indexes["总体常模比较"]] == overall_comparisons[comparison]
            assert abs(float(row[indexes["效度原始分"]]) - float(validity_score)) < 0.011
            assert row[indexes["效度状态"]] == {"good": "有效", "questionable": "存疑"}[validity_status]
        elif kind == "MODULE":
            _, _, code, score, level, norm, comparison = values
            prefix = module_names[code]
            assert abs(float(row[indexes[prefix + "-得分"]]) - float(score)) < 0.011
            assert row[indexes[prefix + "-等级"]] == level_labels[level]
            assert abs(float(row[indexes[prefix + "-常模"]]) - float(norm)) < 0.011
            assert row[indexes[prefix + "-常模比较"]] == module_comparisons[comparison]
        elif kind == "DIMENSION":
            _, _, dimension_id, score, level, norm = values
            prefix = dimension_prefixes[dimension_id]
            assert abs(float(row[indexes[prefix + "-得分"]]) - float(score)) < 0.011
            assert row[indexes[prefix + "-等级"]] == level_labels[level]
            assert abs(float(row[indexes[prefix + "-常模"]]) - float(norm)) < 0.011
        else:
            raise AssertionError(f"unknown fact kind: {kind}")
        fact_count += 1
    assert fact_count == 42, f"database facts={fact_count}, want 42"

    detail_headers = workbook["逐题明细"][0]
    assert len(detail_headers) == 20 and len(workbook["逐题明细"]) == 271
    assert all(value in detail_headers for value in ["v2结果RunID", "原维度编号", "v2维度编号", "v2维度名称", "v2模块", "原始选择值", "原始选择文本", "最终题目得分"])

    dictionary_headers = workbook["题目字典"][0]
    assert len(dictionary_headers) == 14 and len(workbook["题目字典"]) == 91
    assert all(value in dictionary_headers for value in ["v2稳定维度ID", "v2维度编号", "v2维度名称", "v2模块"])

    digest = hashlib.sha256(args.summary.read_bytes()).hexdigest()
    print(f"COMPETENCY_V2_EXPORT_XLSX_PASS sha256={digest} summary=3 answers=270 dictionary=90 columns=75/20/14 dbFacts=42")


if __name__ == "__main__":
    main()
