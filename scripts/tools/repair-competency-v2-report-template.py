#!/usr/bin/env python3
"""FB-190: repair the active phase-1 v2 report template deterministically."""

from __future__ import annotations

import argparse
import re
import shutil
import struct
import subprocess
import tempfile
import zipfile
import xml.etree.ElementTree as ET
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
DEFAULT_TEMPLATE = ROOT / "Go-based Refactored System" / "configs" / "export-templates" / "competency-phase1-report-v2.docx"

DISCLAIMER_SAMPLE = (
    "本报告基于受测者在本次胜任力测评中的作答结果生成，用于辅助了解其当前胜任力表现及发展方向。"
    "测评结果可能受到作答状态、岗位经历和测评环境等因素影响，仅供人才发展、培训与管理决策参考，"
    "不应作为招聘、晋升、淘汰或其他重大人事决定的唯一依据。"
)
VALIDITY_SAMPLE = "本次测评作答效度良好，结果具有较好的参考价值。"
PAGE_BREAK_DIMENSIONS = {
    "digital_application": "数字应用",
    "continuous_learning": "持续学习",
    "cooperation": "合作意识",
    "self_discipline": "自律性",
}


def xml_escape(value: str) -> str:
    return value.replace("&", "&amp;").replace("<", "&lt;").replace(">", "&gt;")


def first_properties(paragraph: str, name: str) -> str:
    match = re.search(rf"<w:{name}\b.*?</w:{name}>", paragraph, re.S)
    return match.group(0) if match else ""


def content_control(tag: str, value: str, run_properties: str) -> str:
    return (
        '<w:sdt><w:sdtPr><w:alias w:val="{0}"/><w:tag w:val="{0}"/>'
        '<w:text/></w:sdtPr><w:sdtContent><w:r>{1}<w:t>{2}</w:t></w:r>'
        '</w:sdtContent></w:sdt>'
    ).format(tag, run_properties, xml_escape(value))


def replace_guidance_paragraph(document: str) -> str:
    if 'w:val="report.disclaimer"' in document:
        return document
    paragraph_pattern = re.compile(r"<w:p\b[^>]*>.*?</w:p>", re.S)
    matches = list(paragraph_pattern.finditer(document))
    target = next((match for match in matches if "3.本报告结果可应用于组织的招聘" in match.group(0)), None)
    if target is None:
        raise RuntimeError("disclaimer target paragraph not found")
    paragraph = target.group(0)
    opening = paragraph[: paragraph.find(">") + 1]
    properties = first_properties(paragraph, "pPr")
    run_properties = first_properties(paragraph, "rPr")
    replacement = (
        opening
        + properties
        + "<w:r>"
        + run_properties
        + "<w:t>免责声明：</w:t></w:r>"
        + content_control("report.disclaimer", DISCLAIMER_SAMPLE, run_properties)
        + "</w:p>"
    )
    return document[: target.start()] + replacement + document[target.end() :]


def insert_validity_paragraph(document: str) -> str:
    if 'w:val="validity.text"' in document:
        return document
    paragraph_pattern = re.compile(r"<w:p\b[^>]*>.*?</w:p>", re.S)
    target = next(
        (match for match in paragraph_pattern.finditer(document) if 'w:val="rule.overallAdvice"' in match.group(0)),
        None,
    )
    if target is None:
        raise RuntimeError("overall-advice paragraph not found")
    paragraph = target.group(0)
    opening = paragraph[: paragraph.find(">") + 1]
    properties = first_properties(paragraph, "pPr")
    run_properties = first_properties(paragraph, "rPr")
    validity = (
        opening
        + properties
        + "<w:r>"
        + run_properties
        + "<w:t>效度说明：</w:t></w:r>"
        + content_control("validity.text", VALIDITY_SAMPLE, run_properties)
        + "</w:p>"
    )
    return document[: target.end()] + validity + document[target.end() :]


def add_property(paragraph: str, property_xml: str) -> str:
    tag_match = re.match(r"<(w:[A-Za-z0-9]+)", property_xml)
    if tag_match is None:
        raise RuntimeError(f"invalid paragraph property: {property_xml}")
    paragraph = re.sub(rf"<{tag_match.group(1)}\b[^>]*/>", "", paragraph)
    if "<w:pPr>" in paragraph:
        return paragraph.replace("<w:pPr>", "<w:pPr>" + property_xml, 1)
    opening_end = paragraph.find(">") + 1
    return paragraph[:opening_end] + "<w:pPr>" + property_xml + "</w:pPr>" + paragraph[opening_end:]


