// 导出页面为PDF格式
import html2Canvas from 'html2canvas';
import JsPDF from 'jspdf';

export default {
  install(Vue, options) {
    Vue.prototype.getPdf = function (title, dom) {
      // 注册getPdf方法，传入两个参数，此处使用了promise处理导出后的操作
      /*
      title: 导出文件名
      dom: 需要导出dom的id
       */
      return new Promise((resolve, reject) => {
        html2Canvas(document.querySelector(dom), {
          useCORS: true, // 由于打印时，会访问dom上的一些图片等资源，解决跨域问题！！重要
          allowTaint: true // 允许跨域
        }).then(function (canvas) {
          let contentWidth = canvas.width;
          let contentHeight = canvas.height;
          // 根据A4纸的大小，计算出dom相应比例的尺寸
          let pageHeight = contentWidth / 592.28 * 841.89;
          let leftHeight = contentHeight;
          let position = 0;
          let imgWidth = 595.28;
          // 根据a4比例计算出需要分割的实际dom位置
          let imgHeight = 592.28 / contentWidth * contentHeight;
          // canvas绘图生成image数据，1.0是质量参数
          let pageData = canvas.toDataURL('image/jpeg', 1.0);
          // a4大小
          let PDF = new JsPDF('', 'pt', 'a4');
          // 当内容达到a4纸的高度时，分割，将一整块画幅分割出一页页的a4大小，导出pdf
          if (leftHeight < pageHeight) {
            PDF.addImage(pageData, 'JPEG', 0, 0, imgWidth, imgHeight);
          } else {
            while (leftHeight > 0) {
              PDF.addImage(pageData, 'JPEG', 0, position, imgWidth, imgHeight);
              leftHeight -= pageHeight;
              position -= 841.89;
              if (leftHeight > 0) {
                PDF.addPage();
              }
            }
          }
          // 导出
          PDF.save(title + '.pdf');
          resolve(true);
        })
          .catch(() => {
            reject(false);
          });
      });
    };
  }
};
