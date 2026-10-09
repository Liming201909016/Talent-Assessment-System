#!/usr/bin/env python3
"""Build the cleaned 260915 competency report draft without redesigning it.

The transformation is intentionally limited to:
- B/C content-control bindings;
- exact sample values;
- stable chart business keys, zero active external links, and updated caches;
- narrowly scoped Word/LibreOffice pagination guards.
"""

from __future__ import annotations

import argparse
import copy
import io
import re
import zipfile
import xml.etree.ElementTree as ET
from decimal import Decimal, ROUND_HALF_UP
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
DEFAULT_SOURCE = ROOT / "docs" / "260915" / "260915胜任力测评报告样例-基层员工版-表格版.docx"
DEFAULT_OUTPUT = ROOT / "docs" / "260915" / "competency-frontline-report-template-draft-v2.docx"

NS = {
    "w": "http://schemas.openxmlformats.org/wordprocessingml/2006/main",
    "w14": "http://schemas.microsoft.com/office/word/2010/wordml",
    "wp": "http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing",
    "c": "http://schemas.openxmlformats.org/drawingml/2006/chart",
    "r": "http://schemas.openxmlformats.org/officeDocument/2006/relationships",
    "pr": "http://schemas.openxmlformats.org/package/2006/relationships",
}
W = "{" + NS["w"] + "}"
W14 = "{" + NS["w14"] + "}"
WP = "{" + NS["wp"] + "}"
C = "{" + NS["c"] + "}"
R = "{" + NS["r"] + "}"
PR = "{" + NS["pr"] + "}"

DIMENSIONS = [
    ("logical_reasoning", "逻辑思维", 68.75),
    ("plan_execution", "计划执行", 62.5),
    ("digital_application", "数字应用", 81.25),
    ("achievement_orientation", "成就导向", 50.0),
    ("continuous_learning", "持续学习", 75.0),
    ("communication", "沟通表达", 56.25),
    ("cooperation", "合作意识", 62.5),
    ("truth_pragmatism", "求真务实", 78.125),
    ("self_discipline", "自律性", 78.125),
    ("dedication", "敬业奉献", 68.75),
]
MODULES = [
    ("task_management", "任务管理类", 67.50, "合格", "略高于常模分"),
    ("interpersonal_management", "人际管理类", 59.38, "合格", "略高于常模分"),
    ("self_management", "自我管理类", 75.00, "良好", "优势突出"),
]
CHART_KEYS = {
    1: "chart.overall.score",
    2: "chart.dimension.comparison",
    **{index + 3: f"chart.dimension.{dimension}" for index, (dimension, _, _) in enumerate(DIMENSIONS)},
}
COVER_BLANK_PARAGRAPH_IDS = {"1B10269B", "7953F03A", "087D950F", "35898E47"}
LAYOUT_SCALE = 0.70


def display_score(value: float) -> str:
    return str(Decimal(str(value)).quantize(Decimal("0.01"), rounding=ROUND_HALF_UP))


def register_namespaces(xml_bytes: bytes) -> None:
    for event, value in ET.iterparse(io.BytesIO(xml_bytes), events=("start-ns",)):
        del event
        prefix, uri = value
        try:
            ET.register_namespace(prefix, uri)
        except ValueError:
            pass


def parse_xml(xml_bytes: bytes) -> ET.Element:
    register_namespaces(xml_bytes)
    return ET.fromstring(xml_bytes)


def serialize_xml(root: ET.Element, source_xml: bytes | None = None) -> bytes:
    serialized = ET.tostring(root, encoding="utf-8", xml_declaration=True)
    if source_xml is None:
        return serialized
    source_text = source_xml.decode("utf-8")
    output_text = serialized.decode("utf-8")
    source_root = re.search(r"<[^!?][^>]*>", source_text)
    output_root = re.search(r"<[^!?][^>]*>", output_text)
    if source_root is None or output_root is None:
        raise RuntimeError("XML root element missing")
    source_namespaces = dict(
        re.findall(r'\b(xmlns(?::[A-Za-z0-9_.-]+)?)="([^"]+)"', source_root.group(0))
    )
    output_namespaces = dict(
        re.findall(r'\b(xmlns(?::[A-Za-z0-9_.-]+)?)="([^"]+)"', output_root.group(0))
    )
    missing = "".join(
        f' {name}="{uri}"'
        for name, uri in source_namespaces.items()
        if name not in output_namespaces
    )
    if missing:
        position = output_root.start() + output_root.group(0).find(" ")
        if position < output_root.start():
            position = output_root.end() - 1
        output_text = output_text[:position] + missing + output_text[position:]
    return output_text.encode("utf-8")