def stabilize_detail_pages(document: str) -> str:
    if 'w14:paraId="FB190001"' in document:
        return document
    document = document.replace("<w:pageBreakBefore />", "").replace("<w:pageBreakBefore/>", "")
    table_pattern = re.compile(r"<w:tbl>.*?</w:tbl>", re.S)
    tables = list(table_pattern.finditer(document))
    targets: list[tuple[re.Match[str], int, list[int]]] = []
    for table in tables:
        value = table.group(0)
        if 'w:val="dimension.logical_reasoning.score"' in value:
            targets.append((table, 1, [6, 12]))
        elif 'w:val="dimension.communication.score"' in value:
            targets.append((table, 3, [3]))
        elif 'w:val="dimension.truth_pragmatism.score"' in value:
            targets.append((table, 4, [3]))
    if len(targets) != 3:
        raise RuntimeError(f"detail table set changed: {len(targets)}")

    for table, first_separator, cuts in sorted(targets, key=lambda item: item[0].start(), reverse=True):
        value = table.group(0)
        rows = list(re.finditer(r"<w:tr\b.*?</w:tr>", value, re.S))
        prefix = value[: rows[0].start()]
        suffix = value[rows[-1].end() :]
        boundaries = [0, *cuts, len(rows)]
        chunks = []
        for chunk_index in range(len(boundaries) - 1):
            chunk_rows = "".join(match.group(0) for match in rows[boundaries[chunk_index] : boundaries[chunk_index + 1]])
            chunks.append(prefix + chunk_rows + suffix)
        rebuilt = chunks[0]
        for chunk_index, chunk in enumerate(chunks[1:]):
            separator = first_separator + chunk_index
            rebuilt += (
                f'<w:p w14:paraId="FB19000{separator}"><w:pPr/><w:r><w:br w:type="page"/></w:r></w:p>'
                + chunk
            )
        document = document[: table.start()] + rebuilt + document[table.end() :]
    return document


def repair_profile_table(document: str) -> str:
    control_at = document.find('w:val="result.userTime"')
    if control_at < 0:
        raise RuntimeError("duration control not found")
    start = document.rfind("<w:tbl>", 0, control_at)
    end = document.find("</w:tbl>", control_at)
    if start < 0 or end < 0:
        raise RuntimeError("profile table not found")
    end += len("</w:tbl>")
    table = document[start:end]
    grid_pattern = re.compile(r"<w:tblGrid>.*?</w:tblGrid>", re.S)
    table, count = grid_pattern.subn(
        '<w:tblGrid><w:gridCol w:w="578"/><w:gridCol w:w="2800"/><w:gridCol w:w="1622"/></w:tblGrid>',
        table,
        count=1,
    )
    if count != 1:
        raise RuntimeError("profile table grid not found")
    table = table.replace('<w:tcW w:w="3382" w:type="dxa"/>', '<w:tcW w:w="4422" w:type="dxa"/>')
    table = table.replace('<w:tcW w:w="3432" w:type="dxa"/>', '<w:tcW w:w="4422" w:type="dxa"/>')
    row_pattern = re.compile(r"<w:tr\b.*?</w:tr>", re.S)
    duration_row = next((match for match in row_pattern.finditer(table) if 'w:val="result.userTime"' in match.group(0)), None)
    if duration_row is None:
        raise RuntimeError("duration row not found")
    row = duration_row.group(0)
    cells = list(re.finditer(r"<w:tc\b.*?</w:tc>", row, re.S))
    if len(cells) != 3:
        raise RuntimeError(f"duration row cell count={len(cells)}")
    values = [cells[0].group(0), cells[1].group(0), cells[2].group(0)]
    values[1] = re.sub(r'<w:tcW\b[^>]*/>', '<w:tcW w:w="2800" w:type="dxa"/>', values[1], count=1)
    values[2] = re.sub(r'<w:tcW\b[^>]*/>', '<w:tcW w:w="1622" w:type="dxa"/>', values[2], count=1)
    values[2] = re.sub(r'<w:gridSpan\b[^>]*/>', '', values[2], count=1)
    rebuilt = row[: cells[0].start()] + "".join(values) + row[cells[-1].end() :]
    table = table[: duration_row.start()] + rebuilt + table[duration_row.end() :]

    phone_row = next((match for match in row_pattern.finditer(table) if 'w:val="participant.telephone"' in match.group(0)), None)
    if phone_row is None:
        raise RuntimeError("telephone row not found")
    phone = phone_row.group(0)
    tag_at = phone.find('w:val="participant.telephone"')
    paragraph_start = phone.rfind("<w:p ", 0, tag_at)
    paragraph_end = phone.find("</w:p>", tag_at)
    if paragraph_start < 0 or paragraph_end < 0:
        raise RuntimeError("telephone paragraph not found")
    paragraph_end += len("</w:p>")
    paragraph = re.sub(r'<w:jc\b[^>]*/>', '', phone[paragraph_start:paragraph_end])
    paragraph = paragraph.replace('<w:pPr>', '<w:pPr><w:jc w:val="left"/>', 1)
    paragraph = re.sub(r'<w:sz w:val="21"\s*/>', '<w:sz w:val="18"/>', paragraph)
    paragraph = re.sub(r'<w:szCs w:val="21"\s*/>', '<w:szCs w:val="18"/>', paragraph)
    phone = phone[:paragraph_start] + paragraph + phone[paragraph_end:]
    table = table[: phone_row.start()] + phone + table[phone_row.end() :]
    return document[:start] + table + document[end:]


