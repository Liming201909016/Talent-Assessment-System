#!/usr/bin/env python3
"""MT-ZERO-RING-01: local, independent Go-DOCX -> LibreOffice -> PDF gate.

No DOCX edits, network, database, old-oracle execution, or old-result writes.
Every invocation owns a new result directory. --red-only intentionally exits 1.
--convert-only retains conversions for inspection; --existing-run validates them.
Short synthetic Go DTOs are not evidence for persisted customer-text pagination.
"""

import argparse
from decimal import Decimal
import hashlib
import importlib.util
import io
import json
import math
from pathlib import Path
import re
import subprocess
import sys
from types import SimpleNamespace
import unittest
import uuid
import xml.etree.ElementTree as ET
import zipfile

import pymupdf as fitz

ROOT = Path(__file__).resolve().parents[2]
RESULTS = ROOT / 'scripts/test/results/management-traits-zero-ring-local-20261003'
HANDOFF = ROOT / 'Go-based Refactored System/tmp/mng-zero-ring-final-20261003-9fd72d7481d7'
OLD = ROOT / 'scripts/test/results/MTRa05eb3af5e49'
TEMPLATE = ROOT / 'Go-based Refactored System/configs/export-templates/management-traits-002-test-only-v2.docx'
TEMPLATE_SHA = '05c55e77e567c6111ba62c08f2e69b4d5e5b416b2dd989c49914cc95b962a84c'
NS = {'c': 'http://schemas.openxmlformats.org/drawingml/2006/chart',
      'a': 'http://schemas.openxmlformats.org/drawingml/2006/main',
      'w': 'http://schemas.openxmlformats.org/wordprocessingml/2006/main',
      'wp': 'http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing'}
# Order is semantic, not physical XML numbering. chart5 is interpersonal.
PARTS = [1, 2, 5, 3, 4]
KEYS = ['overall', 'self', 'interpersonal', 'task', 'development']
SCORES = {
    'all-zero': ['0'] * 5, 'near-zero': ['.004'] * 5,
    'score-25': ['25'] * 5, 'score-50': ['50'] * 5,
    'score-75': ['75'] * 5, 'score-100': ['100'] * 5,
    'mixed': ['25', '0', '.004', '100', '0'],
    'signed-zero': ['-0.000', '0.00', '+0', '-0', '0.000000000000'],
}
BLUE = (72, 116, 203)
GRAY = (231, 230, 230)


def receipt(path):
    data = path.read_bytes()
    return {'path': str(path), 'bytes': len(data), 'sha256': hashlib.sha256(data).hexdigest()}


def require(condition, message, errors):
    if not condition:
        errors.append(message)


def package(path):
    with zipfile.ZipFile(path) as z:
        assert len(z.namelist()) == len(set(z.namelist())), 'duplicate OPC parts'
        assert z.testzip() is None, 'DOCX CRC failure'
        return {name: z.read(name) for name in z.namelist()}


def scrub(data, prefix, local):
    # Preserve ALL other bytes, including whitespace, coordinates and styles.
    return re.sub(rb'(<'+prefix.encode()+b':'+local.encode()+rb'\b[^>]*>).*?(</'+
                  prefix.encode()+b':'+local.encode()+rb'>)', rb'\1\2', data, flags=re.S)