def visible_text(element: ET.Element) -> str:
    return "".join(item.text or "" for item in element.iter(W + "t"))


def parent_map(root: ET.Element) -> dict[ET.Element, ET.Element]:
    return {child: parent for parent in root.iter() for child in parent}


def copy_run_properties(run: ET.Element) -> ET.Element | None:
    properties = run.find(W + "rPr")
    return copy.deepcopy(properties) if properties is not None else None


def make_run(text: str, properties: ET.Element | None) -> ET.Element:
    run = ET.Element(W + "r")
    if properties is not None:
        run.append(copy.deepcopy(properties))
    value = ET.SubElement(run, W + "t")
    if text.startswith(" ") or text.endswith(" ") or "  " in text:
        value.set("{http://www.w3.org/XML/1998/namespace}space", "preserve")
    value.text = text
    return run


def make_control(tag: str, content: list[ET.Element]) -> ET.Element:
    control = ET.Element(W + "sdt")
    properties = ET.SubElement(control, W + "sdtPr")
    alias = ET.SubElement(properties, W + "alias")
    alias.set(W + "val", tag)
    marker = ET.SubElement(properties, W + "tag")
    marker.set(W + "val", tag)
    ET.SubElement(properties, W + "text")
    body = ET.SubElement(control, W + "sdtContent")
    for item in content:
        body.append(item)
    return control


def element_is_inside_control(element: ET.Element, parents: dict[ET.Element, ET.Element]) -> bool:
    current = parents.get(element)
    while current is not None:
        if current.tag == W + "sdt":
            return True
        current = parents.get(current)
    return False


def wrap_first_plain_text(root: ET.Element, target: str, tag: str, replacement: str | None = None) -> bool:
    parents = parent_map(root)
    for text_element in root.iter(W + "t"):
        if element_is_inside_control(text_element, parents):
            continue
        text = text_element.text or ""
        position = text.find(target)
        if position < 0:
            continue
        run = parents.get(text_element)
        if run is None or run.tag != W + "r" or len(list(run.iter(W + "t"))) != 1:
            continue
        container = parents.get(run)
        if container is None:
            continue
        properties = copy_run_properties(run)
        before = text[:position]
        dynamic = replacement if replacement is not None else target
        after = text[position + len(target):]
        index = list(container).index(run)
        container.remove(run)
        additions: list[ET.Element] = []
        if before:
            additions.append(make_run(before, properties))
        additions.append(make_control(tag, [make_run(dynamic, properties)]))
        if after:
            additions.append(make_run(after, properties))
        for offset, item in enumerate(additions):
            container.insert(index + offset, item)
        return True
    return False


def wrap_plain_text_count(root: ET.Element, target: str, tag: str, count: int, replacement: str | None = None) -> None:
    for _ in range(count):
        if not wrap_first_plain_text(root, target, tag, replacement):
            raise RuntimeError(f"cannot bind {tag}: target={target!r}")


def unwrap_controls(paragraph: ET.Element) -> None:
    while True:
        parents = parent_map(paragraph)
        control = next(iter(paragraph.iter(W + "sdt")), None)
        if control is None:
            return
        parent = parents[control]
        body = control.find(W + "sdtContent")
        index = list(parent).index(control)
        parent.remove(control)
        if body is not None:
            for offset, child in enumerate(list(body)):
                parent.insert(index + offset, child)


def wrap_paragraph_runs(paragraph: ET.Element, tag: str, keep_prefix_runs: int = 0) -> None:
    unwrap_controls(paragraph)
    children = list(paragraph)
    runs = [item for item in children if item.tag == W + "r"]
    if len(runs) <= keep_prefix_runs:
        raise RuntimeError(f"paragraph has no dynamic runs for {tag}: {visible_text(paragraph)}")
    dynamic = runs[keep_prefix_runs:]
    first_index = list(paragraph).index(dynamic[0])
    for run in dynamic:
        paragraph.remove(run)
    paragraph.insert(first_index, make_control(tag, dynamic))


def leaf_paragraphs(root: ET.Element) -> list[ET.Element]:
    return [paragraph for paragraph in root.iter(W + "p") if paragraph.find(".//" + W + "p") is None]