def horizontal_scale_png() -> bytes:
    svg = """<svg xmlns="http://www.w3.org/2000/svg" width="500" height="100" viewBox="0 0 500 100">
<rect width="500" height="100" fill="white"/>
<g font-family="Microsoft YaHei, Noto Sans CJK SC, sans-serif" text-anchor="middle">
<rect x="0" y="0" width="100" height="70" fill="#00B050"/><rect x="100" y="0" width="100" height="70" fill="#70AD47"/>
<rect x="200" y="0" width="100" height="70" fill="#A9D18E"/><rect x="300" y="0" width="100" height="70" fill="#F4B183"/>
<rect x="400" y="0" width="100" height="70" fill="#F28E2B"/>
<g fill="white" font-size="24" font-weight="700"><text x="50" y="43">优秀</text><text x="150" y="43">良好</text><text x="250" y="43">合格</text><text x="350" y="43">薄弱</text><text x="450" y="43">不足</text></g>
<g fill="#555" font-size="16"><text x="50" y="92">90–100</text><text x="150" y="92">70–89.99</text><text x="250" y="92">30–69.99</text><text x="350" y="92">10–29.99</text><text x="450" y="92">0–9.99</text></g>
</g></svg>"""
    command = shutil.which("magick") or shutil.which("convert")
    if not command:
        raise RuntimeError("ImageMagick is required to build the grade scale")
    completed = subprocess.run([command, "svg:-", "-strip", "png:-"], input=svg.encode("utf-8"), capture_output=True, check=True)
    png = completed.stdout
    if png[:8] != b"\x89PNG\r\n\x1a\n":
        raise RuntimeError("grade scale renderer did not return PNG")
    width, height = struct.unpack(">II", png[16:24])
    if (width, height) != (500, 100):
        raise RuntimeError(f"grade scale size={width}x{height}")
    return png


def vertical_scale_png() -> bytes:
    svg = """<svg xmlns="http://www.w3.org/2000/svg" width="180" height="650" viewBox="0 0 180 650">
<rect width="180" height="650" rx="10" fill="#F7FAF8"/>
<g font-family="Microsoft YaHei, Noto Sans CJK SC, sans-serif" text-anchor="middle">
<rect x="8" y="8" width="164" height="58" rx="8" fill="#00A651"/>
<rect x="8" y="70" width="164" height="116" rx="8" fill="#38B86A"/>
<rect x="8" y="190" width="164" height="232" rx="8" fill="#A8D889"/>
<rect x="8" y="426" width="164" height="116" rx="8" fill="#F2A45F"/>
<rect x="8" y="546" width="164" height="58" rx="8" fill="#E88937"/>
<rect x="8" y="608" width="164" height="34" rx="8" fill="#5F6668"/>
<g fill="white" font-size="24" font-weight="700">
<text x="90" y="45">优秀</text><text x="90" y="124">良好</text><text x="90" y="300">合格</text>
<text x="90" y="487">薄弱</text><text x="90" y="583">不足</text><text x="90" y="634" font-size="18">等级</text>
</g>
<g fill="#FFFFFF" opacity="0.9" font-size="14">
<text x="90" y="160">70–89.99</text><text x="90" y="390">30–69.99</text><text x="90" y="520">10–29.99</text>
</g>
</g></svg>"""
    command = shutil.which("magick") or shutil.which("convert")
    if not command:
        raise RuntimeError("ImageMagick is required to build the vertical grade scale")
    completed = subprocess.run([command, "svg:-", "-strip", "png:-"], input=svg.encode("utf-8"), capture_output=True, check=True)
    png = completed.stdout
    width, height = struct.unpack(">II", png[16:24])
    if (width, height) != (180, 650):
        raise RuntimeError(f"vertical grade scale size={width}x{height}")
    return png