def docx_contract(path, name, original):
    actual = package(path)
    errors = []
    require(set(actual) == set(original), 'OPC part set changed', errors)
    values = [Decimal(x) for x in SCORES[name]]
    rings = []
    for key, part, value, literal in zip(KEYS, PARTS, values, SCORES[name], strict=True):
        part_name = f'word/charts/chart{part}.xml'
        root = ET.fromstring(actual[part_name])
        points = root.findall('.//c:ser/c:val/c:numLit/c:pt/c:v', NS)
        require(len(points) == 2 and points[0].text == literal and
                Decimal(points[1].text) == 100-value, f'{key}: score/complement changed', errors)
        remainder = next(x for x in root.findall('.//c:ser/c:dPt', NS)
                         if x.find('c:idx', NS).get('val') == '1')
        sp = remainder.find('c:spPr', NS)
        color = sp.find('a:solidFill/a:srgbClr', NS)
        require(sp.find('a:ln/a:noFill', NS) is not None, f'{key}: outline changed', errors)
        if value == 0:
            require(color is not None and color.get('val') == 'E7E6E6' and
                    sp.find('a:noFill', NS) is None, f'{key}: exact-zero gray fill missing', errors)
        else:
            require(color is None and sp.find('a:noFill', NS) is not None,
                    f'{key}: nonzero transparent remainder changed', errors)
        expected = original[part_name]
        if value == 0:
            # Only direct remainder fill is authorized; outline noFill remains.
            pt = re.search(rb'<c:dPt>\s*<c:idx val="1"\s*/>.*?</c:dPt>', expected, re.S)
            assert pt is not None, 'original remainder point missing'
            before = pt.group()
            after = re.sub(rb'(<c:spPr>\s*)<a:noFill\s*/>',
                rb'\1<a:solidFill><a:srgbClr val="E7E6E6"/></a:solidFill>', before, count=1)
            expected = expected[:pt.start()] + after + expected[pt.end():]
        require(scrub(actual[part_name], 'c', 'v') == scrub(expected, 'c', 'v'),
                f'{key}: unauthorized chart style byte change', errors)
        rings.append({'key': key, 'part': part_name, 'literal': literal,
                      'exactZero': value == 0, 'grayRemainder': color is not None})
    for part_name, data in original.items():
        if part_name == 'word/document.xml':
            require(scrub(actual[part_name], 'w', 't') == scrub(data, 'w', 't'),
                    'document non-text/style/coordinate bytes changed', errors)
        elif part_name == 'word/charts/chart6.xml':
            require(scrub(actual[part_name], 'c', 'v') == scrub(data, 'c', 'v'),
                    'comparison chart style changed', errors)
        elif part_name not in {f'word/charts/chart{i}.xml' for i in PARTS}:
            require(actual[part_name] == data, f'unrelated OPC bytes changed: {part_name}', errors)
    comparison = ET.fromstring(actual['word/charts/chart6.xml'])
    series = comparison.findall('.//c:ser', NS)
    expected_dims = [values[i % 5] if name == 'mixed' else values[0] for i in range(13)]
    require(len(series) == 2, 'comparison series count is not two', errors)
    data = [[Decimal(v.text) for v in s.findall('c:val/c:numLit/c:pt/c:v', NS)] for s in series]
    old = ET.fromstring(original['word/charts/chart6.xml']).findall('.//c:ser', NS)
    norms = [Decimal(v.text) for v in old[1].findall('c:val/c:numLit/c:pt/c:v', NS)]
    require(data == [expected_dims, norms] and len(norms) == 13,
            'comparison 13 scores/norms differ', errors)
    return {'file': receipt(path), 'rings': rings, 'errors': errors}


def rgb_matches(pixel, color, tolerance=3):
    return all(abs(a-b) <= tolerance for a, b in zip(pixel, color, strict=True))


def fill_rgb(drawing):
    fill = drawing.get('fill')
    return tuple(round(x*255) for x in fill) if fill is not None else None


def geometry(reference):
    with fitz.open(reference) as pdf:
        candidates = [tuple(d['rect']) for d in pdf[2].get_drawings()
                      if fill_rgb(d) == BLUE and 100 < d['rect'].y0 < 370
                      and 70 < d['rect'].width < 150 and
                      abs(d['rect'].width-d['rect'].height) < 1]
    assert len(candidates) == 5, f'100-score complete annulus geometry missing: {candidates}'
    overall = max(candidates, key=lambda r: r[2]-r[0])
    modules = sorted([r for r in candidates if r != overall], key=lambda r: r[0])
    left = sorted(modules[:2], key=lambda r: r[1])
    right = sorted(modules[2:], key=lambda r: r[1])
    ordered = [overall, left[0], left[1], right[0], right[1]]
    return {key: list(rect) for key, rect in zip(KEYS, ordered, strict=True)}