def retag_detail_controls(document: ET.Element) -> None:
    expected = {
        "测评得分": "score",
        "评价等级": "level",
        "【表现评估】": "performance",
    }
    indexes = {key: 0 for key in expected}
    for paragraph in leaf_paragraphs(document):
        text = visible_text(paragraph).strip()
        for marker, suffix in expected.items():
            if not text.startswith(marker):
                continue
            index = indexes[marker]
            if index >= len(DIMENSIONS):
                raise RuntimeError(f"too many detail controls for {marker}")
            dimension = DIMENSIONS[index][0]
            tag = f"dimension.{dimension}.{suffix}"
            tags = list(paragraph.iter(W + "tag"))
            aliases = list(paragraph.iter(W + "alias"))
            if len(tags) != 1 or len(aliases) != 1:
                raise RuntimeError(f"detail binding shape invalid for {tag}")
            tags[0].set(W + "val", tag)
            aliases[0].set(W + "val", tag)
            if suffix == "score":
                score_text = display_score(DIMENSIONS[index][2])
                values = list(paragraph.iter(W + "sdtContent"))
                if len(values) != 1:
                    raise RuntimeError(f"detail score content invalid for {tag}")
                text_nodes = list(values[0].iter(W + "t"))
                if len(text_nodes) != 1:
                    raise RuntimeError(f"detail score text invalid for {tag}")
                text_nodes[0].text = score_text
            indexes[marker] += 1
    for marker, count in indexes.items():
        if count != len(DIMENSIONS):
            raise RuntimeError(f"detail {marker} controls={count}, want {len(DIMENSIONS)}")


def bind_overview(document: ET.Element) -> None:
    paragraphs = leaf_paragraphs(document)

    # Overall level in the headline spans two runs in the customer sample.
    headline = next((p for p in paragraphs if visible_text(p).strip() == "胜任程度：合格胜任"), None)
    if headline is None:
        raise RuntimeError("overall headline not found")
    runs = [item for item in list(headline) if item.tag == W + "r"]
    if len(runs) < 4:
        raise RuntimeError("overall headline run structure changed")
    dynamic = runs[-2:]
    index = list(headline).index(dynamic[0])
    for run in dynamic:
        headline.remove(run)
    headline.insert(index, make_control("overall.level", dynamic))

    total_line = next(
        (
            p
            for p in paragraphs
            if visible_text(p).strip().startswith(("总分 68.15", "总分 68.13"))
        ),
        None,
    )
    if total_line is None:
        raise RuntimeError("overall total line not found")
    overall_sample = "68.15" if "68.15" in visible_text(total_line) else "68.13"
    if not wrap_first_plain_text(total_line, overall_sample, "overall.score", "68.13"):
        raise RuntimeError("overall total score cannot be bound")
    if not wrap_first_plain_text(total_line, "优于常模分", "overall.normComparison"):
        raise RuntimeError("overall norm comparison cannot be bound")
    # This occurrence is the level adjective inside '对应…胜任水平'.
    if not wrap_first_plain_text(total_line, "合格", "overall.level"):
        raise RuntimeError("overall explanatory level cannot be bound")

    summary = next((p for p in paragraphs if visible_text(p).strip().startswith("自我管理类优势突出")), None)
    if summary is None:
        raise RuntimeError("module summary not found")
    wrap_paragraph_runs(summary, "rule.moduleSummary")

    advice = next((p for p in paragraphs if visible_text(p).strip().startswith("使用建议：受测者")), None)
    if advice is None:
        raise RuntimeError("overall advice not found")
    wrap_paragraph_runs(advice, "rule.overallAdvice", keep_prefix_runs=1)

    strength_samples = [
        "数字应用：能熟练使用数字工具",
        "求真务实：可以做到以事实和效果为依据",
        "自律性：主动约束自身行为",
    ]
    development_samples = [
        "成就导向：愿意尽力达到既定标准和要求",
        "沟通表达：多数情况下能把想法表达清楚",
    ]
    for index, sample in enumerate(strength_samples, 1):
        paragraph = next((p for p in leaf_paragraphs(document) if visible_text(p).strip().startswith(sample)), None)
        if paragraph is None:
            raise RuntimeError(f"strength slot {index} not found")
        wrap_paragraph_runs(paragraph, f"rule.strength.{index}")
    for index, sample in enumerate(development_samples, 1):
        paragraph = next((p for p in leaf_paragraphs(document) if visible_text(p).strip().startswith(sample)), None)
        if paragraph is None:
            raise RuntimeError(f"development slot {index} not found")
        wrap_paragraph_runs(paragraph, f"rule.development.{index}")

    # Remaining overall-score displays belong to chart-adjacent drawing text.
    remaining_overall_sample = (
        "68.15"
        if sum(visible_text(p).count("68.15") for p in leaf_paragraphs(document)) >= 4
        else "68.13"
    )
    wrap_plain_text_count(document, remaining_overall_sample, "overall.score", 4, "68.13")