def resize_scale_drawing(document: str, relationships: bytes) -> str:
    relationship_root = ET.fromstring(relationships)
    relationship_namespace = "{http://schemas.openxmlformats.org/package/2006/relationships}"
    relation_id = next(
        (
            item.get("Id", "")
            for item in relationship_root.findall(relationship_namespace + "Relationship")
            if item.get("Target", "").replace("\\", "/").endswith("media/image17.png")
        ),
        "",
    )
    if not relation_id:
        raise RuntimeError("grade scale image relationship not found")
    marker = f'r:embed="{relation_id}"'
    at = document.find(marker)
    if at < 0:
        raise RuntimeError("grade scale drawing not found")
    start = document.rfind("<pic:pic>", 0, at)
    end = document.find("</pic:pic>", at)
    if start < 0 or end < 0:
        raise RuntimeError("grade scale picture invalid")
    end += len("</pic:pic>")
    picture = document[start:end]
    if '<a:off x="5796" y="56300"/><a:ext cx="3500" cy="700"/>' not in picture:
        picture, count = re.subn(
            r'<a:off x="5796" y="(?:54639|56450)"\s*/><a:ext cx="(?:580|2200)" cy="(?:3637|440)"\s*/>',
            '<a:off x="5796" y="56300"/><a:ext cx="3500" cy="700"/>',
            picture,
            count=1,
        )
        if count != 1:
            raise RuntimeError("grade scale picture transform changed")
    document = document[:start] + picture + document[end:]
    document = re.sub(
        r'<wpg:xfrm><a:off x="(?:6380|8050)" y="54476"\s*/><a:ext cx="(?:8932|7262)" cy="4388"\s*/></wpg:xfrm>',
        '<wpg:xfrm><a:off x="9300" y="54476"/><a:ext cx="6012" cy="4388"/></wpg:xfrm>',
        document,
        count=1,
    )
    document = re.sub(
        r'style="position:absolute;left:5796;top:(?:54639|56450|56300);height:(?:3637|440|700);width:(?:580|2200|3500);"',
        'style="position:absolute;left:5796;top:56300;height:700;width:3500;"',
        document,
        count=1,
    )
    document = re.sub(
        r'style="position:absolute;left:(?:6373|8043);top:54469;height:4403;width:(?:8947|7277);"',
        'style="position:absolute;left:9293;top:54469;height:4403;width:6027;"',
        document,
        count=1,
    )
    return document


