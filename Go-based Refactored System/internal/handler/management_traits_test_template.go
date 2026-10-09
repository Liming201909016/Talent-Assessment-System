package handler

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"

	"github.com/talent-assessment/refactored/internal/service"
)

const ManagementTraitsTestReportTitle = service.ManagementTraitsTestPurposeTitle
const ManagementTraitsTestReportLabel = service.ManagementTraitsTestPurposeLabel

// BuildManagementTraitsTestTemplate derives a separate test-only package.
// Fixed customer text, media and chart styles are unchanged; two predefined
// purpose SDTs are added outside the unsupported DrawingML title Choice.
func BuildManagementTraitsTestTemplate(candidate []byte) ([]byte, error) {
	h := sha256.Sum256(candidate)
	if hex.EncodeToString(h[:]) != managementTraitsTestTemplateSHA {
		return nil, errors.New("管理特质候选模板来源不符")
	}
	z, err := zip.NewReader(bytes.NewReader(candidate), int64(len(candidate)))
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	w := zip.NewWriter(&out)
	for _, f := range z.File {
		r, err := f.Open()
		if err != nil {
			return nil, err
		}
		b, err := io.ReadAll(r)
		closeErr := r.Close()
		if err != nil || closeErr != nil {
			return nil, errors.New("读取候选模板失败")
		}
		if f.Name == "word/document.xml" {
			tree, err := mngWordTree(b)
			if err != nil {
				return nil, err
			}
			bodies := tree.all(mngWordNS, "body")
			if len(bodies) != 1 {
				return nil, errors.New("封面结构不符")
			}
			var p *mngWordNode
			for _, c := range bodies[0].children {
				if c.name.Space == mngWordNS && c.name.Local == "p" {
					p = c
					break
				}
			}
			if p == nil || len(p.all(mngWordNS, "t")) != 0 {
				return nil, errors.New("预定义封面标记位置不符")
			}
			marker := `<w:sdt><w:sdtPr><w:tag w:val="report.testTitle"/><w:id w:val="21003001"/></w:sdtPr><w:sdtContent><w:r><w:rPr><w:b/><w:color w:val="C00000"/><w:sz w:val="24"/></w:rPr><w:t>管理特质 TEST 测试报告</w:t></w:r></w:sdtContent></w:sdt><w:r><w:br/></w:r><w:sdt><w:sdtPr><w:tag w:val="report.testLabel"/><w:id w:val="21003002"/></w:sdtPr><w:sdtContent><w:r><w:rPr><w:b/><w:color w:val="C00000"/><w:sz w:val="22"/></w:rPr><w:t>仅供系统测试，不可作为人才决策依据</w:t></w:r></w:sdtContent></w:sdt>`
			// A page-relative floating text box reserves no paragraph height.
			// This is TEST-purpose layout only, not a change to source objects.
			marker = `<w:r><w:rPr><w:sz w:val="21"/></w:rPr><w:drawing><wp:anchor distT="0" distB="0" distL="0" distR="0" simplePos="0" relativeHeight="251658240" behindDoc="0" locked="0" layoutInCell="1" allowOverlap="1"><wp:simplePos x="0" y="0"/><wp:positionH relativeFrom="page"><wp:posOffset>850000</wp:posOffset></wp:positionH><wp:positionV relativeFrom="page"><wp:posOffset>800000</wp:posOffset></wp:positionV><wp:extent cx="5100000" cy="650000"/><wp:effectExtent l="0" t="0" r="0" b="0"/><wp:wrapNone/><wp:docPr id="21003003" name="mng-test-purpose" title="mng-test-purpose"/><wp:cNvGraphicFramePr/><a:graphic><a:graphicData uri="http://schemas.microsoft.com/office/word/2010/wordprocessingShape"><wps:wsp><wps:cNvSpPr txBox="1"/><wps:spPr><a:xfrm><a:off x="0" y="0"/><a:ext cx="5100000" cy="650000"/></a:xfrm><a:prstGeom prst="rect"><a:avLst/></a:prstGeom><a:noFill/><a:ln><a:noFill/></a:ln></wps:spPr><wps:txbx><w:txbxContent><w:p><w:pPr><w:spacing w:before="0" w:after="0"/></w:pPr>` + marker + `</w:p></w:txbxContent></wps:txbx><wps:bodyPr rot="0" lIns="0" tIns="0" rIns="0" bIns="0" anchor="t"><a:noAutofit/></wps:bodyPr></wps:wsp></a:graphicData></a:graphic></wp:anchor></w:drawing></w:r>`
			b, err = mngWordApply(b, []mngWordEdit{{start: p.contentEnd, end: p.contentEnd, value: marker}})
			if err != nil {
				return nil, err
			}
		}
		entry, err := w.CreateHeader(&zip.FileHeader{Name: f.Name, Method: zip.Deflate})
		if err != nil {
			return nil, err
		}
		if _, err := entry.Write(b); err != nil {
			return nil, err
		}
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
