#!/usr/bin/env python3
"""Repair the customer-authored 00401 v2 template without adding optional fields."""

from __future__ import annotations

import argparse
import re
import tempfile
import zipfile
import xml.etree.ElementTree as ET
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
DEFAULT_SOURCE = ROOT / "docs" / "261010胜任力待完善" / "最新胜任力报告模板-剔除时长-候选.docx"
DEFAULT_OUTPUT = ROOT / "docs" / "261010胜任力待完善" / "最新胜任力报告模板-修复后-staging候选.docx"
REL_NS = "{http://schemas.openxmlformats.org/package/2006/relationships}"


def materialize_chart_references(body: bytes) -> bytes:
    content = body.decode("utf-8")

    def numeric_literal(match: re.Match[str]) -> str:
        cache = re.search(r"<c:numCache>(.*?)</c:numCache>", match.group(0), re.S)
        if cache is None:
            raise RuntimeError("numeric chart reference has no cache")
        return "<c:numLit>" + cache.group(1) + "</c:numLit>"

    content = re.sub(r"<c:numRef>.*?</c:numRef>", numeric_literal, content, flags=re.S)

    def title_literal(match: re.Match[str]) -> str:
        cache = re.search(r"<c:strCache>(.*?)</c:strCache>", match.group(0), re.S)
        if cache is None:
            raise RuntimeError("chart series title cache invalid")
        values = re.findall(r"<c:v>(.*?)</c:v>", cache.group(1), re.S)
        if len(values) != 1:
            raise RuntimeError("chart series title cache invalid")
        return "<c:tx><c:v>" + values[0] + "</c:v></c:tx>"

    content = re.sub(r"<c:tx><c:strRef>.*?</c:strRef></c:tx>", title_literal, content, flags=re.S)

    def category_literal(match: re.Match[str]) -> str:
        cache = re.search(r"<c:strCache>(.*?)</c:strCache>", match.group(0), re.S)
        if cache is None:
            raise RuntimeError("chart category cache invalid")
        return "<c:cat><c:strLit>" + cache.group(1) + "</c:strLit></c:cat>"

    content = re.sub(r"<c:cat><c:strRef>.*?</c:strRef></c:cat>", category_literal, content, flags=re.S)
    content = re.sub(r"<c:externalData\b[^>]*>.*?</c:externalData>", "", content, flags=re.S)
    if any(token in content for token in ("<c:numRef", "<c:strRef", "<c:externalData", "<c:f>")):
        raise RuntimeError("chart reference materialization incomplete")
    return content.encode("utf-8")


def remove_external_chart_artifacts(name: str, body: bytes) -> tuple[bytes, int]:
    removed = 0
    if name.startswith("word/charts/_rels/") and name.endswith(".rels"):
        removed = len(re.findall(rb"<Relationship\b[^>]*\bTargetMode=\"External\"[^>]*/>", body))
        body = re.sub(rb"<Relationship\b[^>]*\bTargetMode=\"External\"[^>]*/>", b"", body)
    elif name.startswith("word/charts/chart") and name.endswith(".xml"):
        removed = len(re.findall(rb"<c:externalData\b", body))
        body = materialize_chart_references(body)
    return body, removed