def ungroup_comparison_chart(document: str, relationships: bytes) -> str:
    """Replace the Word-only wpg group with two ordinary inline drawings."""
    if 'title="chart.dimension.comparison.scale"' in document:
        return document
    relationship_root = ET.fromstring(relationships)
    relationship_namespace = "{http://schemas.openxmlformats.org/package/2006/relationships}"
    targets = {
        item.get("Id", ""): item.get("Target", "").replace("\\", "/")
        for item in relationship_root.findall(relationship_namespace + "Relationship")
    }
    alternate_pattern = re.compile(r"<mc:AlternateContent>.*?</mc:AlternateContent>", re.S)
    target = next(
        (
            match
            for match in alternate_pattern.finditer(document)
            if 'title="chart.dimension.comparison"' in match.group(0)
        ),
        None,
    )
    if target is None:
        raise RuntimeError("comparison chart group not found")
    group = target.group(0)
    scale_relation = next(
        (
            relation_id
            for relation_id, part in targets.items()
            if part.endswith("media/image17.png") and f'r:embed="{relation_id}"' in group
        ),
        "",
    )
    chart_relation = next(
        (
            relation_id
            for relation_id, part in targets.items()
            if part.endswith("charts/chart2.xml") and f'r:id="{relation_id}"' in group
        ),
        "",
    )
    if not scale_relation or not chart_relation:
        raise RuntimeError("comparison chart relationships not found")
    picture_match = re.search(
        rf"<pic:pic>.*?r:embed=\"{re.escape(scale_relation)}\".*?</pic:pic>",
        group,
        re.S,
    )
    if picture_match is None:
        raise RuntimeError("comparison grade-scale picture not found")
    picture = picture_match.group(0)
    picture = re.sub(
        r"<a:xfrm><a:off\b[^>]*/><a:ext\b[^>]*/></a:xfrm>",
        '<a:xfrm><a:off x="0" y="0"/><a:ext cx="368223" cy="2309844"/></a:xfrm>',
        picture,
        count=1,
    )
    existing_ids = [int(value) for value in re.findall(r"<wp:docPr\b[^>]*\bid=\"(\d+)\"", document)]
    scale_id = max(existing_ids, default=0) + 1
    chart_id = scale_id + 1
    scale_drawing = (
        '<w:drawing><wp:inline distT="0" distB="0" distL="0" distR="0">'
        '<wp:extent cx="368223" cy="2309844"/><wp:effectExtent l="0" t="0" r="0" b="0"/>'
        f'<wp:docPr id="{scale_id}" name="等级刻度" title="chart.dimension.comparison.scale"/>'
        '<wp:cNvGraphicFramePr/><a:graphic><a:graphicData '
        'uri="http://schemas.openxmlformats.org/drawingml/2006/picture">'
        f'{picture}</a:graphicData></a:graphic></wp:inline></w:drawing>'
    )
    chart_drawing = (
        '<w:drawing><wp:inline distT="0" distB="0" distL="0" distR="0">'
        '<wp:extent cx="5671071" cy="2786380"/><wp:effectExtent l="0" t="4445" r="15240" b="9525"/>'
        f'<wp:docPr id="{chart_id}" name="十维得分与常模" title="chart.dimension.comparison"/>'
        '<wp:cNvGraphicFramePr/><a:graphic><a:graphicData '
        'uri="http://schemas.openxmlformats.org/drawingml/2006/chart">'
        f'<c:chart r:id="{chart_relation}"/></a:graphicData></a:graphic></wp:inline></w:drawing>'
    )
    replacement = scale_drawing + chart_drawing
    return document[: target.start()] + replacement + document[target.end() :]


