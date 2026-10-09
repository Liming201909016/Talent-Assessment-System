#!/usr/bin/env python3
"""Local candidate contract; --docx ORIGINAL is the intentional source RED."""

import argparse
import ast
import copy
from datetime import datetime, timezone
from fractions import Fraction
import hashlib
import importlib.util
import io
import json
import re
import subprocess
import uuid
from pathlib import Path
import sys
import unittest
import xml.etree.ElementTree as ET
import zipfile

ROOT = Path(__file__).resolve().parents[2]
OUT = ROOT / 'docs/generated/management-traits-word-candidate-20261001'
SOURCE = ROOT / 'docs/260929管理特质测评-优化/260928管理潜质测评报告模板修改稿V2.8.docx'
SHA = 'c82c2dc0cdcaec413866268c557561d5c53abc41ec02d94dc6aea615edca2f48'
NS = {
    'w': 'http://schemas.openxmlformats.org/wordprocessingml/2006/main',
    'wp': 'http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing',
    'a': 'http://schemas.openxmlformats.org/drawingml/2006/main',
    'c': 'http://schemas.openxmlformats.org/drawingml/2006/chart',
    'r': 'http://schemas.openxmlformats.org/officeDocument/2006/relationships',
    'wpg': 'http://schemas.microsoft.com/office/word/2010/wordprocessingGroup',
    'wps': 'http://schemas.microsoft.com/office/word/2010/wordprocessingShape',
}
KEYS = {'chart.overall', 'chart.module.self', 'chart.module.interpersonal',
        'chart.module.task', 'chart.module.development', 'chart.dimension.comparison'}
ARGS = None


def parts(path):
    with zipfile.ZipFile(path) as archive:
        if len(archive.namelist()) != len(set(archive.namelist())):
            raise ValueError('duplicate ZIP entries')
        if archive.testzip() is not None:
            raise ValueError('CRC failure')
        return {name: archive.read(name) for name in archive.namelist()}


def xml(data):
    return ET.fromstring(data)


class PackageContract(unittest.TestCase):
    def setUp(self):
        self.package = parts(ARGS.docx)
        self.doc = xml(self.package['word/document.xml'])

    def test_actual_dynamic_sdt_slots(self):
        tags = [x.get('{'+NS['w']+'}val') for x in self.doc.findall('.//w:sdtPr/w:tag', NS)]
        self.assertIn('participant.name', tags, 'original has zero SDTs')
        self.assertIn('overall.level', tags)
        self.assertIn('overview.high.3.text', tags)
        self.assertNotIn('participant.age', tags)
        self.assertNotIn('result.userTime', tags)
        self.assertNotIn('overall.norm', tags)

    def test_six_native_anchor_business_charts(self):
        charts = self.doc.findall('.//c:chart', NS)
        self.assertEqual(len(charts), 6)
        self.assertEqual(len(self.doc.findall('.//wpg:graphicFrame//c:chart', NS)), 0,
                         'six charts are still grouped graphicFrames')
        anchors = self.doc.findall('.//wp:anchor', NS)
        bound = {x.find('wp:docPr', NS).get('title') for x in anchors
                 if x.find('a:graphic/a:graphicData/c:chart', NS) is not None}
        self.assertEqual(bound, KEYS)

    def test_all_opc_literal_only_no_external_relationships(self):
        errors = []
        forbidden = ('f', 'externalData', 'numRef', 'strRef', 'multiLvlStrRef')
        for name, data in self.package.items():
            if name.endswith(('.xml', '.rels')):
                root = xml(data)
                for node in root.iter():
                    if node.tag in {'{'+NS['c']+'}'+tag for tag in forbidden}:
                        errors.append(name + ':' + node.tag.split('}')[-1])
                    if name.endswith('.rels') and node.get('TargetMode') == 'External':
                        errors.append(name + ':External')
        self.assertEqual(errors, [], 'original has formula/ref/external artifacts')

    def test_visual_restore_native_anchors_keep_source_frame_coordinates(self):
        original = xml(parts(SOURCE)['word/document.xml'])
        frames = original.findall('.//wpg:graphicFrame', NS)
        parents = {child: parent for parent in original.iter() for child in parent}
        for frame in frames:
            rid = frame.find('.//c:chart', NS).get('{'+NS['r']+'}id')
            anchors = [a for a in self.doc.findall('.//wp:anchor', NS)
                       if a.find('a:graphic/a:graphicData/c:chart', NS) is not None
                       and a.find('a:graphic/a:graphicData/c:chart', NS).get('{'+NS['r']+'}id') == rid]
            self.assertEqual(len(anchors), 1, 'native anchor missing: '+rid)
            group = parents[frame].find('wpg:grpSpPr/a:xfrm', NS)
            xfrm = frame.find('wpg:xfrm', NS)
            anchor = anchors[0]
            for axis, size in [('x','cx'), ('y','cy')]:
                scale = Fraction(int(group.find('a:ext', NS).get(size)),
                                 int(group.find('a:chExt', NS).get(size)))
                offset = round((int(xfrm.find('a:off', NS).get(axis))-
                                int(group.find('a:chOff', NS).get(axis)))*scale)
                position = anchor.find('wp:position'+('H' if axis == 'x' else 'V'), NS)
                self.assertEqual(position.get('relativeFrom'), 'column' if axis == 'x' else 'paragraph')
                outer = parents[parents[parents[parents[frame]]]]
                width = int(outer.find('wp:extent', NS).get('cx'))
                # Both source groups share a centered body paragraph, section 2.
                base = ((11906-850-850)*635-width)//2 if axis == 'x' else (
                    0 if outer.find('wp:docPr', NS).get('id') == '67' else round(Fraction('183.95')*12700))
                self.assertEqual(int(position.find('wp:posOffset', NS).text), base+offset)
                self.assertEqual(int(anchor.find('wp:extent', NS).get(size)),
                                 round(int(xfrm.find('a:ext', NS).get(size))*scale))
            self.assertIsNotNone(anchor.find('wp:wrapNone', NS))
        self.assertEqual(len(self.doc.findall('.//wp:inline//c:chart', NS)), 0,
                         'chart flow must not append six ordinary inline charts')

    def test_visual_restore_fixed_background_shapes_are_native_and_co_located(self):
        # The fixed nested group also drifts in LO after its chart is removed.
        self.assertFalse(self.doc.findall('.//wpg:wgp', NS))
        shapes = [a for a in self.doc.findall('.//wp:anchor', NS)
                  if a.find('wp:docPr', NS).get('title', '').startswith('fixed.grade.')]
        self.assertEqual(len(shapes), 5)
        for shape in shapes:
            self.assertIsNotNone(shape.find('a:graphic/a:graphicData/wps:wsp/wps:txbx', NS))
            self.assertEqual(shape.find('wp:positionV', NS).get('relativeFrom'), 'paragraph')

    def test_visual_restore_ring_titles_are_score_independent_and_bounded(self):
        for i in range(1, 6):
            root = xml(self.package[f'word/charts/chart{i}.xml'])
            self.assertIsNone(root.find('c:chart/c:title', NS), 'LO chart title ignores source text-box width')
            self.assertFalse(root.findall('.//c:dLbl/c:tx', NS))
        labels = [a for a in self.doc.findall('.//wp:anchor', NS)
                  if a.find('wp:docPr', NS).get('title', '').endswith('.numeric-label')]
        self.assertEqual(len(labels), 5, 'five independent source-sized text boxes required')
        for label in labels:
            self.assertTrue(label.findall('.//wps:txbx/w:txbxContent/w:p/w:r/w:rPr', NS))
            self.assertGreater(int(label.find('wp:extent', NS).get('cx')), 0)
            self.assertEqual(label.find('wp:positionV', NS).get('relativeFrom'), 'paragraph')