def ungroup_comparison_chart(document: str, relationships: bytes) -> str:
    alternate_pattern = re.compile(r"<mc:AlternateContent>.*?</mc:AlternateContent>", re.S)
    target = next(
        (match for match in alternate_pattern.finditer(document) if 'title="chart.dimension.comparison"' in match.group(0)),
        None,
    )
    if target is None:
        if 'title="chart.dimension.comparison.scale"' in document:
            return document
        raise RuntimeError("comparison chart group not found")
    group = target.group(0)
    relationship_root = ET.fromstring(relationships)
    targets = {
        item.get("Id", ""): item.get("Target", "").replace("\\", "/")
        for item in relationship_root.findall(REL_NS + "Relationship")
    }
    chart_id_match = re.search(r'<c:chart\b[^>]*\br:id="([^"]+)"', group)
    if chart_id_match is None:
        raise RuntimeError("comparison chart relationship id not found")
    chart_relation = chart_id_match.group(1)
    if not targets.get(chart_relation, "").startswith("charts/chart"):
        raise RuntimeError("comparison chart relationship target invalid")
    scale_id_match = re.search(r'<a:blip\b[^>]*\br:embed="([^"]+)"', group)
    if scale_id_match is None:
        raise RuntimeError("comparison scale relationship id not found")
    scale_relation = scale_id_match.group(1)
    if not targets.get(scale_relation, "").startswith("media/"):
        raise RuntimeError("comparison scale relationship target invalid")
    picture_match = re.search(
        rf"<pic:pic\b[^>]*>.*?r:embed=\"{re.escape(scale_relation)}\".*?</pic:pic>", group, re.S
    )
    if picture_match is None:
        raise RuntimeError("comparison scale picture not found")
    picture = re.sub(
        r"<a:xfrm><a:off\b[^>]*/><a:ext\b[^>]*/></a:xfrm>",
        '<a:xfrm><a:off x="0" y="0"/><a:ext cx="368223" cy="2309844"/></a:xfrm>',
        picture_match.group(0),
        count=1,
    )
    existing_ids = [int(value) for value in re.findall(r'<wp:docPr\b[^>]*\bid="(\d+)"', document)]
    scale_id = max(existing_ids, default=0) + 1
    chart_id = scale_id + 1
    scale_drawing = (
        '<w:drawing><wp:inline distT="0" distB="0" distL="0" distR="0">'
        '<wp:extent cx="368223" cy="2309844"/><wp:effectExtent l="0" t="0" r="0" b="0"/>'
        f'<wp:docPr id="{scale_id}" name="等级刻度" title="chart.dimension.comparison.scale"/>'
        '<wp:cNvGraphicFramePr/><a:graphic xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" '
        'xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><a:graphicData '
        'uri="http://schemas.openxmlformats.org/drawingml/2006/picture">'
        f"{picture}</a:graphicData></a:graphic></wp:inline></w:drawing>"
    )
    chart_drawing = (
        '<w:drawing><wp:inline distT="0" distB="0" distL="0" distR="0">'
        '<wp:extent cx="5671071" cy="2786380"/><wp:effectExtent l="0" t="4445" r="15240" b="9525"/>'
        f'<wp:docPr id="{chart_id}" name="十维得分与常模" title="chart.dimension.comparison"/>'
        '<wp:cNvGraphicFramePr/><a:graphic xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" '
        'xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><a:graphicData '
        'uri="http://schemas.openxmlformats.org/drawingml/2006/chart">'
        f'<c:chart xmlns:c="http://schemas.openxmlformats.org/drawingml/2006/chart" r:id="{chart_relation}"/>'
        '</a:graphicData></a:graphic></wp:inline></w:drawing>'
    )
    return document[: target.start()] + scale_drawing + chart_drawing + document[target.end() :]