def beautify_overview(document: str) -> str:
    summary_at = document.find('w:val="rule.moduleSummary"')
    if summary_at < 0:
        raise RuntimeError("overview summary card not found")
    table_start = document.rfind("<w:tbl>", 0, summary_at)
    table_end = document.find("</w:tbl>", summary_at)
    if table_start < 0 or table_end < 0:
        raise RuntimeError("overview summary table invalid")
    table_end += len("</w:tbl>")
    table = document[table_start:table_end]
    cell_properties = re.search(r"<w:tcPr>.*?</w:tcPr>", table, re.S)
    if cell_properties is None:
        raise RuntimeError("overview summary cell properties missing")
    properties = cell_properties.group(0)
    properties = re.sub(r"<w:tcBorders>.*?</w:tcBorders>", "", properties, flags=re.S)
    properties = re.sub(r"<w:shd\b[^>]*/>", "", properties)
    properties = re.sub(r"<w:tcMar>.*?</w:tcMar>", "", properties, flags=re.S)
    decoration = (
        '<w:tcBorders><w:top w:val="single" w:sz="6" w:space="0" w:color="BFE3CC"/>'
        '<w:left w:val="single" w:sz="28" w:space="0" w:color="00B050"/>'
        '<w:bottom w:val="single" w:sz="6" w:space="0" w:color="BFE3CC"/>'
        '<w:right w:val="single" w:sz="6" w:space="0" w:color="BFE3CC"/></w:tcBorders>'
        '<w:shd w:val="clear" w:color="auto" w:fill="F1F8F4"/>'
        '<w:tcMar><w:top w:w="120" w:type="dxa"/><w:left w:w="180" w:type="dxa"/>'
        '<w:bottom w:w="120" w:type="dxa"/><w:right w:w="180" w:type="dxa"/></w:tcMar>'
    )
    properties = properties.replace("</w:tcPr>", decoration + "</w:tcPr>")
    table = table[: cell_properties.start()] + properties + table[cell_properties.end() :]
    document = document[:table_start] + table + document[table_end:]

    alternate_pattern = re.compile(r"<mc:AlternateContent>.*?</mc:AlternateContent>", re.S)
    center = next(
        (
            match
            for match in alternate_pattern.finditer(document)
            if 'w:val="overall.score"' in match.group(0) and "总体评价" in match.group(0)
        ),
        None,
    )
    if center is None:
        raise RuntimeError("overview doughnut centre label not found")
    block = center.group(0)
    paragraph_pattern = re.compile(r"<w:p\b.*?</w:p>", re.S)

    def tune_center_paragraph(match: re.Match[str]) -> str:
        paragraph = match.group(0)
        if "总体评价" in paragraph:
            paragraph = paragraph.replace("总体评价", "总体得分")
            paragraph = re.sub(r'<w:spacing w:line="\d+" w:lineRule="exact"\s*/>', '<w:spacing w:line="180" w:lineRule="exact"/>', paragraph)
            paragraph = re.sub(r'<w:color w:val="[0-9A-Fa-f]+"(?:\s+w:themeColor="[^"]+")?\s*/>', '<w:color w:val="666666"/>', paragraph)
            paragraph = re.sub(r'<w:sz(?:Cs)? w:val="(?:28|36)"\s*/>', lambda value: '<w:szCs w:val="18"/>' if "szCs" in value.group(0) else '<w:sz w:val="18"/>', paragraph)
        elif 'w:val="overall.score"' in paragraph:
            paragraph = re.sub(r'<w:spacing w:line="\d+" w:lineRule="exact"\s*/>', '<w:spacing w:line="340" w:lineRule="exact"/>', paragraph)
            paragraph = re.sub(r'<w:color w:val="[0-9A-Fa-f]+"(?:\s+w:themeColor="[^"]+")?\s*/>', '<w:color w:val="00B050"/>', paragraph)
            paragraph = re.sub(r'<w:sz(?:Cs)? w:val="(?:28|36)"\s*/>', lambda value: '<w:szCs w:val="32"/>' if "szCs" in value.group(0) else '<w:sz w:val="32"/>', paragraph)
        return paragraph

    block = paragraph_pattern.sub(tune_center_paragraph, block)
    document = document[: center.start()] + block + document[center.end() :]

    replacements = {
        '<wp:extent cx="368223" cy="2309844"/>': '<wp:extent cx="650000" cy="2309844"/>',
        '<a:ext cx="368223" cy="2309844"/>': '<a:ext cx="650000" cy="2309844"/>',
        '<wp:extent cx="5671071" cy="2786380"/>': '<wp:extent cx="5389294" cy="2786380"/>',
    }
    for old, new in replacements.items():
        if old not in document:
            raise RuntimeError(f"overview drawing geometry changed: {old}")
        document = document.replace(old, new, 1)
    return document


def repair(source: Path, output: Path) -> None:
    with zipfile.ZipFile(source) as archive:
        parts = {name: archive.read(name) for name in archive.namelist()}
        order = archive.namelist()
    document = parts["word/document.xml"].decode("utf-8")
    document = repair_profile_table(document)
    document = replace_guidance_paragraph(document)
    document = insert_validity_paragraph(document)
    document = stabilize_detail_pages(document)
    document = ungroup_comparison_chart(document, parts["word/_rels/document.xml.rels"])
    document = beautify_overview(document)
    parts["word/document.xml"] = document.encode("utf-8")
    parts["word/media/image17.png"] = vertical_scale_png()

    output.parent.mkdir(parents=True, exist_ok=True)
    with tempfile.NamedTemporaryFile(dir=output.parent, suffix=".docx", delete=False) as temporary:
        temporary_path = Path(temporary.name)
    try:
        with zipfile.ZipFile(temporary_path, "w", compression=zipfile.ZIP_DEFLATED) as archive:
            for name in order:
                info = zipfile.ZipInfo(name, date_time=(1980, 1, 1, 0, 0, 0))
                info.compress_type = zipfile.ZIP_STORED if name.endswith("/") or not parts[name] else zipfile.ZIP_DEFLATED
                archive.writestr(info, parts[name])
        temporary_path.replace(output)
    finally:
        temporary_path.unlink(missing_ok=True)
    print(f"COMPETENCY_V2_REPORT_TEMPLATE_REPAIRED output={output}")


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--source", type=Path, default=DEFAULT_TEMPLATE)
    parser.add_argument("--output", type=Path, default=DEFAULT_TEMPLATE)
    args = parser.parse_args()
    repair(args.source.resolve(), args.output.resolve())


if __name__ == "__main__":
    main()