class BuildAndDemoContract(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        path = ROOT / 'scripts/tools/build-management-traits-002-word-candidate.py'
        spec = importlib.util.spec_from_file_location('mng_candidate', path)
        cls.builder = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(cls.builder)
        cls.original = parts(SOURCE)
        cls.data, cls.manifest = cls.builder.build_candidate(SOURCE)
        cls.package = parts(io.BytesIO(cls.data))
        cls.demo, cls.evidence = cls.builder.make_demo(cls.data, cls.manifest)

    def test_source_hash_determinism_and_reject_source_target(self):
        before = SOURCE.read_bytes()
        self.assertEqual(hashlib.sha256(before).hexdigest(), SHA)
        again, manifest = self.builder.build_candidate(SOURCE)
        self.assertEqual(again, self.data)
        self.assertEqual(manifest, self.manifest)
        demo, evidence = self.builder.make_demo(again, manifest)
        self.assertEqual(demo, self.demo)
        self.assertEqual(evidence, self.evidence)
        with self.assertRaises(ValueError):
            self.builder.write_products(SOURCE, SOURCE, OUT / 'field-position-manifest.json', OUT / 'synthetic-demo.docx')
        self.assertEqual(SOURCE.read_bytes(), before)
        with self.assertRaises(ValueError):
            self.builder.build_candidate(SOURCE, expected_sha='0'*64)

    def test_media_fixed_layers_and_untouched_parts(self):
        self.assertEqual(set(self.package), set(self.original))
        for name in self.original:
            if not (name == 'word/document.xml' or name.startswith('word/charts/')):
                self.assertEqual(self.package[name], self.original[name], name)
        old = xml(self.original['word/document.xml'])
        new = xml(self.package['word/document.xml'])
        old_layers = old.findall('.//wpg:grpSp/wps:wsp', NS)
        new_layers = [a.find('a:graphic/a:graphicData/wps:wsp', NS)
                      for a in new.findall('.//wp:anchor', NS)
                      if a.find('wp:docPr', NS).get('title', '').startswith('fixed.grade.')]
        self.assertEqual(len(old_layers), len(new_layers), 'fixed grade/background lost')
        for before, after in zip(old_layers, new_layers, strict=True):
            before = copy.deepcopy(before); after = copy.deepcopy(after)
            # Only the coordinate transform changes from nested to native EMU.
            for node in (before, after):
                props = node.find('wps:spPr', NS)
                props.remove(props.find('a:xfrm', NS))
            self.assertEqual(ET.tostring(before), ET.tostring(after))
        self.assertEqual(len(new.findall('.//w:tbl', NS)), 20)

    def test_graph_shapes_styles_and_hierarchy_preservation(self):
        self.assertEqual({c['key'] for c in self.manifest['charts']}, KEYS)
        for chart in self.manifest['charts']:
            root = xml(self.package[chart['part']])
            series = root.findall('.//c:ser', NS)
            count = 13 if chart['key'].endswith('comparison') else 2
            self.assertEqual(len(series), 2 if count == 13 else 1)
            for ser in series:
                pts = ser.findall('c:val/c:numLit/c:pt', NS)
                self.assertEqual([int(p.get('idx')) for p in pts], list(range(count)))
            self.assertEqual(self.builder.chart_style_fingerprint(root),
                             self.builder.chart_style_fingerprint(xml(self.original[chart['part']])))
            if count == 2:
                source = xml(self.original[chart['part']])
                old = source.find('.//c:dLbl/c:tx/c:rich', NS)
                doc = xml(self.package['word/document.xml'])
                label = next(a for a in doc.findall('.//wp:anchor', NS)
                             if a.find('wp:docPr', NS).get('title') == chart['key']+'.numeric-label')
                body = copy.deepcopy(label.find('.//wps:bodyPr', NS)); body.tag = '{'+NS['a']+'}bodyPr'
                original_body = copy.deepcopy(old.find('a:bodyPr', NS))
                for node in (original_body, body):
                    for inset in ('lIns', 'rIns', 'tIns', 'bIns'):
                        node.attrib.pop(inset, None)
                self.assertEqual(ET.tostring(original_body), ET.tostring(body),
                                 'only the four approved label insets may change')
                default = old.find('a:p/a:pPr/a:defRPr', NS)
                source_sizes = {str(int(x.get('sz', default.get('sz')))//50)
                                for x in old.findall('a:p/a:r/a:rPr', NS)}
                source_colors = {x.get('val') for x in old.findall('.//a:solidFill/a:srgbClr', NS)}
                source_fonts = {x.get('typeface') for x in old.iter()
                                if x.tag in {'{'+NS['a']+'}'+font for font in ('latin','ea','cs')}}
                for props in label.findall('.//w:rPr', NS):
                    self.assertTrue(set(props.find('w:rFonts', NS).attrib.values()) <= source_fonts)
                    self.assertIn(props.find('w:sz', NS).get('{'+NS['w']+'}val'), source_sizes)
                    color = props.find('w:color', NS)
                    if color.get('{'+NS['w']+'}themeColor') is not None:
                        self.assertIn(color.get('{'+NS['w']+'}themeColor'),
                                      [{'tx1':'text1','tx2':'text2'}.get(x.get('val'), x.get('val'))
                                       for x in old.findall('.//a:solidFill/a:schemeClr', NS)])
                    else:
                        self.assertIn(color.get('{'+NS['w']+'}val'), source_colors)
        levels = self.manifest['category_hierarchy']
        self.assertEqual(len(levels), 2)
        self.assertEqual([p['idx'] for p in levels[1]], [0, 3, 7, 10])
        self.assertEqual(len(levels[0]), 13)

    def test_approved_layout_whitelist_preserves_data_styles_and_other_paragraph_properties(self):
        doc = xml(self.package['word/document.xml'])
        source = xml(self.original['word/document.xml'])
        source_paths = self.builder.paths(source)
        allowed = set(source.findall('.//w:tbl', NS)[3].findall('.//w:tc/w:p', NS))
        allowed.update(p for p in source.findall('w:body/w:p', NS)
                       if self.builder.wtext(p).startswith('【诊断】'))
        tables = source.findall('.//w:tbl',NS)
        allowed.update(p for table in tables[5:18]+tables[19:20] for p in table.findall('w:tr/w:tc/w:p',NS))
        body = list(source.find('w:body',NS))
        advice_start = next(i for i,p in enumerate(body) if self.builder.wtext(p)=='★ 各维度发展建议')
        for i,p in enumerate(body):
            text = self.builder.wtext(p)
            if text.startswith(('管理特质详细分析','★ 管理特质综合表现','呈现作答者','◆ ')) or text in (
                    '★ 总体发展建议','★ 各维度发展建议'):
                allowed.add(p)
            if i > advice_start and text in [d[1] for d in self.builder.DIMENSIONS]:
                allowed.update((p,body[i+1]))
        for p in source.findall('w:body/w:p', NS)+source.findall('.//w:tbl/w:tr/w:tc/w:p', NS):
            path = '.'+source_paths[p].removeprefix('/w:document[1]')
            old = copy.deepcopy(p.find('w:pPr', NS))
            new = copy.deepcopy(doc.find(path, NS).find('w:pPr', NS))
            if p in allowed:
                for props in (old,new):
                    for name in ('keepLines','keepNext','pageBreakBefore'):
                        node = props.find('w:'+name, NS)
                        if node is not None:
                            props.remove(node)
            self.assertEqual(None if old is None else ET.tostring(old),
                             None if new is None else ET.tostring(new), path)
        for chart in self.manifest['charts']:
            root = xml(self.package[chart['part']])
            expected = xml(self.original[chart['part']])
            self.builder.materialize_chart(expected, {})
            for path in ('.//c:ser/c:val', './/c:ser/c:cat', './/c:ser/c:tx'):
                self.assertEqual([ET.tostring(x) for x in root.findall(path,NS)],
                                 [ET.tostring(x) for x in expected.findall(path,NS)])
            if chart['type'] == 'doughnut':
                label = next(a for a in doc.findall('.//wp:anchor', NS)
                             if a.find('wp:docPr', NS).get('title') == chart['key']+'.numeric-label')
                bounds = chart['source_label_bounds_points']
                for i,(axis,coord) in enumerate((('cx','x'),('cy','y'))):
                    scale = Fraction(int(chart['anchor_extent_emu'][axis]),1)/Fraction(bounds[4])
                    old_size = round(Fraction(bounds[i+2])*scale)
                    old_offset = int(chart['source_inline_origin_emu'][coord])+int(chart['anchor_offset_emu'][coord])+round(Fraction(bounds[i])*scale)
                    position = int(label.find('wp:position'+('H' if coord=='x' else 'V')+'/wp:posOffset',NS).text)
                    size = int(label.find('wp:extent',NS).get(axis))
                    dx,dy = {'chart.overall':(0,108),'chart.module.self':(-96,0),
                             'chart.module.interpersonal':(-96,0),'chart.module.task':(100,0),
                             'chart.module.development':(100,0)}[chart['key']]
                    shift = (dx if coord=='x' else dy)*12700
                    self.assertLessEqual(abs(2*position+size-(2*old_offset+old_size+2*shift)),1)
                    self.assertGreater(size,old_size)
                self.assertTrue(all(label.find('.//wps:bodyPr',NS).get(k)=='0'
                                    for k in ('lIns','rIns','tIns','bIns')))
            else:
                layout = root.find('c:chart/c:legend/c:layout/c:manualLayout',NS)
                self.assertEqual({axis:layout.find('c:'+axis,NS).get('val') for axis in ('x','y','w','h')},
                                 {'x':'0.66','y':'0.88','w':'0.33','h':'0.09'})
                mutation = copy.deepcopy(root)
                font = mutation.find('c:chart/c:legend/c:txPr/a:p/a:pPr/a:defRPr',NS)
                font.set('sz','800')
                self.assertNotEqual(self.builder.chart_style_fingerprint(root),
                                    self.builder.chart_style_fingerprint(mutation))

    def test_local_flow_flags_stop_at_each_module_and_non_scope_rows_unchanged(self):
        source = xml(self.original['word/document.xml'])
        doc = xml(self.package['word/document.xml'])
        for table_i,(before,after) in enumerate(zip(source.findall('.//w:tbl',NS),
                                                   doc.findall('.//w:tbl',NS),strict=True)):
            for row_i,(old,row) in enumerate(zip(before.findall('w:tr',NS),after.findall('w:tr',NS),strict=True)):
                old_props = copy.deepcopy(old.find('w:trPr',NS))
                props = copy.deepcopy(row.find('w:trPr',NS))
                if 5 <= table_i < 18 or table_i == 19:
                    self.assertEqual(props.find('w:cantSplit',NS).get('{'+NS['w']+'}val'),'1')
                    if 5 <= table_i < 18:
                        for p in row.findall('w:tc/w:p',NS):
                            self.assertEqual(p.find('w:pPr/w:keepNext',NS).get('{'+NS['w']+'}val'),
                                             '1' if row_i<2 else '0')
                    for node in (old_props,props):
                        if node is not None and node.find('w:cantSplit',NS) is not None:
                            node.remove(node.find('w:cantSplit',NS))
                self.assertEqual(None if old_props is None else ET.tostring(old_props),
                                 None if props is None or (old_props is None and len(props)==0 and not props.attrib)
                                 else ET.tostring(props))

    def test_fields_locations_cardinality_and_labels(self):
        doc = xml(self.package['word/document.xml'])
        fields = self.manifest['fields']
        tags = [x.get('{'+NS['w']+'}val') for x in doc.findall('.//w:tag', NS)]
        self.assertEqual(set(tags), set(fields))
        self.assertEqual(len(tags), len(fields))
        self.assertEqual(len(fields), 6 + 2 + 65 + 12 + 3)
        for key in self.builder.DIMENSIONS:
            for suffix in ('score', 'level', 'norm', 'diagnosis', 'advice'):
                self.assertIn('dimension.'+key[0]+'.'+suffix, fields)
        for tag, locations in fields.items():
            self.assertEqual(len(locations), 1)
            self.assertTrue(locations[0]['source_xpath'])
            self.assertTrue(locations[0]['candidate_xpath'])
            source_path = '.' + locations[0]['source_xpath'].removeprefix('/w:document[1]')
            candidate_path = '.' + locations[0]['candidate_xpath'].removeprefix('/w:document[1]')
            original_p = xml(self.original['word/document.xml']).find(source_path, NS)
            candidate_sdt = doc.find(candidate_path, NS)
            self.assertIsNotNone(original_p, tag)
            self.assertEqual(candidate_sdt.find('w:sdtPr/w:tag', NS).get('{'+NS['w']+'}val'), tag)
            lo, hi = locations[0]['source_char_range']
            self.assertEqual(self.builder.wtext(original_p)[lo:hi], locations[0]['sample'])
        self.assertEqual([f for f in fields if f.startswith('overall.advice.')],
                         ['overall.advice.1', 'overall.advice.2', 'overall.advice.3'])
        self.assertTrue(self.manifest['pending'])
        for sdt in doc.findall('.//w:sdt', NS):
            text = ''.join(t.text or '' for t in sdt.findall('.//w:t', NS))
            self.assertFalse(any(label in text for label in ('得分：', '等级：', '常模分：', '【等级】', '【诊断】')))

    def test_summary_bold_name_plain_body(self):
        doc = xml(self.package['word/document.xml'])
        for sdt in doc.findall('.//w:sdt', NS):
            tag = sdt.find('w:sdtPr/w:tag', NS).get('{'+NS['w']+'}val')
            if tag.startswith('overview.'):
                run = sdt.find('w:sdtContent/w:r', NS)
                self.assertEqual(run.find('w:rPr/w:b', NS).get('{'+NS['w']+'}val'),
                                 '1' if tag.endswith('.name') else '0')

    def test_demo_fields_graph_data_and_richtext_numeric_slots_agree(self):
        package = parts(io.BytesIO(self.demo))
        doc = xml(package['word/document.xml'])
        self.assertTrue(self.evidence['synthetic'])
        for sdt in doc.findall('.//w:sdt', NS):
            tag = sdt.find('w:sdtPr/w:tag', NS).get('{'+NS['w']+'}val')
            actual = ''.join(t.text or '' for t in sdt.findall('.//w:t', NS))
            self.assertEqual(actual, self.evidence['values'][tag])
        for chart in self.manifest['charts']:
            root = xml(package[chart['part']])
            expected = self.evidence['chart_values'][chart['key']]
            for ser, values in zip(root.findall('.//c:ser', NS), expected, strict=True):
                self.assertEqual([p.find('c:v', NS).text for p in ser.findall('c:val/c:numLit/c:pt', NS)], values)
            for slot in chart['numeric_text_slots']:
                self.assertEqual(doc.find(slot['xpath'], NS).text, self.evidence['values'][slot['field']])
        self.assertIn('synthetic', self.evidence['values']['participant.name'])

    def test_incomplete_cache_and_invalid_demo_inputs_rejected(self):
        chart = xml(self.original['word/charts/chart6.xml'])
        cache = chart.find('.//c:numCache', NS)
        cache.remove(cache.find('c:pt', NS))
        with self.assertRaises(ValueError):
            self.builder.materialize_chart(chart, {})
        snapshot = copy.deepcopy(self.manifest)
        self.builder.make_demo(self.data, self.manifest)
        self.assertEqual(snapshot, self.manifest)

    def test_zero_quarter_three_quarter_full_scores_keep_exact_data_and_slot(self):
        for score in (0, 25, 75, 100, 28.85):
            with self.subTest(score=score):
                package, doc = self.boundary_fixture(score)
                for chart in self.manifest['charts']:
                    if chart['type'] != 'doughnut':
                        continue
                    values = xml(package[chart['part']]).findall('.//c:val/c:numLit/c:pt/c:v', NS)
                    self.assertEqual([v.text for v in values], [str(score),str(100-score)])
                    slot = chart['numeric_text_slots'][0]
                    self.assertEqual(doc.find(slot['xpath'], NS).text, f'{score:.2f}')
                    self.assertEqual(slot['part'], 'word/document.xml')

    @classmethod
    def boundary_fixture(cls, score):
        package = parts(io.BytesIO(cls.demo))
        doc = xml(package['word/document.xml'])
        for chart in cls.manifest['charts']:
            if chart['type'] != 'doughnut':
                continue
            root = xml(package[chart['part']])
            values = root.findall('.//c:val/c:numLit/c:pt/c:v', NS)
            values[0].text = str(score); values[1].text = str(100-score)
            doc.find(chart['numeric_text_slots'][0]['xpath'], NS).text = f'{score:.2f}'
            package[chart['part']] = cls.builder.serialize(root, package[chart['part']])
        package['word/document.xml'] = cls.builder.serialize(doc, package['word/document.xml'])
        return package, doc


class SourceLayoutContract(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        spec = importlib.util.spec_from_file_location('source_layout_builder',
                    ROOT / 'scripts/tools/build-management-traits-002-word-candidate.py')
        cls.builder = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(cls.builder)
        cls.original = parts(SOURCE)

    def build(self):
        return self.builder.build_candidate(SOURCE, source_layout=True)

    def projection(self, document):
        """Exhaustive document comparison, excluding only the two promoted groups.

        SDT wrappers and text-run splits are binding-only changes. All other
        nodes/attributes/text (including stars and layout/style properties) stay.
        """
        root = copy.deepcopy(document)
        for parent in list(root.iter()):
            for child in list(parent):
                if child.tag == self.builder.q('w:sdt'):
                    index = list(parent).index(child)
                    parent.remove(child)
                    for node in list(child.find('w:sdtContent', NS)):
                        parent.insert(index, node); index += 1
        mc = '{http://schemas.openxmlformats.org/markup-compatibility/2006}AlternateContent'
        for parent in list(root.iter()):
            for child in list(parent):
                if child.tag == mc and child.findall('.//wpg:graphicFrame//c:chart', NS):
                    parent.remove(child)
                elif child.tag == self.builder.q('w:drawing'):
                    prop = child.find('.//wp:docPr', NS)
                    title = prop.get('title', '') if prop is not None else ''
                    if prop is not None and (prop.get('id') in ('67','20') or
                            title in KEYS or title.startswith('fixed.grade.') or
                            title.endswith('.numeric-label')):
                        parent.remove(child)
        # Coalesce only pure text runs with identical attributes and rPr.
        # This permits binding splits but rejects any style/layout/text change.
        for parent in list(root.iter()):
            previous = None; signature = None
            for node in list(parent):
                pure = node.tag == self.builder.q('w:r') and all(
                    x.tag in (self.builder.q('w:rPr'),self.builder.q('w:t')) for x in node)
                if not pure:
                    previous = None; signature = None; continue
                props = node.find('w:rPr', NS)
                current = (tuple(sorted(node.attrib.items())),
                           ET.tostring(props) if props is not None else None)
                text = self.builder.wtext(node)
                for x in list(node):
                    if x.tag == self.builder.q('w:t'):
                        node.remove(x)
                ET.SubElement(node,self.builder.q('w:t')).text = text
                if previous is not None and current == signature:
                    previous.find('w:t',NS).text = self.builder.wtext(previous)+text
                    parent.remove(node)
                else:
                    previous = node; signature = current
        return ET.tostring(root)

    def test_cli_source_layout_flag_exists(self):
        result = subprocess.run([sys.executable,'-B',str(ROOT /
            'scripts/tools/build-management-traits-002-word-candidate.py'),'--help'],
            capture_output=True,check=True)
        self.assertIn(b'--source-layout',result.stdout)

    def test_entire_document_text_styles_layout_and_stars_preserved(self):
        data, manifest = self.build()
        doc = xml(parts(io.BytesIO(data))['word/document.xml'])
        source = xml(self.original['word/document.xml'])
        self.assertEqual(self.projection(doc),self.projection(source))
        # Prove the whitelist does not hide pPr/rPr/text/star mutations.
        for path, attribute, value in [('.//w:pPr','test','changed'),
                                        ('.//w:rPr','test','changed'),
                                        ('.//w:tblPr','test','changed'),
                                        ('.//w:trPr','test','changed'),
                                        ('.//w:tcPr','test','changed')]:
            mutated = copy.deepcopy(doc); mutated.find(path,NS).set(attribute,value)
            self.assertNotEqual(self.projection(mutated),self.projection(source),path)
        mutated = copy.deepcopy(doc)
        star = next(t for t in mutated.findall('.//w:t',NS) if '★' in (t.text or ''))
        star.text = star.text.replace('★','')
        self.assertNotEqual(self.projection(mutated),self.projection(source))

    def test_all_package_parts_exhaustive_allowlist_and_footer_font_media_bytes(self):
        data, manifest = self.build(); package = parts(io.BytesIO(data))
        allowed = {'word/document.xml'} | {f'word/charts/chart{i}.xml' for i in range(1,7)} | {
            f'word/charts/_rels/chart{i}.xml.rels' for i in range(1,7)}
        self.assertEqual(set(package),set(self.original))
        self.assertEqual(set(manifest['changed_parts']),allowed)
        for name, original in self.original.items():
            if name not in allowed:
                self.assertEqual(package[name],original,name)
            elif name.endswith('.rels'):
                expected = xml(original)
                for relationship in list(expected):
                    if relationship.get('TargetMode') == 'External':
                        expected.remove(relationship)
                self.assertEqual(ET.tostring(xml(package[name])),ET.tostring(expected),name)

    def test_six_charts_exact_styles_data_except_literal_and_label_conversion(self):
        data, manifest = self.build(); package = parts(io.BytesIO(data))
        for item in manifest['charts']:
            expected = xml(self.original[item['part']])
            self.builder.materialize_chart(expected,{})
            if item['type'] == 'doughnut':
                self.builder.detach_ring_label(expected,item['key'])
                expected.find('c:chart',NS).remove(expected.find('c:chart/c:title',NS))
                deleted = expected.find('c:chart/c:autoTitleDeleted',NS)
                if deleted is not None:
                    deleted.set('val','1')
            self.assertEqual(ET.tostring(xml(package[item['part']])),ET.tostring(expected),item['part'])

    def test_numeric_labels_original_bounds_insets_fonts_and_sample_text(self):
        data, manifest = self.build(); doc = xml(parts(io.BytesIO(data))['word/document.xml'])
        for item in manifest['charts']:
            if item['type'] != 'doughnut':
                continue
            label = next(a for a in doc.findall('.//wp:anchor',NS)
                if a.find('wp:docPr',NS).get('title') == item['key']+'.numeric-label')
            bounds = self.builder.SOURCE_LABEL_BOUNDS[item['key']]
            for i,(axis,coord) in enumerate((('cx','x'),('cy','y'))):
                scale = Fraction(item['anchor_extent_emu'][axis])/Fraction(bounds[4])
                self.assertEqual(int(label.find('wp:extent',NS).get(axis)),round(Fraction(bounds[i+2])*scale))
                position = int(item['source_inline_origin_emu'][coord])+int(item['anchor_offset_emu'][coord])+round(Fraction(bounds[i])*scale)
                self.assertEqual(int(label.find('wp:position'+('H' if coord=='x' else 'V')+'/wp:posOffset',NS).text),position)
            rich = xml(self.original[item['part']]).find('.//c:dLbl/c:tx/c:rich',NS)
            old_body = copy.deepcopy(rich.find('a:bodyPr',NS)); old_body.tag = self.builder.q('wps:bodyPr')
            self.assertEqual(ET.tostring(label.find('.//wps:bodyPr',NS)),ET.tostring(old_body))
            self.assertEqual(self.builder.wtext(label),''.join(t.text or '' for t in rich.findall('.//a:t',NS)))
            slot = item['numeric_text_slots'][0]
            sample = re.search(r'\d+\.\d+',''.join(t.text or '' for t in rich.findall('.//a:t',NS))).group()
            self.assertEqual(doc.find(slot['xpath'],NS).text,sample)
            # DrawingML -> Word units only, with original source font/color/b/i.
            source_p = rich.find('a:p',NS); default = source_p.find('a:pPr/a:defRPr',NS)
            def style(run):
                direct = run.find('a:rPr',NS)
                attrs = dict(default.attrib); attrs.update(direct.attrib)
                fonts = []
                for kind in ('latin','ea','cs'):
                    font = direct.find('a:'+kind,NS)
                    if font is None:
                        font = default.find('a:'+kind,NS)
                    fonts.append(font.get('typeface') if font is not None else None)
                fill = direct.find('a:solidFill',NS)
                if fill is None:
                    fill = default.find('a:solidFill',NS)
                return attrs,fonts,ET.tostring(fill)
            source_styles = [style(r) for r in source_p.findall('a:r',NS)]
            for props in label.findall('.//w:rPr',NS):
                def matches(source_style):
                    attrs, fonts, fill = source_style
                    actual_fonts = props.find('w:rFonts',NS)
                    if [actual_fonts.get(self.builder.q('w:'+k)) for k in ('ascii','eastAsia','cs')] != fonts:
                        return False
                    if actual_fonts.get(self.builder.q('w:hAnsi')) != fonts[0]:
                        return False
                    for k in ('b','i'):
                        value = props.find('w:'+k,NS)
                        if (value.get(self.builder.q('w:val')) if value is not None else None) != attrs.get(k):
                            return False
                    for size in ('sz','szCs'):
                        if props.find('w:'+size,NS).get(self.builder.q('w:val')) != str(int(attrs['sz'])//50):
                            return False
                    source_color = xml(fill)[0]
                    color = props.find('w:color',NS)
                    if source_color.tag == self.builder.q('a:srgbClr'):
                        if color.get(self.builder.q('w:val')) != source_color.get('val'):
                            return False
                    elif color.get(self.builder.q('w:themeColor')) != {
                            'tx1':'text1','tx2':'text2'}.get(source_color.get('val'),source_color.get('val')):
                        return False
                    language = props.find('w:lang',NS)
                    return (language.get(self.builder.q('w:val')) if language is not None else None) == attrs.get('lang')
                self.assertTrue(any(matches(s) for s in source_styles),'source label style lost')

    def test_structural_binding_literal_and_coordinates_contract(self):
        data, manifest = self.build()
        original_args = ARGS.docx
        try:
            ARGS.docx = io.BytesIO(data)
            for method in ('test_actual_dynamic_sdt_slots','test_six_native_anchor_business_charts',
                'test_all_opc_literal_only_no_external_relationships',
                'test_visual_restore_native_anchors_keep_source_frame_coordinates',
                'test_visual_restore_fixed_background_shapes_are_native_and_co_located'):
                case = PackageContract(method); case.setUp(); getattr(case,method)()
        finally:
            ARGS.docx = original_args
        self.assertEqual(len(manifest['fields']),88)
        self.assertEqual(sum(len(c['numeric_text_slots']) for c in manifest['charts']),5)
        self.assertEqual({c['key'] for c in manifest['charts']},KEYS)

    def test_determinism_default_unchanged_and_source_target_rejected(self):
        data, manifest = self.build(); again, repeated = self.build()
        self.assertEqual(data,again); self.assertEqual(manifest,repeated)
        self.assertEqual(hashlib.sha256(SOURCE.read_bytes()).hexdigest(),SHA)
        self.assertEqual(self.builder.build_candidate(SOURCE)[0],
                         (OUT/'management-traits-002-candidate.docx').read_bytes())
        with self.assertRaises(ValueError):
            self.builder.write_products(SOURCE,SOURCE,OUT/'source-layout-manifest.json',
                OUT/'source-layout-synthetic-demo.docx',source_layout=True)
        with self.assertRaises(ValueError):
            self.builder.write_products(SOURCE,OUT/'management-traits-002-candidate.docx',
                OUT/'field-position-manifest.json',OUT/'management-traits-002-synthetic-demo.docx',
                source_layout=True)
        with self.assertRaises(ValueError):
            self.builder.build_candidate(SOURCE,expected_sha='0'*64,source_layout=True)

    def test_manifest_exact_provenance_and_no_runtime_claim(self):
        data, manifest = self.build()
        self.assertEqual(manifest['layout_mode'],'source-layout')
        self.assertEqual(manifest['source_path'],str(SOURCE.resolve()))
        self.assertEqual(manifest['source_bytes'],660351)
        self.assertEqual(manifest['source_sha256'],SHA)
        self.assertEqual(manifest['candidate_sha256'],hashlib.sha256(data).hexdigest())
        self.assertFalse(manifest['runtime_connected'])
        self.assertFalse(manifest['attachment_byte_equality_verified'])
        self.assertTrue(manifest['sample_values_preserved_not_scoring_gold_standard'])
        for filename,key in [('management-traits-002-source-layout-template.docx','template'),
                             ('source-layout-manifest.json','manifest'),
                             ('source-layout-synthetic-demo.docx','demo')]:
            self.assertEqual(manifest['output_paths'][key],str((OUT/filename).resolve()))

    def test_demo_all_bound_values_and_source_layout_unchanged(self):
        data, manifest = self.build(); demo, evidence = self.builder.make_demo(data,manifest)
        again, repeated = self.builder.make_demo(data,manifest)
        self.assertEqual(demo,again); self.assertEqual(evidence,repeated)
        package = parts(io.BytesIO(demo)); doc = xml(package['word/document.xml'])
        for sdt in doc.findall('.//w:sdt',NS):
            tag = sdt.find('w:sdtPr/w:tag',NS).get(self.builder.q('w:val'))
            self.assertEqual(self.builder.wtext(sdt),evidence['values'][tag])
        for item in manifest['charts']:
            chart = xml(package[item['part']])
            for series,values in zip(chart.findall('.//c:ser',NS),evidence['chart_values'][item['key']],strict=True):
                self.assertEqual([v.text for v in series.findall('c:val/c:numLit/c:pt/c:v',NS)],values)
            for slot in item['numeric_text_slots']:
                self.assertEqual(doc.find(slot['xpath'],NS).text,evidence['values'][slot['field']])
        blank = xml(parts(io.BytesIO(data))['word/document.xml'])
        for path in ('.//w:pPr','.//w:tblPr','.//w:trPr','.//w:tcPr','.//w:rPr','.//wps:bodyPr'):
            self.assertEqual([ET.tostring(x) for x in doc.findall(path,NS)],
                             [ET.tostring(x) for x in blank.findall(path,NS)],path)
        self.assertTrue(evidence['synthetic']); self.assertTrue(evidence['no_scoring_gold_standard'])

    def test_fixed_background_layers_and_each_field_source_position(self):
        data, manifest = self.build(); package = parts(io.BytesIO(data))
        case = BuildAndDemoContract('test_media_fixed_layers_and_untouched_parts')
        case.builder = self.builder; case.package = package
        case.original = self.original; case.manifest = manifest
        case.test_media_fixed_layers_and_untouched_parts()
        case.test_fields_locations_cardinality_and_labels()


class ActualPDFVisualContract(BuildAndDemoContract):
    rendered = {}
    output_dir = None

    def pdf_path(self, name):
        return (self.output_dir or OUT) / (name+'.pdf')

    def pdf_pages(self, name, data):
        directory = self.output_dir or OUT
        directory.mkdir(parents=True, exist_ok=True)
        cache_key = (str(directory), name, hashlib.sha256(data).hexdigest())
        if cache_key in self.rendered:
            return self.rendered[cache_key]
        docx = directory / (name+'.docx')
        docx.write_bytes(data)
        lo = subprocess.run([str(ARGS.soffice), '-env:UserInstallation='+
                             (directory/('lo-'+uuid.uuid4().hex[:8])).resolve().as_uri(),
                             '--headless', '--convert-to', 'pdf', '--outdir', str(directory), str(docx)],
                            capture_output=True, timeout=60)
        self.assertEqual(lo.returncode, 0, lo.stderr.decode('utf-8', errors='replace'))
        pdf = docx.with_suffix('.pdf')
        self.assertTrue(pdf.exists(),lo.stdout.decode('utf-8',errors='replace')+' '+
                lo.stderr.decode('utf-8',errors='replace'))
        self.assertGreaterEqual(pdf.stat().st_mtime, docx.stat().st_mtime,
                                'stale PDF must never count as fresh conversion evidence')
        result = subprocess.run(['pdftotext', '-bbox-layout', str(pdf), '-'],
                                capture_output=True, check=True, timeout=15)
        pages = [x for x in ET.fromstring(result.stdout).iter() if x.tag.endswith('}page')]
        print(f'PDF_PAGES {name}={len(pages)}')
        self.rendered[cache_key] = pages
        return pages

    @staticmethod
    def page_words(page):
        return [(x.text or '',float(x.get('xMin')),float(x.get('yMin')),
                 float(x.get('xMax')),float(x.get('yMax')))
                for x in page.iter() if x.tag.endswith('}word')]

    @staticmethod
    def intersects(a, b):
        return a[0] < b[2] and b[0] < a[2] and a[1] < b[3] and b[1] < a[3]

    def test_actual_pdf_ring_labels_clear_real_ring_envelopes_and_plot(self):
        # A real 100% PDF supplies complete annulus outer envelopes, not
        # guessed OOXML boxes or text-presence-only evidence. All cases retain
        # the same five chart matrices (separately checked by PackageContract).
        full = self.builder.pack(self.boundary_fixture(100)[0])
        self.pdf_pages('boundary-100', full)
        svg = (self.output_dir or OUT) / 'ring-envelope-100.svg'
        subprocess.run(['pdftocairo','-svg','-f','3','-l','3',
                str(self.pdf_path('boundary-100')),str(svg)], check=True, timeout=20)
        rings = []
        for path in xml(svg.read_bytes()).iter():
            if not path.tag.endswith('}path') or path.get('fill') is None:
                continue
            if path.get('fill') in ('none', 'rgb(100%, 100%, 100%)'):
                continue
            if 'C' not in path.get('d', ''):
                continue
            self.assertIsNone(path.get('transform'), 'unhandled ring path transform')
            coords = [float(x) for x in re.findall(r'-?\d+(?:\.\d+)?', path.get('d'))]
            box = (min(coords[::2]), min(coords[1::2]), max(coords[::2]), max(coords[1::2]))
            if 180 < box[1] < 370 and box[2]-box[0] > 70:
                rings.append(box)
        self.assertEqual(len(rings), 5, 'actual SVG must expose all five complete rings')
        for score in (0,25,75,100,28.85):
            pages = self.pdf_pages(f'boundary-{score}', self.builder.pack(self.boundary_fixture(score)[0]))
            lines = [x for x in pages[2].iter() if x.tag.endswith('}line')]
            labels = []
            for line in lines:
                words = self.page_words(line)
                if any(name in ''.join(w[0] for w in words) for name in self.builder.LABEL_KEYS):
                    labels.append((min(w[1] for w in words), min(w[2] for w in words),
                                   max(w[3] for w in words), max(w[4] for w in words)))
            self.assertEqual(len(labels), 5)
            for i, label in enumerate(labels):
                with self.subTest(score=score, label=i):
                    self.assertFalse(any(self.intersects(label, ring) for ring in rings),
                                     f'label intersects real ring envelope: {label}; rings={rings}')
                    self.assertFalse(any(self.intersects(label, other) for other in labels[i+1:]))
                    # Plot identity is its 13 leaf category axis / original
                    # full-score bars, not an arbitrary body-density threshold.
                    plot_words = [w for w in self.page_words(pages[2]) if w[2] > 380 and
                                  (w[0] in [d[1] for d in self.builder.DIMENSIONS] or w[0]=='100.00')]
                    self.assertTrue(plot_words)
                    self.assertLess(label[3], min(w[2] for w in plot_words))
            print(f'PDF_RING_CLEAR score={score} labels=5 rings=5')

    def test_actual_pdf_label_pixels_have_no_ring_colored_ink(self):
        for score in (0,25,75,100,28.85):
            pages = self.pdf_pages(f'boundary-{score}',self.builder.pack(self.boundary_fixture(score)[0]))
            result = subprocess.run(['pdftoppm','-r','144','-f','3','-l','3','-singlefile',
                                     str(self.pdf_path(f'boundary-{score}'))],capture_output=True,check=True,timeout=20)
            image = io.BytesIO(result.stdout)
            tokens = []
            while len(tokens)<4:
                line = image.readline()
                if not line.startswith(b'#'):
                    tokens.extend(line.split())
            self.assertEqual(tokens[0],b'P6')
            width,height,maximum = map(int,tokens[1:])
            self.assertEqual(maximum,255)
            pixels = image.read()
            self.assertEqual(len(pixels),width*height*3)
            # Actual full-ring SVG uses RGB(72,116,203). Label glyphs retain
            # the distinct source RGB(47,84,150); no dependency/image guessing.
            checked = 0
            for line in pages[2].iter():
                if not line.tag.endswith('}line'):
                    continue
                words = self.page_words(line)
                if not any(name in ''.join(w[0] for w in words) for name in self.builder.LABEL_KEYS):
                    continue
                x0,y0 = int(min(w[1] for w in words)*2),int(min(w[2] for w in words)*2)
                x1,y1 = int(max(w[3] for w in words)*2)+1,int(max(w[4] for w in words)*2)+1
                count = 0
                for y in range(y0,y1):
                    row = pixels[(y*width+x0)*3:(y*width+x1)*3]
                    count += sum(abs(row[i]-72)<=2 and abs(row[i+1]-116)<=2 and abs(row[i+2]-203)<=2
                                 for i in range(0,len(row),3))
                with self.subTest(score=score,label=checked):
                    self.assertEqual(count,0,'real ring pixels intersect label bbox')
                checked += 1
            self.assertEqual(checked,5)
            print(f'PDF_PIXEL_CLEAR score={score} labels=5 ring_ink=0')

    def paragraph_location(self, pages, text, start_page=1, left=None, x_range=None, start_position=None):
        normalize = lambda s: re.sub(r'\s+', '', s)
        target = normalize(text)
        stream = ''; locations = []
        for i, page in enumerate(pages, 1):
            if i < start_page:
                continue
            for word in sorted(self.page_words(page),key=lambda w:(round(w[2],0),w[1])):
                if not 45 < word[2] < 790 or (left is not None and (word[1] < 297) != left):
                    continue
                if x_range is not None and not x_range[0] <= word[1] < x_range[1]:
                    continue
                if start_position is not None and (i,word[2]) < start_position:
                    continue
                value = normalize(word[0]); stream += value; locations.extend([i]*len(value))
        offset = stream.find(target)
        self.assertGreaterEqual(offset, 0, 'paragraph missing: '+target[:70])
        return set(locations[offset:offset+len(target)])

    def test_actual_pdf_no_summary_only_page_by_content_identity(self):
        for name, data in [('management-traits-002-candidate',self.data),
                           ('management-traits-002-synthetic-demo',self.demo)]:
            pages = self.pdf_pages(name, data)
            doc = xml(parts(io.BytesIO(data))['word/document.xml'])
            cells = doc.findall('.//w:tbl',NS)[3].findall('w:tr/w:tc',NS)
            for left, cell in zip((True,False),cells,strict=True):
                for p in cell.findall('w:p',NS):
                    where = self.paragraph_location(pages,self.builder.wtext(p),left=left)
                    self.assertEqual(len(where),1,'summary paragraph split across pages')
                    page = pages[next(iter(where))-1]
                    text = re.sub(r'\s+','',''.join(w[0] for w in self.page_words(page)))
                    identities = ('综合均值','管理特质详细分析','自信心得分：')
                    self.assertTrue(any(identity in text for identity in identities),
                                    f'{name}: summary-only orphan page {where}')

    def test_actual_pdf_all_13_detail_blocks_and_advice_stay_local(self):
        for name,data in [('management-traits-002-candidate',self.data),
                          ('management-traits-002-synthetic-demo',self.demo)]:
            pages = self.pdf_pages(name,data)
            doc = xml(parts(io.BytesIO(data))['word/document.xml'])
            detail_start = next(i for i,p in enumerate(pages,1)
                                if 'Detailed' in ''.join(w[0] for w in self.page_words(p)))
            tables = doc.findall('.//w:tbl',NS)
            for table, dimension in zip(tables[5:18],self.builder.DIMENSIONS,strict=True):
                with self.subTest(document=name,dimension=dimension[0]):
                    titles = [(i,w) for i,page in enumerate(pages,1) if i >= detail_start
                              for w in self.page_words(page) if w[0]==dimension[1] and
                              45<w[1]<55 and 45<w[2]<790]
                    self.assertTrue(titles,'exact detailed table title missing')
                    title_page,title_word = titles[0]
                    spans = [{title_page}]
                    for row_i,row in enumerate(table.findall('w:tr',NS)):
                        for col_i,cell in enumerate(row.findall('w:tc',NS)):
                            if col_i == 0:
                                continue  # Repeated static row labels are not paragraph identities.
                            region = (0,120) if col_i == 0 else ((460,596) if col_i == 2 else
                                      ((120,460) if row_i == 0 else (120,596)))
                            for p in cell.findall('w:p',NS):
                                if self.builder.wtext(p):
                                    spans.append(self.paragraph_location(pages,self.builder.wtext(p),detail_start,
                                                    x_range=region,start_position=(title_page,title_word[2]-2)))
                    self.assertEqual(len(set.union(*spans)),1,'title/score/definition/diagnosis split')
            body = list(doc.find('w:body',NS))
            advice_start = next(i for i,p in enumerate(pages,1)
                                if 'Development' in ''.join(w[0] for w in self.page_words(p)))
            for i,p in enumerate(body):
                if p.tag == self.builder.q('w:p') and self.builder.wtext(p) in [d[1] for d in self.builder.DIMENSIONS]:
                    heading = self.paragraph_location(pages,self.builder.wtext(p),advice_start)
                    advice = self.paragraph_location(pages,self.builder.wtext(body[i+1]),advice_start)
                    self.assertEqual(len(heading | advice),1,'advice heading/complete paragraph split')
            print(f'PDF_LOCAL_BLOCKS_CHECKED {name} details=13 advice=13')

    def test_actual_pdf_all_five_numeric_labels_are_unbroken_and_not_clipped(self):
        # 28.85 is label-only stress data, not a formal scoring fixture.
        cases = [(score,self.builder.pack(self.boundary_fixture(score)[0])) for score in (0,25,75,100,28.85)]
        for score, data in cases:
            with self.subTest(score=score):
                pages = self.pdf_pages(f'boundary-{score}', data)
                self.assertGreaterEqual(len(pages), 3)
                words = self.page_words(pages[2])
                # All chart labels are on the restored overview page, before the
                # grade/chart area. Checking indivisible PDF tokens catches the
                # previous 28.85 -> "2" + "8.85" wrapping, not just text presence.
                overview = [x for x in words if 180 < x[2] < 360]
                full = [x for x in overview if re.search(r'(?<!\d)'+re.escape(f'{score:.2f}')+r'(?!\d)',x[0])]
                print(f'PDF_BOUNDARY score={score} full_numeric_tokens={len(full)} expected=5')
                self.assertEqual(len(full), 5, 'score token missing, wrapped or clipped in actual PDF')
                for name in ('综合均值','自我管理均值','人际管理均值','任务管理均值','发展管理均值'):
                    self.assertTrue(any(name in x[0] for x in overview), 'incomplete label: '+name)

    def test_actual_pdf_legend_complete_and_separate_from_full_score_labels(self):
        pages = self.pdf_pages('legend-full-score', self.builder.pack(self.boundary_fixture(100)[0]))
        words = self.page_words(pages[2])
        legend = [x for x in words if '常模' in x[0] or '参照' in x[0]]
        self.assertIn('常模参照分', ''.join(x[0] for x in legend), 'legend truncated in real PDF')
        numbers = [x for x in words if '100.00' in x[0] and x[2] > 360]
        self.assertGreaterEqual(len(numbers), 2)
        for label in legend:
            for number in numbers:
                self.assertFalse(label[1] < number[3] and number[1] < label[3] and
                                 label[2] < number[4] and number[2] < label[4],
                                 'legend overlaps a full-score data label')

    def test_actual_pdf_long_demo_all_36_texts_complete_no_orphan_page(self):
        pages = self.pdf_pages('management-traits-002-synthetic-demo', self.demo)
        normalize = lambda text: re.sub(r'\s+', '', text)
        text = normalize(''.join(x[0] for page in pages for x in self.page_words(page)))
        columns = [normalize(''.join(x[0] for page in pages for x in self.page_words(page)
                        if 40 < x[2] < 760 and (x[1] < 297) == left))
               for left in (True, False)]
        values = self.evidence['values']
        keys = [key for key in values if key.endswith(('.diagnosis','.advice','.text')) or
                key.startswith('overall.advice.')]
        self.assertEqual(len(keys), 36)
        for key in keys:
            with self.subTest(field=key):
                self.assertTrue(any(normalize(values[key]) in stream for stream in [text]+columns),
                                'long text lost or clipped: '+key)
        for index, page in enumerate(pages, 1):
            # Exclude header/footer by position, never waive the orphan page.
            content = normalize(''.join(x[0] for x in self.page_words(page) if 75 < x[2] < 760))
            self.assertGreater(len(content), 10, f'orphan/empty body on physical page {index}')


class LibreOfficeCompatibleContract(SourceLayoutContract):
    def build(self):
        return self.builder.build_candidate(SOURCE, libreoffice_compatible=True)

    def test_cli_source_layout_flag_exists(self):
        result = subprocess.run([sys.executable,'-B',str(ROOT /
            'scripts/tools/build-management-traits-002-word-candidate.py'),'--help'],
            capture_output=True,check=True)
        self.assertIn(b'--libreoffice-compatible',result.stdout)

    def test_entire_document_text_styles_layout_and_stars_preserved(self):
        data, manifest = self.build()
        doc = xml(parts(io.BytesIO(data))['word/document.xml'])
        source = xml(self.original['word/document.xml'])
        # Only explicit local page-flow flags are normalized; whole-document
        # source projection still checks every other node, text and rPr.
        def projected(document):
            clone = copy.deepcopy(document)
            baseline = copy.deepcopy(source)
            self.builder.local_page_flow(baseline)
            summary_paths = {self.builder.paths(source)[p] for p in source.findall('.//w:tbl',NS)[3].findall('.//w:tc/w:p',NS)}
            old_paths = self.builder.paths(source)
            nodes = source.findall('w:body/w:p',NS)+source.findall('.//w:tbl/w:tr/w:tc/w:p',NS)+source.findall('.//w:tbl/w:tr',NS)
            for old in nodes:
                path = '.'+old_paths[old].removeprefix('/w:document[1]')
                expected = baseline.find(path,NS)
                actual = clone.find(path,NS)
                if old.tag not in (self.builder.q('w:p'),self.builder.q('w:tr')):
                    continue
                prop_tag = 'w:pPr' if old.tag == self.builder.q('w:p') else 'w:trPr'
                old_prop = old.find(prop_tag,NS)
                expected_prop = expected.find(prop_tag,NS)
                for flag in ('keepLines','keepNext','pageBreakBefore','cantSplit'):
                    a = old_prop.find('w:'+flag,NS) if old_prop is not None else None
                    b = expected_prop.find('w:'+flag,NS) if expected_prop is not None else None
                    summary_flag = old_paths[old] in summary_paths and flag in ('keepLines','keepNext')
                    diagnosis_flag = self.builder.wtext(old).startswith('【诊断】') and flag == 'keepLines'
                    if not (summary_flag or diagnosis_flag) and (ET.tostring(a) if a is not None else None) == (ET.tostring(b) if b is not None else None):
                        continue
                    props = actual.find(prop_tag,NS)
                    node = props.find('w:'+flag,NS) if props is not None else None
                    if node is not None:
                        props.remove(node)
                    if old_prop is None and props is not None and not len(props) and not props.attrib:
                        actual.remove(props)
            return self.projection(clone)
        self.assertEqual(projected(doc),projected(source))
        for path in ('.//w:rPr','.//w:tblPr','.//w:tcPr','.//w:pPr'):
            mutation = copy.deepcopy(doc); mutation.find(path,NS).set('mutation','forbidden')
            self.assertNotEqual(projected(mutation),projected(source),path)
        mutation = copy.deepcopy(doc)
        star = next(t for t in mutation.findall('.//w:t',NS) if '★' in (t.text or ''))
        star.text = star.text.replace('★','')
        self.assertNotEqual(projected(mutation),projected(source))

    def test_six_charts_exact_styles_data_except_literal_and_label_conversion(self):
        data, manifest = self.build(); package = parts(io.BytesIO(data))
        strict, _ = self.builder.build_candidate(SOURCE,source_layout=True)
        strict_parts = parts(io.BytesIO(strict))
        for item in manifest['charts']:
            actual = xml(package[item['part']]); expected = xml(strict_parts[item['part']])
            if item['type'] == 'bar+line':
                path = 'c:chart/c:legend/c:layout/c:manualLayout'
                layout = expected.find(path,NS)
                for axis,value in {'x':'0.66','y':'0.88','w':'0.33','h':'0.09'}.items():
                    layout.find('c:'+axis,NS).set('val',value)
            self.assertEqual(ET.tostring(actual),ET.tostring(expected),item['part'])

    def test_numeric_labels_original_bounds_insets_fonts_and_sample_text(self):
        data, manifest = self.build(); doc = xml(parts(io.BytesIO(data))['word/document.xml'])
        strict, _ = self.builder.build_candidate(SOURCE,source_layout=True)
        strict_doc = xml(parts(io.BytesIO(strict))['word/document.xml'])
        labels = lambda root: {a.find('wp:docPr',NS).get('title'):a for a in root.findall('.//wp:anchor',NS)
            if a.find('wp:docPr',NS).get('title','').endswith('.numeric-label')}
        for title,label in labels(doc).items():
            before = labels(strict_doc)[title]
            self.assertEqual(self.builder.wtext(label),self.builder.wtext(before))
            self.assertEqual([ET.tostring(x) for x in label.findall('.//w:rPr',NS)],
                             [ET.tostring(x) for x in before.findall('.//w:rPr',NS)])
        self.assertEqual(doc.find(next(c['numeric_text_slots'][0]['xpath'] for c in manifest['charts']
            if c['key']=='chart.overall'),NS).text,'59.45')
        # Reuse the tested exact size/offset/inset/legend and local-flow checks.
        case = BuildAndDemoContract('test_approved_layout_whitelist_preserves_data_styles_and_other_paragraph_properties')
        case.builder = self.builder; case.package = parts(io.BytesIO(data))
        case.original = self.original; case.manifest = manifest
        case.test_approved_layout_whitelist_preserves_data_styles_and_other_paragraph_properties()
        case.test_local_flow_flags_stop_at_each_module_and_non_scope_rows_unchanged()
        for cell in doc.findall('.//w:tbl',NS)[3].findall('w:tr/w:tc',NS):
            paragraphs = cell.findall('w:p',NS)
            for index,p in enumerate(paragraphs):
                self.assertEqual(p.find('w:pPr/w:keepLines',NS).get(self.builder.q('w:val')),'1')
                if index < len(paragraphs)-1:
                    self.assertEqual(p.find('w:pPr/w:keepNext',NS).get(self.builder.q('w:val')),'1')

    def test_manifest_exact_provenance_and_no_runtime_claim(self):
        data, manifest = self.build()
        self.assertEqual(manifest['layout_mode'],'libreoffice-compatible')
        self.assertEqual(manifest['source_sha256'],SHA)
        self.assertEqual(manifest['source_bytes'],660351)
        self.assertFalse(manifest['runtime_connected'])
        self.assertFalse(manifest['attachment_byte_equality_verified'])
        self.assertTrue(manifest['sample_values_preserved_not_scoring_gold_standard'])
        self.assertEqual(manifest['candidate_sha256'],hashlib.sha256(data).hexdigest())
        self.assertEqual([Path(manifest['output_paths'][k]).name for k in ('template','manifest','demo')],
            ['management-traits-002-lo-compatible-template.docx','lo-compatible-manifest.json',
             'lo-compatible-synthetic-demo.docx'])

    def test_determinism_default_unchanged_and_source_target_rejected(self):
        data, manifest = self.build(); again, repeated = self.build()
        self.assertEqual(data,again); self.assertEqual(manifest,repeated)
        default, _ = self.builder.build_candidate(SOURCE)
        strict, _ = self.builder.build_candidate(SOURCE,source_layout=True)
        self.assertEqual(hashlib.sha256(default).hexdigest(),
            '1f236407713143d0e945d60391e9179bc761c62f0a42e750b50075c98f029273')
        self.assertEqual(hashlib.sha256(strict).hexdigest(),
            '465358b1ad88bb9442029c2c03609ec1816c64e2d8e9b3021019859a521ff471')
        protected = [SOURCE,OUT/'management-traits-002-candidate.docx',
                     OUT/'management-traits-002-source-layout-template.docx']
        for target in protected:
            with self.assertRaises(ValueError):
                self.builder.write_products(SOURCE,target,OUT/'lo-compatible-manifest.json',
                    OUT/'lo-compatible-synthetic-demo.docx',libreoffice_compatible=True)
        with self.assertRaises(ValueError):
            self.builder.build_candidate(SOURCE,source_layout=True,libreoffice_compatible=True)
        with self.assertRaises(ValueError):
            self.builder.build_candidate(SOURCE,expected_sha='0'*64,libreoffice_compatible=True)


class LibreOfficeCompatiblePDFContract(ActualPDFVisualContract):
    output_dir = OUT/'lo-compatible-evidence'
    rendered = {}

    @classmethod
    def setUpClass(cls):
        SourceLayoutContract.setUpClass.__func__(cls)
        cls.data, cls.manifest = cls.builder.build_candidate(SOURCE,source_layout=True) if ARGS.strict_visual_red else (
            cls.builder.build_candidate(SOURCE,libreoffice_compatible=True))
        cls.package = parts(io.BytesIO(cls.data))
        cls.demo, cls.evidence = cls.builder.make_demo(cls.data,cls.manifest)
        cls.output_dir = OUT/('lo-compatible-strict-red' if ARGS.strict_visual_red else 'lo-compatible-evidence')


class EvidenceTestResult(unittest.TextTestResult):
    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.passed_tests = []
        self.passed_subtests = 0

    def addSuccess(self, test):
        super().addSuccess(test)
        self.passed_tests.append(test.id())

    def addSubTest(self, test, subtest, error):
        super().addSubTest(test,subtest,error)
        if error is None:
            self.passed_subtests += 1


def compatible_evidence(result, protected, started):
    """Explicit offline evidence generation; never export/save through Word."""
    regressions = {}
    ActualPDFVisualContract.output_dir = OUT/'lo-compatible-optimized-regression'
    for name in ('source-layout','optimized'):
        suite = unittest.defaultTestLoader.loadTestsFromTestCase(
            SourceLayoutContract if name=='source-layout' else PackageContract)
        if name=='optimized':
            suite.addTests(unittest.defaultTestLoader.loadTestsFromTestCase(BuildAndDemoContract))
            suite.addTests(ActualPDFVisualContract(method) for method in dir(ActualPDFVisualContract)
                           if method.startswith('test_actual_pdf_'))
        regression = unittest.TextTestRunner(verbosity=1,resultclass=EvidenceTestResult).run(suite)
        regressions[name] = {'tests':regression.testsRun,'passed':len(regression.passed_tests),
            'failures':len(regression.failures),'errors':len(regression.errors),'skipped':len(regression.skipped),
            'passed_subtests':regression.passed_subtests,'test_ids':regression.passed_tests}
        if not regression.wasSuccessful():
            raise AssertionError(name+' regression failed; evidence cannot claim completion')
    scripts = [ROOT/'scripts/tools/build-management-traits-002-word-candidate.py',Path(__file__)]
    syntax = {}
    for script in scripts:
        text = script.read_text(encoding='utf-8')
        ast.parse(text); compile(text,str(script),'exec')
        syntax[str(script.relative_to(ROOT))] = hashlib.sha256(script.read_bytes()).hexdigest()
    LibreOfficeCompatiblePDFContract.setUpClass()
    case = LibreOfficeCompatiblePDFContract('test_actual_pdf_long_demo_all_36_texts_complete_no_orphan_page')
    case.output_dir = OUT
    builder = case.builder
    manifest = builder.write_products(SOURCE,OUT/'management-traits-002-lo-compatible-template.docx',
        OUT/'lo-compatible-manifest.json',OUT/'lo-compatible-synthetic-demo.docx',libreoffice_compatible=True)
    names = ['management-traits-002-lo-compatible-template','lo-compatible-synthetic-demo']
    pdf_info = {}
    for name in names:
        docx = OUT/(name+'.docx')
        pages = case.pdf_pages(name,docx.read_bytes())
        pdf = case.pdf_path(name)
        pdf_info[name] = {'pages':len(pages),'sha256':builder.sha(pdf.read_bytes()),'bytes':pdf.stat().st_size}
        subprocess.run(['pdftoppm','-png','-r','72',str(pdf),str(OUT/(name+'-review'))],check=True,timeout=60)
    # Installed Windows drawing API only; no Pillow or new dependency.
    sheet_script = r'''
Add-Type -AssemblyName System.Drawing
$dir=$args[0]
foreach($name in @('management-traits-002-lo-compatible-template','lo-compatible-synthetic-demo')) {
    $files=@(Get-ChildItem -LiteralPath $dir -Filter ($name+'-review-*.png') | Sort-Object Name)
    $width=250; $height=370; $cols=3; $rows=[int][Math]::Ceiling($files.Count/3)
    $sheet=New-Object System.Drawing.Bitmap ($cols*$width),($rows*$height)
    $g=[System.Drawing.Graphics]::FromImage($sheet)
    try {
        $g.Clear([System.Drawing.Color]::White)
        $g.InterpolationMode=[System.Drawing.Drawing2D.InterpolationMode]::HighQualityBicubic
        for($i=0;$i -lt $files.Count;$i++) {
            $image=[System.Drawing.Image]::FromFile($files[$i].FullName)
            try { $g.DrawImage($image,([int]($i%3)*$width),([int][Math]::Floor($i/3)*$height),$width,$height) }
            finally { $image.Dispose() }
        }
        $sheet.Save((Join-Path $dir ($name+'-contactsheet.png')),[System.Drawing.Imaging.ImageFormat]::Png)
    } finally { $g.Dispose(); $sheet.Dispose() }
}
'''
    # Pass the directory as a literal, not an untrusted interpolated command.
    escaped_dir = str(OUT.resolve()).replace("'","''")
    subprocess.run(['powershell','-NoProfile','-Command',
        '& { '+sheet_script+" } '"+escaped_dir+"'"],check=True,capture_output=True,timeout=60)
    word_script = r'''
$dir=$args[0]; $word=$null; $results=@()
try {
    $word=New-Object -ComObject Word.Application
    $word.Visible=$false; $word.DisplayAlerts=0
    foreach($name in @('management-traits-002-lo-compatible-template.docx','lo-compatible-synthetic-demo.docx')) {
        $path=Join-Path $dir $name; $before=(Get-FileHash -LiteralPath $path).Hash; $doc=$null
        try {
            $doc=$word.Documents.Open($path,$false,$true)
            $results+=@{name=$name; pages=$doc.ComputeStatistics(2); controls=$doc.ContentControls.Count;
                read_only=$doc.ReadOnly; word_version=$word.Version}
        } finally { if($doc) {$doc.Close(0); [void][Runtime.InteropServices.Marshal]::ReleaseComObject($doc)} }
        $results[-1].sha_unchanged=($before -eq (Get-FileHash -LiteralPath $path).Hash)
    }
} finally { if($word) {$word.Quit(); [void][Runtime.InteropServices.Marshal]::ReleaseComObject($word)} }
ConvertTo-Json -Depth 5 -InputObject @($results) -Compress
'''
    word = subprocess.run(['powershell','-NoProfile','-Command',
        '& { '+word_script+" } '"+escaped_dir+"'"],capture_output=True,check=True,timeout=120)
    word_info = json.loads(word.stdout.decode('utf-8-sig'))
    if len(word_info)!=2 or any(not x['read_only'] or not x['sha_unchanged'] or x['controls']!=88 for x in word_info):
        raise AssertionError('read-only Word opening/control/hash contract failed')
    current = {name:builder.sha((OUT/name).read_bytes()) for name in protected}
    if current != protected:
        raise AssertionError('historical product changed during isolated validation')
    evidence = {
        'schema':'management-traits-002-lo-compatible-validation-v1','local_only':True,
        'started_utc':started,'ended_utc':datetime.now(timezone.utc).isoformat(),
         'red':{'kind':'observed pre-implementation baseline from this 2026-10-02 task, not replayed with current builder',
             'third_mode':{'tests':10,'failures':1,'errors':9,'exit':1},
               'strict_source_actual_pdf':{'tests':7,'failures':57,'errors':0,'skipped':0,'exit':1}},
        'green':{'tests':result.testsRun,'passed':len(result.passed_tests),'failures':len(result.failures),
                 'errors':len(result.errors),'skipped':len(result.skipped),'passed_subtests':result.passed_subtests,
                 'test_ids':result.passed_tests},
        'old_mode_regressions':regressions,'ast_compile_verified_script_sha256':syntax,
        'commands':{'compatible':str(Path(__file__).relative_to(ROOT))+' --libreoffice-compatible --visual-boundaries --write-evidence',
                    'source_layout':str(Path(__file__).relative_to(ROOT))+' --source-layout-contract',
                    'optimized':str(Path(__file__).relative_to(ROOT))+' --visual-boundaries --isolated-regression'},
        'environment_issue':{'isolated_optimized_initial_failures':11,
            'cause_evidence':'long profile produced no PDF; same DOCX with short isolated profile converted successfully',
            'resolution':'short unique LO profile names; full optimized 23-test suite rerun, no assertions weakened'},
        'source':{'sha256':builder.sha(SOURCE.read_bytes()),'bytes':SOURCE.stat().st_size,
                  'attachment_full_byte_equality_verified':False,'converted_source':False},
        'new_docx_sha256':{'template':manifest['candidate_sha256'],'demo':manifest['synthetic_demo']['demo_sha256']},
        'protected_existing_root_products':{'count':len(protected),'before':protected,'after':current,'unchanged':True},
        'contracts':{'SDT':88,'numeric_slots':5,'charts':6,'parts_added':0,'parts_removed':0,
                     'changed_parts':manifest['changed_parts'],'media_sha256':manifest['media_sha256'],
                     'source_text_rPr_stars_footer_preserved':True,'source_sample_overall':'59.45',
                     'summary_name_body_rPr':'source exact; no forced bold/plain',
                     'data_style_box_legend_pageflow':'exhaustive source projection with explicit local whitelist',
                     'boundary_scores':[0,25,75,100,28.85],'unbroken_numeric_tokens':[5,5,5,5,5],
                     'ring_envelope_checks':25,'label_ring_pixel_checks':25,'ring_pixel_intersections':0,
                     'complete_long_texts':36,'detail_blocks_same_page_per_doc':13,
                     'advice_blocks_same_page_per_doc':13,'summary_only_pages':0},
        'pdf':pdf_info,'word_readonly':word_info,
        'libreoffice_version':subprocess.run([str(ARGS.soffice),'--version'],capture_output=True,check=True).stdout.decode('utf-8',errors='replace').strip(),
        'remaining':['final visual/customer acceptance not closed','target server PDF not tested',
                     'Word PDF/pagination parity not tested; read-only opening only',
                     'visible: overview page lower area and detail page 7 remain sparse; no global compression authorized',
                     'visible: source footer includes total pages; fixed sample stars remain, neither revised',
                     'original footer/stars/content/sample inconsistencies retained, not corrected or approved',
                     'runtime integration/activation/DB/deployment absent'],
        'review_images':[str(OUT/(name+'-contactsheet.png')) for name in names],
        'evidence_directory':str(LibreOfficeCompatiblePDFContract.output_dir),
        'primary_agent_followup':'maintain docs/project-memory and regression/business-branch ledgers; do not treat local contracts as approval'}
    path = OUT/'lo-compatible-validation.json'
    path.write_text(json.dumps(evidence,ensure_ascii=False,indent=2)+'\n',encoding='utf-8')
    print('COMPATIBLE_EVIDENCE='+str(path))
    print('WORD_READONLY='+json.dumps(word_info))


def main():
    global ARGS
    parser = argparse.ArgumentParser()
    parser.add_argument('--docx', type=Path, default=OUT / 'management-traits-002-candidate.docx')
    parser.add_argument('--source-red', action='store_true')
    parser.add_argument('--source-layout-contract', action='store_true')
    parser.add_argument('--libreoffice-compatible', action='store_true')
    parser.add_argument('--strict-visual-red', action='store_true')
    parser.add_argument('--isolated-regression', action='store_true')
    parser.add_argument('--write-evidence', action='store_true',help='compatible mode only: independent products/PDF/PNG/Word read-only evidence')
    parser.add_argument('--visual-boundaries', action='store_true')
    parser.add_argument('--soffice', type=Path, default=Path('C:/Program Files/LibreOffice/program/soffice.com'))
    ARGS = parser.parse_args()
    if ARGS.write_evidence and (not ARGS.libreoffice_compatible or not ARGS.visual_boundaries or ARGS.strict_visual_red):
        parser.error('--write-evidence requires compatible actual PDF tests, not strict RED')
    started = datetime.now(timezone.utc).isoformat()
    protected = {p.name:hashlib.sha256(p.read_bytes()).hexdigest() for p in OUT.iterdir() if p.is_file()
        and not (p.name.startswith(('management-traits-002-lo-compatible','lo-compatible-')))}
    visual_methods = [name for name in dir(ActualPDFVisualContract) if name.startswith('test_actual_pdf_')]
    if ARGS.libreoffice_compatible or ARGS.strict_visual_red:
        suite = unittest.TestSuite()
        if not ARGS.strict_visual_red:
            suite.addTests(unittest.defaultTestLoader.loadTestsFromTestCase(LibreOfficeCompatibleContract))
        if ARGS.visual_boundaries or ARGS.strict_visual_red:
            suite.addTests(LibreOfficeCompatiblePDFContract(name) for name in visual_methods)
    else:
        suite = unittest.defaultTestLoader.loadTestsFromTestCase(
            SourceLayoutContract if ARGS.source_layout_contract else PackageContract)
    if not (ARGS.source_red or ARGS.source_layout_contract or ARGS.libreoffice_compatible or ARGS.strict_visual_red):
        suite.addTests(unittest.defaultTestLoader.loadTestsFromTestCase(BuildAndDemoContract))
        if ARGS.visual_boundaries:
            if ARGS.isolated_regression:
                ActualPDFVisualContract.output_dir = OUT/'lo-compatible-optimized-regression'
            suite.addTest(ActualPDFVisualContract('test_actual_pdf_all_five_numeric_labels_are_unbroken_and_not_clipped'))
            suite.addTest(ActualPDFVisualContract('test_actual_pdf_legend_complete_and_separate_from_full_score_labels'))
            suite.addTest(ActualPDFVisualContract('test_actual_pdf_long_demo_all_36_texts_complete_no_orphan_page'))
            suite.addTest(ActualPDFVisualContract('test_actual_pdf_ring_labels_clear_real_ring_envelopes_and_plot'))
            suite.addTest(ActualPDFVisualContract('test_actual_pdf_label_pixels_have_no_ring_colored_ink'))
            suite.addTest(ActualPDFVisualContract('test_actual_pdf_no_summary_only_page_by_content_identity'))
            suite.addTest(ActualPDFVisualContract('test_actual_pdf_all_13_detail_blocks_and_advice_stay_local'))
    result = unittest.TextTestRunner(verbosity=2,resultclass=EvidenceTestResult).run(suite)
    print('CONTRACT_TESTS=%d FAILED=%d ERRORS=%d SKIPPED=%d' %
          (result.testsRun, len(result.failures), len(result.errors), len(result.skipped)))
    if ARGS.write_evidence and result.wasSuccessful():
        compatible_evidence(result,protected,started)
    return 0 if result.wasSuccessful() else 1


if __name__ == '__main__':
    sys.exit(main())