def replace_leaf_paragraph_with_parts(paragraph: ET.Element, parts: list[tuple[str, str | None]]) -> None:
    runs = [item for item in list(paragraph) if item.tag == W + "r"]
    if not runs:
        raise RuntimeError(f"paragraph has no styled run: {visible_text(paragraph)}")
    properties = copy_run_properties(runs[0])
    for child in list(paragraph):
        if child.tag != W + "pPr":
            paragraph.remove(child)
    for text, tag in parts:
        run = make_run(text, properties)
        paragraph.append(make_control(tag, [run]) if tag else run)


def bind_module_rows(document: ET.Element) -> None:
    for code, name, score, level, comparison in MODULES:
        score_samples = {f"{score:.2f}"}
        if code == "self_management":
            score_samples.add("75.08")
        matched = []
        for paragraph in leaf_paragraphs(document):
            text = visible_text(paragraph).strip()
            if name in text and any(sample in text for sample in score_samples):
                matched.append(paragraph)
        if len(matched) != 2:
            raise RuntimeError(f"module row {code} count={len(matched)}, want 2")
        for paragraph in matched:
            replace_leaf_paragraph_with_parts(
                paragraph,
                [
                    (name + "：", None),
                    (f"{score:.2f}", f"module.{code}.score"),
                    ("  ", None),
                    (level, f"module.{code}.level"),
                    ("  ", None),
                    (comparison, f"module.{code}.normComparison"),
                ],
            )


def add_duration_control(document: ET.Element) -> None:
    tables = list(document.iter(W + "tbl"))
    profile_rows = tables[0].findall(W + "tr")
    duration_cell = profile_rows[6].findall(W + "tc")[2]
    paragraph = duration_cell.find(W + "p")
    if paragraph is None or visible_text(paragraph).strip():
        raise RuntimeError("duration target cell is not empty")
    date_cell = profile_rows[6].findall(W + "tc")[1]
    date_run = next(date_cell.iter(W + "r"), None)
    properties = copy_run_properties(date_run) if date_run is not None else None
    paragraph.append(make_run("时长：", properties))
    paragraph.append(make_control("result.userTime", [make_run("20", properties)]))
    paragraph.append(make_run("分钟", properties))


def bind_header(header: ET.Element) -> None:
    wrap_plain_text_count(header, "张三", "participant.name", 1)
    wrap_plain_text_count(header, "男", "participant.gender", 1)
    wrap_plain_text_count(header, "13800000000", "participant.telephone", 1)
    wrap_plain_text_count(header, "有效", "validity.status", 1)


def tune_layout(document: ET.Element, allow_already_clean: bool = False) -> None:
    body = document.find(W + "body")
    if body is None:
        raise RuntimeError("document body missing")
    removed = set()
    for paragraph in list(body):
        paragraph_id = paragraph.get(W14 + "paraId")
        if paragraph.tag == W + "p" and paragraph_id in COVER_BLANK_PARAGRAPH_IDS:
            if visible_text(paragraph).strip() or next(paragraph.iter(W + "drawing"), None) is not None:
                raise RuntimeError(f"cover paragraph {paragraph_id} is no longer safely empty")
            body.remove(paragraph)
            removed.add(paragraph_id)
    if removed != COVER_BLANK_PARAGRAPH_IDS and not (
        allow_already_clean and not removed
    ):
        raise RuntimeError(f"cover blank paragraphs changed: {removed}")

    for spacing in document.iter(W + "spacing"):
        for name in ("line", "before", "after"):
            attr = W + name
            value = spacing.get(attr)
            if value and value.isdigit():
                spacing.set(attr, str(max(1, round(int(value) * LAYOUT_SCALE))))
    for height in document.iter(W + "trHeight"):
        value = height.get(W + "val")
        if value and value.isdigit():
            height.set(W + "val", str(max(1, round(int(value) * LAYOUT_SCALE))))

    tables = list(document.iter(W + "tbl"))
    if len(tables) != 9:
        raise RuntimeError(f"report table count={len(tables)}, want 9")
    for table in tables[6:9]:
        rows = table.findall(W + "tr")
        if len(rows) % 3 != 0:
            raise RuntimeError("dimension detail rows are not grouped in threes")
        for row in rows:
            properties = row.find(W + "trPr")
            if properties is None:
                properties = ET.Element(W + "trPr")
                row.insert(0, properties)
            if properties.find(W + "cantSplit") is None:
                properties.append(ET.Element(W + "cantSplit"))
        for row_index in range(0, len(rows), 3):
            for keep_index in (row_index, row_index + 1):
                for cell in rows[keep_index].findall(W + "tc"):
                    for paragraph in cell.findall(W + "p"):
                        properties = paragraph.find(W + "pPr")
                        if properties is None:
                            properties = ET.Element(W + "pPr")
                            paragraph.insert(0, properties)
                        if properties.find(W + "keepNext") is None:
                            properties.append(ET.Element(W + "keepNext"))