def ungroup_overall_chart(document: str, relationships: bytes) -> str:
    alternate_pattern = re.compile(r"<mc:AlternateContent>.*?</mc:AlternateContent>", re.S)
    target = next(
        (match for match in alternate_pattern.finditer(document) if 'title="chart.overall.score"' in match.group(0)),
        None,
    )
    if target is None:
        if re.search(r'<wp:docPr\b[^>]*title="chart\.overall\.score"', document):
            return document
        raise RuntimeError("overall score chart group not found")
    group = target.group(0)
    relationship_root = ET.fromstring(relationships)
    targets = {
        item.get("Id", ""): item.get("Target", "").replace("\\", "/")
        for item in relationship_root.findall(REL_NS + "Relationship")
    }
    chart_id_match = re.search(r'<c:chart\b[^>]*\br:id="([^"]+)"', group)
    if chart_id_match is None:
        raise RuntimeError("overall score chart relationship id not found")
    chart_relation = chart_id_match.group(1)
    if not targets.get(chart_relation, "").startswith("charts/chart"):
        raise RuntimeError("overall score chart relationship target invalid")
    extent_match = re.search(r'<wp:extent\b[^>]*\bcx="(\d+)"[^>]*\bcy="(\d+)"', group)
    if extent_match is None:
        raise RuntimeError("overall score chart extent missing")
    chart_width, chart_height = extent_match.groups()
    existing_ids = [int(value) for value in re.findall(r'<wp:docPr\b[^>]*\bid="(\d+)"', document)]
    chart_id = max(existing_ids, default=0) + 1
    paragraph_starts = list(re.finditer(r"<w:p(?:\s|>)", document[: target.start()]))
    paragraph_start = paragraph_starts[-1].start() if paragraph_starts else -1
    paragraph_end = document.find("</w:p>", target.end())
    if paragraph_start < 0 or paragraph_end < 0:
        raise RuntimeError("overall score chart paragraph invalid")
    paragraph_end += len("</w:p>")
    chart_paragraph = (
        '<w:p><w:pPr><w:jc w:val="center"/></w:pPr><w:r><w:drawing>'
        '<wp:inline distT="0" distB="0" distL="0" distR="0">'
        f'<wp:extent cx="{chart_width}" cy="{chart_height}"/>'
        '<wp:effectExtent l="0" t="0" r="0" b="0"/>'
        f'<wp:docPr id="{chart_id}" name="总体评价环图" title="chart.overall.score"/>'
        '<wp:cNvGraphicFramePr/><a:graphic xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" '
        'xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><a:graphicData '
        'uri="http://schemas.openxmlformats.org/drawingml/2006/chart">'
        f'<c:chart xmlns:c="http://schemas.openxmlformats.org/drawingml/2006/chart" r:id="{chart_relation}"/>'
        '</a:graphicData></a:graphic></wp:inline></w:drawing></w:r>'
        '<w:r><w:drawing><wp:anchor distT="0" distB="0" distL="0" distR="0" simplePos="0" '
        'relativeHeight="251659000" behindDoc="0" locked="0" layoutInCell="1" allowOverlap="1">'
        '<wp:simplePos x="0" y="0"/><wp:positionH relativeFrom="column"><wp:align>center</wp:align></wp:positionH>'
        '<wp:positionV relativeFrom="paragraph"><wp:posOffset>420000</wp:posOffset></wp:positionV>'
        f'<wp:extent cx="{chart_width}" cy="760000"/><wp:effectExtent l="0" t="0" r="0" b="0"/><wp:wrapNone/>'
        f'<wp:docPr id="{chart_id + 1}" name="总体评价中心文字" title="chart.overall.score.center"/>'
        '<wp:cNvGraphicFramePr/><a:graphic xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main">'
        '<a:graphicData uri="http://schemas.microsoft.com/office/word/2010/wordprocessingShape">'
        '<wps:wsp><wps:cNvSpPr txBox="1"/><wps:spPr><a:xfrm><a:off x="0" y="0"/>'
        f'<a:ext cx="{chart_width}" cy="760000"/></a:xfrm><a:prstGeom prst="rect"><a:avLst/></a:prstGeom>'
        '<a:noFill/><a:ln><a:noFill/></a:ln></wps:spPr><wps:txbx><w:txbxContent>'
        '<w:p><w:pPr><w:jc w:val="center"/></w:pPr><w:r><w:rPr><w:rFonts w:ascii="微软雅黑" w:hAnsi="微软雅黑" '
        'w:eastAsia="微软雅黑"/><w:sz w:val="20"/></w:rPr><w:t>总体评价</w:t></w:r></w:p>'
        '<w:p><w:pPr><w:jc w:val="center"/></w:pPr><w:sdt><w:sdtPr><w:alias w:val="overall.score"/>'
        '<w:tag w:val="overall.score"/><w:text/></w:sdtPr><w:sdtContent><w:r><w:rPr><w:rFonts w:ascii="微软雅黑" '
        'w:hAnsi="微软雅黑" w:eastAsia="微软雅黑"/><w:b/><w:sz w:val="30"/></w:rPr><w:t>68.13</w:t></w:r>'
        '</w:sdtContent></w:sdt><w:r><w:rPr><w:rFonts w:ascii="微软雅黑" w:hAnsi="微软雅黑" w:eastAsia="微软雅黑"/>'
        '<w:sz w:val="20"/></w:rPr><w:t>分</w:t></w:r></w:p>'
        '</w:txbxContent></wps:txbx><wps:bodyPr rot="0" vert="horz" wrap="none" lIns="0" tIns="0" rIns="0" '
        'bIns="0" anchor="ctr" anchorCtr="1"><a:spAutoFit/></wps:bodyPr></wps:wsp>'
        '</a:graphicData></a:graphic></wp:anchor></w:drawing></w:r></w:p>'
    )
    return document[:paragraph_start] + chart_paragraph + document[paragraph_end:]