def annulus(page, rect, color):
    bounds = fitz.Rect(rect)
    center = ((bounds.x0+bounds.x1)/2, (bounds.y0+bounds.y1)/2)
    radius = bounds.width/2
    scale = 3  # 216 DPI: interior sampling avoids antialiased edges.
    pix = page.get_pixmap(matrix=fitz.Matrix(scale, scale), colorspace=fitz.csRGB, alpha=False)
    samples = pix.samples

    def pixel(x, y):
        ix, iy = round(x*scale)-pix.x, round(y*scale)-pix.y
        if not (0 <= ix < pix.width and 0 <= iy < pix.height):
            return (0, 0, 0)
        offset = iy*pix.stride+ix*pix.n
        return tuple(samples[offset:offset+3])

    occupancy = []
    # Six radial samples in the ring, one angular bin per degree, full 360 deg.
    for degree in range(360):
        angle = math.radians(degree+.5)
        hits = sum(rgb_matches(pixel(center[0]+radius*f*math.cos(angle),
                                    center[1]+radius*f*math.sin(angle)), color)
                   for f in (.64, .69, .74, .79, .84, .89))
        occupancy.append(hits/6)
    hole = [rgb_matches(pixel(center[0]+radius*f*math.cos(math.radians(a)),
                              center[1]+radius*f*math.sin(math.radians(a))), (255, 255, 255))
            for f in (0, .1, .2, .3, .4) for a in range(0, 360, 5)]
    # Genuine annulus must span its full envelope and contain a white hole.
    vectors = [d for d in page.get_drawings() if fill_rgb(d) == color and
               max(abs(a-b) for a, b in zip(d['rect'], rect, strict=True)) <= 1]
    target_pixels = sum(rgb_matches(pixel(x/scale, y/scale), color)
                        for y in range(math.ceil(bounds.y0*scale), math.floor(bounds.y1*scale))
                        for x in range(math.ceil(bounds.x0*scale), math.floor(bounds.x1*scale))
                        if .51*radius <= math.hypot(x/scale-center[0], y/scale-center[1]) <= .99*radius)
    return {'bbox': list(rect), 'center': list(center), 'outerRadius': radius,
            'rgb': list(color), 'dpi': 216, 'angularBins': 360,
            'occupiedBins': sum(x >= 5/6 for x in occupancy),
            'angularOccupancy': sum(x >= 5/6 for x in occupancy)/360,
            'minRadialOccupancy': min(occupancy), 'matchingPixels': target_pixels,
            'whiteHoleOccupancy': sum(hole)/len(hole),
            'completeVectorCount': len(vectors)}


def spans(page):
    return [s for b in page.get_text('dict')['blocks'] for line in b.get('lines', [])
            for s in line['spans']]


def text_signature(page):
    return [{'text': s['text'], 'bbox': [round(v, 3) for v in s['bbox']],
             'font': s['font'], 'size': round(s['size'], 4), 'flags': s['flags'],
             'color': s['color']} for s in spans(page)]


def comparison_paths(page):
    # Geometry is explicitly below the five annuli, never counted as rings.
    paths = []
    for d in page.get_drawings():
        r = d['rect']
        color = d.get('color')
        if color is not None and 380 < r.y0 < 550 and r.width > 300 and r.height > 3:
            rgb = tuple(round(v*255) for v in color)
            if max(rgb)-min(rgb) <= 2 and 50 < rgb[0] < 240:
                paths.append({'bbox': [round(v, 3) for v in r], 'rgb': list(rgb),
                              'items': repr(d['items']), 'width': d['width']})
    return paths