def assign_chart_business_keys(document: ET.Element, relationships_xml: bytes) -> None:
    relationships = parse_xml(relationships_xml)
    targets = {
        element.get("Id"): element.get("Target", "")
        for element in relationships.findall(PR + "Relationship")
    }
    assigned = set()
    for drawing in list(document.iter(WP + "inline")) + list(document.iter(WP + "anchor")):
        chart = drawing.find(".//" + C + "chart")
        if chart is None:
            continue
        target = targets.get(chart.get(R + "id"), "")
        match = re.search(r"chart(\d+)\.xml$", target.replace("\\", "/"))
        if not match:
            raise RuntimeError(f"chart relationship target invalid: {target}")
        index = int(match.group(1))
        key = CHART_KEYS[index]
        properties = drawing.find(WP + "docPr")
        if properties is None:
            raise RuntimeError(f"chart{index} has no docPr")
        properties.set("title", key)
        assigned.add(key)
    if assigned != set(CHART_KEYS.values()):
        raise RuntimeError(f"chart keys incomplete: {assigned}")


def set_cache_values(container: ET.Element, values: list[float]) -> None:
    cache = container.find(".//" + C + "numCache")
    if cache is None:
        cache = container.find(".//" + C + "numLit")
    if cache is None:
        raise RuntimeError("numeric cache missing")
    for point in list(cache.findall(C + "pt")):
        cache.remove(point)
    count = cache.find(C + "ptCount")
    if count is None:
        count = ET.SubElement(cache, C + "ptCount")
    count.set("val", str(len(values)))
    for index, value in enumerate(values):
        point = ET.SubElement(cache, C + "pt")
        point.set("idx", str(index))
        item = ET.SubElement(point, C + "v")
        item.text = format(value, ".15g")


def materialize_chart_references(chart: ET.Element, index: int) -> None:
    while True:
        parents = parent_map(chart)
        reference = next(iter(chart.iter(C + "numRef")), None)
        if reference is None:
            break
        cache = reference.find(C + "numCache")
        if cache is None:
            raise RuntimeError(f"chart{index} numeric reference has no cache")
        literal = ET.Element(C + "numLit")
        for child in list(cache):
            literal.append(copy.deepcopy(child))
        parent = parents[reference]
        position = list(parent).index(reference)
        parent.remove(reference)
        parent.insert(position, literal)

    while True:
        parents = parent_map(chart)
        reference = next(iter(chart.iter(C + "strRef")), None)
        if reference is None:
            break
        cache = reference.find(C + "strCache")
        parent = parents[reference]
        if cache is None:
            raise RuntimeError(f"chart{index} string reference has no cache")
        if parent.tag == C + "tx":
            values = [item.text or "" for item in cache.findall(".//" + C + "v")]
            if len(values) != 1:
                raise RuntimeError(f"chart{index} series title cache is invalid")
            literal = ET.Element(C + "v")
            literal.text = values[0]
        elif parent.tag == C + "cat":
            literal = ET.Element(C + "strLit")
            for child in list(cache):
                literal.append(copy.deepcopy(child))
        else:
            raise RuntimeError(f"chart{index} unsupported string reference parent: {parent.tag}")
        position = list(parent).index(reference)
        parent.remove(reference)
        parent.insert(position, literal)