def normalize_page_number_sections(document: str) -> str:
    def normalize_section(match: re.Match[str]) -> str:
        section = match.group(0)
        if 'w:type="default"' not in section or '<w:footerReference' not in section:
            return section
        section = re.sub(r'<w:footerReference\b[^>]*\bw:type="first"[^>]*/>', "", section)
        section = re.sub(r"<w:titlePg(?:\s[^>]*)?/>", "", section)
        return section

    return re.sub(r"<w:sectPr>.*?</w:sectPr>", normalize_section, document, flags=re.S)


def last_run_start(content: str, before: int) -> int:
    return max(content.rfind("<w:r", 0, before), content.rfind("<W:R", 0, before))


def normalize_footer(footer: bytes, simple_layout: bool = True) -> bytes:
    content = footer.decode("utf-8")

    def remove_total_page_fallback(match: re.Match[str]) -> str:
        value = match.group(0)
        return "<mc:Fallback/>" if "NUMPAGES" in value.upper() or "页 共" in value else value

    content = re.sub(r"<mc:Fallback>.*?</mc:Fallback>", remove_total_page_fallback, content, flags=re.S)

    def remove_field_at(value: str, instruction: int) -> str:
        upper_value = value.upper()
        begin_marker = upper_value.rfind('W:FLDCHARTYPE="BEGIN"', 0, instruction)
        field_start = last_run_start(upper_value, begin_marker)
        end_offset = upper_value.find('W:FLDCHARTYPE="END"', instruction)
        if begin_marker < 0 or field_start < 0 or end_offset < 0:
            raise RuntimeError("page field structure invalid")
        run_end = upper_value.find("</W:R>", end_offset)
        if run_end < 0:
            raise RuntimeError("page field end invalid")
        field_end = run_end + len("</w:r>")
        remove_start = field_start
        previous_end = upper_value.rfind("</W:R>", 0, field_start)
        if previous_end >= 0:
            previous_start = last_run_start(upper_value, previous_end)
            if previous_start >= 0:
                previous = value[previous_start : previous_end + len("</w:r>")]
                texts = re.findall(r"<w:t(?:\s[^>]*)?>(.*?)</w:t>", previous, re.S)
                if len(texts) == 1 and (texts[0].strip() == "/" or "共" in texts[0]):
                    remove_start = previous_start
        return value[:remove_start] + value[field_end:]

    upper = content.upper()
    while "NUMPAGES" in upper:
        instruction = upper.index("NUMPAGES")
        content = remove_field_at(content, instruction)
        upper = content.upper()
    page_pattern = re.compile(
        r"<w:instrText(?:\s[^>]*)?>\s*PAGE(?:\s+\\\*\s+MERGEFORMAT)?\s*</w:instrText>", re.I
    )
    while "页 共" in content and len(list(page_pattern.finditer(content))) > 1:
        content = remove_field_at(content, list(page_pattern.finditer(content))[1].start())
    content = re.sub(
        r'<w:r><w:t xml:space="preserve">\s*页\s+共\s*</w:t></w:r>'
        r'<w:r(?:\s[^>]*)?>.*?<w:t(?:\s[^>]*)?>\s*\d+\s*</w:t></w:r>',
        "",
        content,
        flags=re.S,
    )
    if not re.search(r"<w:instrText(?:\s[^>]*)?>\s*PAGE(?:\s+\\\*\s+MERGEFORMAT)?\s*</w:instrText>", content, re.I):
        raise RuntimeError("PAGE field missing")
    if not simple_layout:
        return content.encode("utf-8")
    namespace_match = re.search(r"<w:ftr\b([^>]*)>", content)
    if namespace_match is None:
        raise RuntimeError("footer root invalid")
    return (
        '<?xml version="1.0" encoding="UTF-8" standalone="yes"?>'
        f'<w:ftr{namespace_match.group(1)}>'
        '<w:p><w:pPr><w:jc w:val="center"/></w:pPr>'
        '<w:r><w:t xml:space="preserve">第 </w:t></w:r>'
        '<w:r><w:fldChar w:fldCharType="begin"/></w:r>'
        '<w:r><w:instrText xml:space="preserve"> PAGE </w:instrText></w:r>'
        '<w:r><w:fldChar w:fldCharType="separate"/></w:r>'
        '<w:r><w:t>1</w:t></w:r>'
        '<w:r><w:fldChar w:fldCharType="end"/></w:r>'
        '<w:r><w:t xml:space="preserve"> 页</w:t></w:r>'
        '</w:p></w:ftr>'
    ).encode("utf-8")


