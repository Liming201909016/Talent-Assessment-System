// 页面导出为pdf格式
import html2Canvas from 'html2canvas'
import jsPDF from 'jspdf'

const htmlToPdf = {
  // getPdf(title,url) {
  //   html2Canvas(document.querySelector('#pdfDom'), {
  //     allowTaint: false,
  //     taintTest: false,
  //     logging: false,
  //     useCORS: true,
  //     dpi: window.devicePixelRatio*4, //将分辨率提高到特定的DPI 提高四倍
  //     scale:4 //按比例增加分辨率
  //   }).then(canvas=>{
  //     var pdf = new jsPDF('p', 'mm', 'a4');    //A4纸，纵向
  //     var ctx = canvas.getContext('2d'),
  //       a4w = 190, a4h = 277,    //A4大小，210mm x 297mm，四边各保留10mm的边距，显示区域190x277
  //       imgHeight = Math.floor(a4h * canvas.width / a4w),    //按A4显示比例换算一页图像的像素高度
  //       renderedHeight = 0;
  //
  //     while(renderedHeight < canvas.height) {
  //       var page = document.createElement("canvas");
  //       page.width = canvas.width;
  //       page.height = Math.min(imgHeight, canvas.height - renderedHeight);//可能内容不足一页
  //
  //       //用getImageData剪裁指定区域，并画到前面创建的canvas对象中
  //       page.getContext('2d').putImageData(ctx.getImageData(0, renderedHeight, canvas.width, Math.min(imgHeight, canvas.height - renderedHeight)), 0, 0);
  //       pdf.addImage(page.toDataURL('image/jpeg', 1.0), 'JPEG', 10, 10, a4w, Math.min(a4h, a4w * page.height / page.width));    //添加图像到页面，保留10mm边距
  //
  //       renderedHeight += imgHeight;
  //       if(renderedHeight < canvas.height) {
  //         pdf.addPage();//如果后面还有内容，添加一个空页
  //       }
  //       // delete page;
  //     }
  //     //保存文件
  //     pdf.save(title + '.pdf')
  //   })
  // }

  getPdf(isPrint) {
    let title = "你想要的保存pdf文件的名字"
    // 如果这个页面有左右移动,canvas 也要做响应的移动，不然会出现canvas 内容不全
    const xOffset = window.pageXOffset
    // 避免笔下误 灯下黑 统一写
    const A4_WIDTH = 592.28
    // const A4_HEIGHT = 841.89
    const A4_HEIGHT = 880
    let printDom = document.querySelector('#pdfDom')
    // 根据A4的宽高计算DOM页面一页应该对应的高度
    let pageHeight = printDom.offsetWidth / A4_WIDTH * A4_HEIGHT
    // 将所有不允许被截断的元素进行处理
    let wholeNodes = document.querySelectorAll('.whole-node')
    for (let i = 0; i < wholeNodes.length; i++) {
      //1、 判断当前的不可分页元素是否在两页显示
      const topPageNum = Math.ceil((wholeNodes[i].offsetTop) / pageHeight)

      // const bottomPageNum = Math.ceil((wholeNodes[i].offsetTop + wholeNodes[i] + offsetHeight) / pageHeight)
      const bottomPageNum = Math.ceil((wholeNodes[i].offsetTop + wholeNodes[i].offsetHeight) / pageHeight)

      if (topPageNum !== bottomPageNum) {
        //说明该dom会被截断
        // 2、插入空白块使被截断元素下移
        let divParent = wholeNodes[i].parentNode
        let newBlock = document.createElement('div')
        newBlock.className = 'emptyDiv'
        newBlock.style.background = '#fff'

        // 3、计算插入空白块的高度 可以适当流出空间使得内容太靠边，根据自己需求而定
        let _H = topPageNum * pageHeight - wholeNodes[i].offsetTop
        newBlock.style.height = _H + 30 + 'px'
        divParent.insertBefore(newBlock, wholeNodes[i])
      }

      // 以上完成dom层面的分页 可以转为图片进一步处理了

      html2Canvas(printDom, { height: printDom.offsetHeight, width: printDom.offsetWidth, scrollX: -xOffset, allowTaint: true }).then(canvas => {
        //dom 已经转换为canvas 对象，可以将插入的空白块删除了
        let emptyDivs = document.querySelectorAll('.emptyDiv')
        for (let i = 0; i < emptyDivs.length; i++) {
          emptyDivs[i].parentNode.removeChild(emptyDivs[i])
        }
        // 有一点重复的代码
        let contentWidth = canvas.width
        let contentHeight = canvas.height
        let pageHeight = contentWidth / A4_WIDTH * A4_HEIGHT
        let leftHeight = contentHeight
        let position = 0

        let imgWidth = A4_WIDTH
        let imgHeight = A4_WIDTH / contentWidth * contentHeight
        let pageData = canvas.toDataURL('image/jpeg', 1.0)
        // if (isPrint) {
        //   //如果是打印，可以拿着分号页的数据 直接使用
        //   printJs({ printable: pageData, type: 'image', base64: true, documentTitle: '\u200E' })
        //   return
        // }
        //计算分页的pdf
        let PDF = new jsPDF('', 'pt', 'a4')
        if (leftHeight <= pageHeight) {
          PDF.addImage(pageData, 'JPEG', 0, 0, imgWidth, imgHeight)
        } else {
          while (leftHeight > 0) {
            PDF.addImage(pageData, 'JPEG', 0, position, imgWidth, imgHeight)
            leftHeight -= pageHeight
            position -= A4_HEIGHT
            if (leftHeight > 0) {
              PDF.addPage()
            }
          }
        }
        PDF.save(title + '.pdf')
      })
    }
  },
};

export default htmlToPdf;