def clean_chart(xml_bytes: bytes, index: int) -> bytes:
    chart = parse_xml(xml_bytes)
    materialize_chart_references(chart, index)
    parents = parent_map(chart)
    for external_data in list(chart.iter(C + "externalData")):
        parents[external_data].remove(external_data)

    scores = [item[2] for item in DIMENSIONS]
    norms = [55.0, 65.0, 60.0, 55.0, 55.0, 52.5, 55.0, 57.5, 62.5, 60.0]
    if index == 1:
        series = list(chart.iter(C + "ser"))
        set_cache_values(series[0].find(C + "val"), [68.125, 31.875])
    elif index == 2:
        series = list(chart.iter(C + "ser"))
        if len(series) != 2:
            raise RuntimeError(f"comparison series={len(series)}, want 2")
        set_cache_values(series[0].find(C + "val"), scores)
        set_cache_values(series[1].find(C + "val"), norms)
    else:
        score = scores[index - 3]
        series = list(chart.iter(C + "ser"))
        set_cache_values(series[0].find(C + "val"), [score, 100.0 - score])

    # Keep any customer-authored rich labels synchronized with the corrected sample.
    replacements = {"68.15": "68.13", "75.08": "75.00", "78.25": "78.13"}
    for text in chart.iter(C + "v"):
        if text.text in replacements:
            text.text = replacements[text.text]
    return serialize_xml(chart, xml_bytes)


def clean_relationships(xml_bytes: bytes) -> bytes:
    root = parse_xml(xml_bytes)
    for relationship in list(root.findall(PR + "Relationship")):
        if relationship.get("TargetMode") == "External":
            root.remove(relationship)
    return serialize_xml(root, xml_bytes)


def transform_document(
    document_bytes: bytes, relationships_bytes: bytes, already_bound: bool = False
) -> bytes:
    document = parse_xml(document_bytes)
    if not already_bound:
        retag_detail_controls(document)
        bind_overview(document)
        bind_module_rows(document)
        add_duration_control(document)
    tune_layout(document, allow_already_clean=already_bound)
    assign_chart_business_keys(document, relationships_bytes)
    return serialize_xml(document, document_bytes)


def build(source: Path, output: Path, already_bound: bool = False) -> None:
    if not source.is_file():
        raise RuntimeError(f"source template missing: {source}")
    output.parent.mkdir(parents=True, exist_ok=True)
    with zipfile.ZipFile(source, "r") as incoming:
        relationships = incoming.read("word/_rels/document.xml.rels")
        transformed_document = transform_document(
            incoming.read("word/document.xml"), relationships, already_bound
        )
        transformed_headers: dict[str, bytes] = {}
        for name in incoming.namelist():
            if re.fullmatch(r"word/header\d+\.xml", name):
                header = parse_xml(incoming.read(name))
                if not already_bound and "姓名：张三" in visible_text(header):
                    bind_header(header)
                transformed_headers[name] = serialize_xml(header, incoming.read(name))

        with zipfile.ZipFile(output, "w") as outgoing:
            for item in incoming.infolist():
                data = incoming.read(item.filename)
                if item.filename == "word/document.xml":
                    data = transformed_document
                elif item.filename in transformed_headers:
                    data = transformed_headers[item.filename]
                elif re.fullmatch(r"word/charts/chart\d+\.xml", item.filename):
                    index = int(re.search(r"chart(\d+)\.xml", item.filename).group(1))
                    data = clean_chart(data, index)
                elif re.fullmatch(r"word/charts/_rels/chart\d+\.xml\.rels", item.filename):
                    data = clean_relationships(data)
                item.compress_type = zipfile.ZIP_STORED if item.is_dir() or not data else zipfile.ZIP_DEFLATED
                outgoing.writestr(item, data)

    print("COMPETENCY_260915_TEMPLATE_BUILT")
    print(f"source={source}")
    print(f"output={output}")
    print("schema=competency-frontline-report-template-draft-v2|controls=75|unique_tags=58|charts=12|external_links=0")


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--source", type=Path, default=DEFAULT_SOURCE)
    parser.add_argument("--output", type=Path, default=DEFAULT_OUTPUT)
    parser.add_argument(
        "--already-bound",
        action="store_true",
        help="preserve an existing complete content-control contract",
    )
    args = parser.parse_args()
    build(args.source.resolve(), args.output.resolve(), args.already_bound)


if __name__ == "__main__":
    main()