def pdf_contract(path, name, envelopes, directory):
    errors = []
    pdf = fitz.open(path)
    text = ''.join(p.get_text(sort=False) for p in pdf)
    require(len(pdf) >= 6, 'fewer than six physical pages', errors)
    require('\ufffd' not in text and not re.search(r'\{\{|\}\}|chart\.module\.|overview\.high\.', text),
            'unresolved placeholder/replacement character', errors)
    require('TEST' in text and '不可作为人才决策依据' in text, 'TEST purpose label missing', errors)
    pages = []
    directory.mkdir(parents=True, exist_ok=False)
    for i, page in enumerate(pdf):
        require(abs(page.rect.width-595.3) < 2 and abs(page.rect.height-841.9) < 2,
                f'page {i+1}: not A4', errors)
        pix = page.get_pixmap(matrix=fitz.Matrix(1, 1), colorspace=fitz.csRGB, alpha=False)
        sample_bytes = pix.samples
        ink = sum(min(sample_bytes[j:j+3]) < 240 for j in range(0, len(sample_bytes), 3))
        require(bool(page.get_text().strip()) and ink > 1000, f'page {i+1}: blank', errors)
        pix.save(directory / f'page-{i+1:02d}.png')
        pages.append({'page': i+1, 'size': list(page.rect), 'textLength': len(page.get_text()),
                      'inkPixels72DPI': ink})
    page = pdf[2]
    labels = [s for s in spans(page) if re.match(r'^\d+\.\d{2} 分', s['text'])
              and 150 < s['bbox'][1] < 385]
    require(len(labels) == 5, f'five full numeric slots missing: {len(labels)}', errors)
    ordered_labels = sorted(labels, key=lambda s: s['bbox'][0])
    if len(labels) == 5:
        overall = next(s for s in labels if 200 < s['bbox'][0] < 400)
        left = sorted([s for s in labels if s['bbox'][0] < 200], key=lambda s: s['bbox'][1])
        right = sorted([s for s in labels if s['bbox'][0] > 400], key=lambda s: s['bbox'][1])
        ordered_labels = [overall, left[0], left[1], right[0], right[1]]
        for key, value, s in zip(KEYS, SCORES[name], ordered_labels, strict=True):
            expected = f'{Decimal(value):.2f}'.replace('-0.00', '0.00')
            require(s['text'].startswith(expected+' 分'), f'{key}: numeric token not {expected}', errors)
    rings = []
    for key, value in zip(KEYS, SCORES[name], strict=True):
        zero = Decimal(value) == 0
        gray = annulus(page, envelopes[key], GRAY)
        blue = annulus(page, envelopes[key], BLUE)
        if zero:
            require(gray['angularOccupancy'] >= .95 and gray['completeVectorCount'] == 1
                    and gray['whiteHoleOccupancy'] >= .99,
                    f'{key}: missing complete E7E6E6 gray annulus (360-degree gate)', errors)
            require(blue['matchingPixels'] == 0, f'{key}: zero has blue score pixels', errors)
        else:
            require(gray['matchingPixels'] == 0, f'{key}: nonzero has added gray', errors)
            score = float(Decimal(value))
            if score >= 25:
                require(abs(blue['angularOccupancy']-score/100) <= .025,
                        f'{key}: blue arc occupancy differs from original fraction', errors)
            require(blue['whiteHoleOccupancy'] >= .99, f'{key}: white hole lost', errors)
        rings.append({'key': key, 'exactZero': zero, 'gray': gray, 'blue': blue})
    norm_paths = comparison_paths(page)
    require(len(norm_paths) == 1, f'13-dimension gray norm polyline missing: {len(norm_paths)}', errors)
    if norm_paths:
        path_points = sum(1 for x in page.get_drawings()
                          if x.get('color') is not None and repr(x['items']) == norm_paths[0]['items']
                          for item in x['items'] if item[0] == 'l')
        require(path_points == 12, f'norm polyline not 13 connected dimensions: {path_points}', errors)
    compact = re.sub(r'\s+', '', page.get_text(sort=False))
    for label in ['综合均值', '自我管理均值', '人际管理均值', '任务管理均值', '发展管理均值', '常模参照分']:
        require(label in compact, f'page3 fixed chart label missing: {label}', errors)
    advice_pages = [i+1 for i, p in enumerate(pdf)
                    if '各维度发展建议' in re.sub(r'\s+', '', p.get_text())]
    require(len(advice_pages) == 1, 'development-advice fixed label missing/duplicated', errors)
    result = {'file': receipt(path), 'pages': len(pdf), 'pageEvidence': pages,
              'producer': pdf.metadata.get('producer'), 'creator': pdf.metadata.get('creator'),
              'developmentAdviceLabelPages': advice_pages,
              'numericSlots': [{k: s[k] for k in ('text', 'bbox', 'font', 'size', 'flags', 'color')}
                               for s in ordered_labels], 'rings': rings, 'normPolyline': norm_paths,
              'visibleCompleteGrayRings': sum(r['exactZero'] and r['gray']['angularOccupancy'] >= .95
                                            and r['gray']['completeVectorCount'] == 1 for r in rings),
              'visibleDonutCount': sum(r['gray']['completeVectorCount'] == 1 or
                                      r['blue']['matchingPixels'] > 100 for r in rings),
              'sixCharts': all(r['gray']['completeVectorCount'] == 1 or
                               r['blue']['matchingPixels'] > 100 for r in rings) and len(norm_paths) == 1,
              'errors': errors}
    pdf.close()
    return result