def repair(
    source: Path,
    output: Path,
    repair_overall: bool = True,
    simple_footer: bool = True,
    materialize_charts: bool = True,
    repair_comparison: bool = True,
    repair_footers: bool = True,
) -> None:
    with zipfile.ZipFile(source) as archive:
        infos = archive.infolist()
        parts = {info.filename: archive.read(info.filename) for info in infos}
    if "word/document.xml" not in parts or "word/_rels/document.xml.rels" not in parts:
        raise RuntimeError("DOCX main parts missing")
    removed = 0
    if materialize_charts:
        for name, body in list(parts.items()):
            parts[name], count = remove_external_chart_artifacts(name, body)
            removed += count
    document = parts["word/document.xml"].decode("utf-8")
    if repair_comparison:
        document = ungroup_comparison_chart(document, parts["word/_rels/document.xml.rels"])
    if repair_overall:
        document = ungroup_overall_chart(document, parts["word/_rels/document.xml.rels"])
    document = normalize_page_number_sections(document)
    parts["word/document.xml"] = document.encode("utf-8")
    footer_count = 0
    if repair_footers:
        for name, body in list(parts.items()):
            if re.fullmatch(r"word/footer\d+\.xml", name) and (b"NUMPAGES" in body.upper() or "页 共" in body.decode("utf-8")):
                parts[name] = normalize_footer(body, simple_footer)
                footer_count += 1
        if footer_count == 0:
            raise RuntimeError("NUMPAGES footer not found")
        if b"NUMPAGES" in b"".join(body.upper() for name, body in parts.items() if name.startswith("word/footer")):
            raise RuntimeError("NUMPAGES remains")
    if b'w:val="validity.text"' in parts["word/document.xml"]:
        raise RuntimeError("optional validity.text was unexpectedly added")
    output.parent.mkdir(parents=True, exist_ok=True)
    with tempfile.NamedTemporaryFile(dir=output.parent, suffix=".docx", delete=False) as temporary:
        temporary_path = Path(temporary.name)
    try:
        with zipfile.ZipFile(temporary_path, "w") as archive:
            for source_info in infos:
                info = zipfile.ZipInfo(source_info.filename, source_info.date_time)
                info.compress_type = zipfile.ZIP_STORED if source_info.is_dir() or not parts[source_info.filename] else zipfile.ZIP_DEFLATED
                info.external_attr = source_info.external_attr
                info.internal_attr = source_info.internal_attr
                info.create_system = source_info.create_system
                archive.writestr(info, parts[source_info.filename])
        with zipfile.ZipFile(temporary_path) as verification:
            if verification.testzip() is not None:
                raise RuntimeError("repaired DOCX ZIP invalid")
        temporary_path.replace(output)
    finally:
        temporary_path.unlink(missing_ok=True)
    print(f"CUSTOMER_V2_TEMPLATE_REPAIRED externalArtifacts={removed} footers={footer_count} output={output}")


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--source", type=Path, default=DEFAULT_SOURCE)
    parser.add_argument("--output", type=Path, default=DEFAULT_OUTPUT)
    parser.add_argument("--skip-overall-repair", action="store_true")
    parser.add_argument("--keep-floating-footer", action="store_true")
    parser.add_argument("--skip-chart-materialization", action="store_true")
    parser.add_argument("--skip-comparison-repair", action="store_true")
    parser.add_argument("--skip-footer-repair", action="store_true")
    args = parser.parse_args()
    repair(
        args.source.resolve(),
        args.output.resolve(),
        repair_overall=not args.skip_overall_repair,
        simple_footer=not args.keep_floating_footer,
        materialize_charts=not args.skip_chart_materialization,
        repair_comparison=not args.skip_comparison_repair,
        repair_footers=not args.skip_footer_repair,
    )


if __name__ == "__main__":
    main()