def convert(soffice, source, target, profile, timeout):
    target.mkdir(parents=True, exist_ok=False)
    command = [str(soffice), '-env:UserInstallation='+profile.resolve().as_uri(),
               '--headless', '--nologo', '--nodefault', '--norestore',
               '--convert-to', 'pdf:writer_pdf_Export', '--outdir', str(target), str(source)]
    result = subprocess.run(command, capture_output=True, timeout=timeout, check=False)
    evidence = {'command': command, 'exit': result.returncode,
                'stdout': result.stdout.decode('utf-8', errors='replace'),
                'stderr': result.stderr.decode('utf-8', errors='replace')}
    (target/'conversion.json').write_text(json.dumps(evidence, indent=2), encoding='utf-8')
    assert result.returncode == 0, f'LibreOffice failed: {evidence}'
    pdf = target/(source.stem+'.pdf')
    assert pdf.is_file() and pdf.read_bytes().startswith(b'%PDF-'), 'conversion did not produce true PDF'
    return pdf


def upstream_contracts(directory, soffice):
    """Run unchanged original 10/17 assertions; redirect only PDF output state.

    The original CLI lacks a compatible --outdir and would overwrite old PDFs.
    Import it without invoking main/write-evidence. No test method is replaced.
    The PDF setup wrapper changes only the per-class output directory after the
    original setup; exact source/template assertions and all thresholds remain.
    """
    source = ROOT/'scripts/test/management-traits-002-word-candidate-contract-test.py'
    spec = importlib.util.spec_from_file_location('mng_zero_upstream', source)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    module.ARGS = SimpleNamespace(strict_visual_red=False, soffice=soffice,
                                 docx=module.OUT/'management-traits-002-candidate.docx')
    setup = module.LibreOfficeCompatiblePDFContract.setUpClass.__func__

    def isolated_setup(cls):
        setup(cls)
        cls.output_dir = directory/'upstream-compatible'

    module.LibreOfficeCompatiblePDFContract.setUpClass = classmethod(isolated_setup)
    summaries = {}
    for name in ('source-layout', 'compatible'):
        suite = unittest.defaultTestLoader.loadTestsFromTestCase(
            module.SourceLayoutContract if name == 'source-layout' else module.LibreOfficeCompatibleContract)
        if name == 'compatible':
            for method in sorted(x for x in dir(module.ActualPDFVisualContract)
                                 if x.startswith('test_actual_pdf_')):
                suite.addTest(module.LibreOfficeCompatiblePDFContract(method))
        stream = io.StringIO()
        result = unittest.TextTestRunner(stream=stream, verbosity=2,
                    resultclass=module.EvidenceTestResult).run(suite)
        (directory/('upstream-'+name+'.txt')).write_text(stream.getvalue(), encoding='utf-8')
        summaries[name] = {'tests': result.testsRun, 'failures': len(result.failures),
                           'errors': len(result.errors), 'skipped': len(result.skipped),
                           'passedSubtests': result.passed_subtests,
                           'unchangedTestMethods': True,
                           'success': result.wasSuccessful()}
        print('UPSTREAM_CONTRACT='+name+' '+json.dumps(summaries[name]), flush=True)
    return summaries


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--input-dir', type=Path, default=HANDOFF)
    parser.add_argument('--soffice', type=Path, default=Path('C:/Program Files/LibreOffice/program/soffice.com'))
    parser.add_argument('--timeout', type=int, default=90)
    parser.add_argument('--red-only', action='store_true')
    parser.add_argument('--convert-only', action='store_true')
    parser.add_argument('--existing-run', type=Path)
    parser.add_argument('--upstream-safe', action='store_true', help='unchanged original 10/17 contracts, isolated PDF outputs')
    args = parser.parse_args()
    assert 1 <= args.timeout <= 180, 'finite conversion timeout must be 1..180 seconds'
    directory = RESULTS/uuid.uuid4().hex
    directory.mkdir(parents=True, exist_ok=False)
    print('ZERO_RING_RESULTS='+str(directory), flush=True)
    report = {'task': 'MT-ZERO-RING-01', 'localOnly': True, 'directory': str(directory),
              'python': sys.version, 'pymupdf': fitz.__version__, 'errors': [],
              'scope': 'supplied short synthetic Go DTOs; not persisted full customer-text acceptance'}
    try:
        protected = [TEMPLATE, Path(__file__), ROOT/'scripts/test/management-traits-four-real-pdf-oracle.py',
                     ROOT/'scripts/test/management-traits-002-word-candidate-contract-test.py']
        protected += sorted(p for p in OLD.rglob('*') if p.is_file())
        protected += sorted(args.input_dir.glob('*.docx'))
        if args.upstream_safe:
            prior = ROOT/'docs/generated/management-traits-word-candidate-20261001'
            protected += sorted(p for p in prior.rglob('*') if p.is_file())
        before = {str(p): receipt(p) for p in protected}
        assert receipt(TEMPLATE)['sha256'] == TEMPLATE_SHA, 'runtime template SHA changed'
        red_geometry = geometry(OLD/'00202-candidate.pdf')
        red = pdf_contract(OLD/'00201-tester.pdf', 'all-zero', red_geometry, directory/'red-pages')
        report['red'] = red
        missing = [r['key'] for r in red['rings'] if r['gray']['angularOccupancy'] < .95]
        assert len(missing) == 5 and red['visibleCompleteGrayRings'] == 0, 'old real PDF failed to reproduce five missing rings'
        print('ACTUAL_RED_MISSING='+json.dumps(missing)+' GRAY_GATE_EXIT=1', flush=True)
        if args.red_only:
            report['errors'] = red['errors']
        else:
            original = package(TEMPLATE)
            report['docx'] = [docx_contract(args.input_dir/(name+'.docx'), name, original) for name in SCORES]
            for d in report['docx']:
                report['errors'] += d['errors']
            assert not report['errors'], 'Go handoff DOCX contract failed'
            pdfs = {}
            if args.existing_run:
                previous = args.existing_run.resolve()
                assert previous.is_relative_to(RESULTS.resolve()), 'existing run must be inside independent result root'
                pdfs = {name: previous/name/(name+'.pdf') for name in SCORES}
            else:
                assert args.soffice.is_file(), 'existing soffice.com missing'
                version = subprocess.run([str(args.soffice), '--version'], capture_output=True, timeout=15, check=True)
                report['libreofficeVersion'] = version.stdout.decode('utf-8', errors='replace').strip()
                for name in SCORES:
                    source = args.input_dir/(name+'.docx')
                    assert source.is_file(), f'missing actual Go artifact {source}'
                    pdfs[name] = convert(args.soffice, source, directory/name, directory/('profile-'+name), args.timeout)
                    print('CONVERTED='+name+' '+json.dumps(receipt(pdfs[name])), flush=True)
            report['convertedPDFs'] = {name: receipt(path) for name, path in pdfs.items()}
            if not args.convert_only:
                envelopes = geometry(pdfs['score-100'])
                report['geometryReference'] = receipt(pdfs['score-100'])
                report['geometry'] = envelopes
                report['pdf'] = []
                for name, path in pdfs.items():
                    evidence = pdf_contract(path, name, envelopes, directory/('pages-'+name))
                    report['pdf'].append(evidence)
                    report['errors'] += [name+': '+e for e in evidence['errors']]
                    print('PDF_GATE='+json.dumps({'name': name, 'pages': evidence['pages'],
                          'grayRings': evidence['visibleCompleteGrayRings'], 'numericSlots': len(evidence['numericSlots']),
                          'errors': evidence['errors']}), flush=True)
                # The five independent numeric slots are invariant. Comparison
                # data-label y follows genuine 0 vs .004 bar data, not a style edit.
                with fitz.open(pdfs['all-zero']) as zero, fitz.open(pdfs['near-zero']) as near:
                    signatures = [text_signature(p) for p in zero]
                    other = [text_signature(p) for p in near]
                    dynamic = lambda s: s['text'].strip() == '0.00' and 380 < s['bbox'][1] < 550
                    plot_zero = [s for s in signatures[2] if dynamic(s)]
                    plot_near = [s for s in other[2] if dynamic(s)]
                    frozen = [[s for s in page if not dynamic(s)] for page in signatures]
                    frozen_near = [[s for s in page if not dynamic(s)] for page in other]
                    same = frozen == frozen_near
                    slots_equal = report['pdf'][0]['numericSlots'] == report['pdf'][1]['numericSlots']
                    shifts = []
                    require(len(plot_zero) == len(plot_near) == 13,
                            'zero/near comparison data labels are not thirteen', report['errors'])
                    for a, b in zip(plot_zero, plot_near, strict=True):
                        aa, bb = dict(a), dict(b)
                        ab, bc = aa.pop('bbox'), bb.pop('bbox')
                        delta = bc[1]-ab[1]
                        require(aa == bb and ab[0] == bc[0] and ab[2] == bc[2] and
                                abs((bc[3]-ab[3])-delta) < .002 and abs(delta) <= .05,
                                'comparison data label has non-value-only geometry/style drift', report['errors'])
                        shifts.append({'zero': ab, 'near': bc, 'deltaYpt': delta})
                    report['zeroNearTextGeometry'] = {'allPagesExactEqual': signatures == other,
                        'allNonPlotSpansExactEqual': same, 'fiveNumericSlotsExactEqual': slots_equal,
                        'pages': len(zero), 'page3': signatures[2], 'comparisonValueLabelOffsets': shifts}
                    require(same and slots_equal, 'zero vs .004 fixed text/font/bbox/page layout drift', report['errors'])
                baseline = report['pdf'][0]['normPolyline']
                require(all(x['normPolyline'] == baseline for x in report['pdf']),
                        'gray comparison norm geometry changed between scores', report['errors'])
                if args.upstream_safe:
                    report['upstream'] = upstream_contracts(directory, args.soffice)
                    require(all(x['success'] for x in report['upstream'].values()),
                            'unchanged original upstream contracts failed', report['errors'])
        after = {str(p): receipt(p) for p in protected}
        report['protectedUnchanged'] = before == after
        report['protectedFiles'] = after
        require(before == after, 'original input/runtime template/old evidence bytes changed', report['errors'])
    except Exception as error:
        report['errors'].append(type(error).__name__+': '+str(error))
    report['status'] = 'FAIL' if report['errors'] else ('CONVERSION_ONLY' if args.convert_only else 'PASS')
    output = directory/'contract.json'
    output.write_text(json.dumps(report, indent=2, ensure_ascii=False), encoding='utf-8')
    print('ZERO_RING_CONTRACT='+json.dumps({'status': report['status'], 'errors': report['errors'],
                                         'evidence': receipt(output)}), flush=True)
    return 1 if report['errors'] else 0


if __name__ == '__main__':
    sys.exit(main())