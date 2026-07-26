<template>
  <div style="text-align:center;">
    <div ref="chart" class="total-chart-inner pdf-group-item" style="width: 100%; height: 100px;"></div>
  </div>
</template>

<script>
import * as echarts from 'echarts';

export default {
  name: 'TotalChart',
  props: ['count8', 'totalAvg'],
  data(){
    return {
      inner_radius: 14,
      inner_left: 0,
      inner_top: 24,
      inner_color: '#FFA500',
      outer_radius: 15,
      out_color: '#FFA500',
      outer_line_width: 2,

      line_length: 50,
      rect_width: 110,
      rect_height: 16,
      // rect_height: 20,
      rect_span:3,
      first_rect_left: 0,
      // first_rect_left: 100,
      first_rect_top: 30,
      text_width:60,
      topText: '',
      topTextColor: '#FFA500'
    }
  },
  computed:{
    textPos(){
      return (this.rect_width - this.text_width) / 2
    }
  },
  mounted() {
    // 基于准备好的dom，初始化echarts实例
    let myChart = echarts.init(this.$refs.chart, null, { devicePixelRatio: 2 });
    if (this.totalAvg <= 7 && this.count8 <= 1) {
      this.inner_left = 40
      this.inner_color = '#22B633'
      this.out_color = '#22B633'
      // this.topText = '状态良好'
      this.topTextColor = '#22B633'
    } else if (this.totalAvg <= 7 && this.count8 <= 3) {
      this.inner_left = 153
      this.inner_color = '#22B6AA'
      this.out_color = '#22B6AA'
      // this.topText = '状态正常'
      this.topTextColor = '#22B6AA'
    } else if (this.totalAvg <= 7 && this.count8 <= 5) {
      this.inner_left = 266
      this.inner_color = '#F4EA25'
      this.out_color = '#F4EA25'
      // this.topText = '状态一般'
      this.topTextColor = '#F4EA25'
    } else if (this.totalAvg > 7 && this.count8 <= 6) {
      this.inner_left = 379
      this.inner_color = '#F48A25'
      this.out_color = '#F48A25'
      // this.topText = '中度困扰'
      this.topTextColor = '#F48A25'
    } else if (this.totalAvg > 7 || this.count8 > 8) {
      this.inner_left = 492
      this.inner_color = '#F42525'
      this.out_color = '#F42525'
      // this.topText = '重度困扰'
      this.topTextColor = '#F42525'
    }

    // 指定图表的配置项和数据
    let option = {
      graphic: {
        elements: [
          {
            type: 'rect',
            left: this.first_rect_left,
            top: this.first_rect_top,
            shape: {
              width: this.rect_width,
              height: this.rect_height,
            },
            style: {
              fill: '#22B633' // 浅绿色填充
            }
          },
          {
            type: 'text',
            left: this.first_rect_left + this.textPos,
            top: this.first_rect_top + this.rect_height + this.rect_span + 10,
            style: {
              text: '状态良好',
              fill: '#333333',
              textAlign: 'center', // 文本对齐方式
              fontSize: 14
            }
          },
         /* {
            type: 'line',
            // left: this.first_rect_left + this.rect_width + this.rect_span,
            // top: this.first_rect_top,
            shape: {
              x1: this.first_rect_left + this.rect_width + this.rect_span,
              y1: this.first_rect_top + this.rect_height,
              x2: this.first_rect_left + this.rect_width + this.rect_span,
              y2: this.first_rect_top + this.line_length,
            },
            style: {
              fill: '#30c0b4', // 浅蓝色
              stroke: 'blue'
            }
          },*/
          {
            type: 'rect',
            left: this.first_rect_left + this.rect_width + this.rect_span,
            top: this.first_rect_top,
            shape: {
              width: this.rect_width,
              height: this.rect_height,
            },
            style: {
              fill: '#22B6AA' // 浅蓝色
            }
          },
          {
            type: 'text',
            left: this.first_rect_left + this.rect_width + this.rect_span + this.textPos,
            top: this.first_rect_top + this.rect_height + this.rect_span + 10,
            style: {
              text: '状态正常',
              fill: '#333333',
              fontSize: 14
            }
          },
          /*{
            type: 'line',
            // left: this.first_rect_left + (this.rect_width + this.rect_span) * 2,
            // top: this.first_rect_top,
            shape: {
              x1: this.first_rect_left + (this.rect_width + this.rect_span) * 2,
              y1: this.first_rect_top + this.rect_height,
              x2: this.first_rect_left + (this.rect_width + this.rect_span) * 2,
              y2: this.first_rect_top + this.line_length,
            },
            style: {
              fill: '#22B6AA', // 浅蓝色
              stroke: 'blue'
            }
          },*/
          {
            type: 'rect',
            left: this.first_rect_left + (this.rect_width + this.rect_span)*2,
            top: this.first_rect_top,
            shape: {
              width: this.rect_width,
              height: this.rect_height,
            },
            style: {
              fill: '#F4EA25' // 黄色填充
            }
          },
          {
            type: 'text',
            left: this.first_rect_left + (this.rect_width + this.rect_span)*2 + this.textPos,
            top: this.first_rect_top + this.rect_height + this.rect_span + 10,
            style: {
              text: '状态一般',
              fill: '#333333',
              fontSize: 14
            }
          },
          /*{
            type: 'line',
            // left: this.first_rect_left + (this.rect_width + this.rect_span) * 3,
            // top: this.first_rect_top,
            shape: {
              x1: this.first_rect_left + (this.rect_width + this.rect_span) * 3,
              y1: this.first_rect_top + this.rect_height,
              x2: this.first_rect_left + (this.rect_width + this.rect_span) * 3,
              y2: this.first_rect_top + this.line_length,
            },
            style: {
              fill: '#22B6AA', // 浅蓝色
              stroke: 'blue'
            }
          },*/
          {
            type: 'rect',
            left: this.first_rect_left + (this.rect_width + this.rect_span)*3,
            top: this.first_rect_top,
            shape: {
              width: this.rect_width,
              height: this.rect_height,
            },
            style: {
              fill: '#F48A25' // 红色
            }
          },
          {
            type: 'text',
            left: this.first_rect_left + (this.rect_width + this.rect_span)*3 + this.textPos,
            top: this.first_rect_top + this.rect_height + this.rect_span + 10,
            style: {
              text: '中度困扰',
              fill: '#333333',
              fontSize: 14
            }
          },
          /*{
            type: 'line',
            // left: this.first_rect_left + (this.rect_width + this.rect_span) * 4,
            // top: this.first_rect_top,
            shape: {
              x1: this.first_rect_left + (this.rect_width + this.rect_span) * 4,
              y1: this.first_rect_top + this.rect_height,
              x2: this.first_rect_left + (this.rect_width + this.rect_span) * 4,
              y2: this.first_rect_top + this.line_length,
            },
            style: {
              fill: '#22B6AA', // 浅蓝色
              stroke: 'blue'
            }
          },*/
          {
            type: 'rect',
            left: this.first_rect_left + (this.rect_width + this.rect_span)*4,
            top: this.first_rect_top,
            shape: {
              width: this.rect_width,
              height: this.rect_height,
            },
            style: {
              fill: '#F42525' // 深红色
            }
          },
          {
            type: 'text',
            left: this.first_rect_left + (this.rect_width + this.rect_span)*4 + this.textPos,
            top: this.first_rect_top + this.rect_height + this.rect_span + 10,
            style: {
              text: '重度困扰',
              fill: '#333333',
              fontSize: 14
            }
          },
          {
            type: 'circle',
            left: this.inner_left,
            top: this.inner_top,
            shape: {
              r: this.inner_radius // 内层圆的半径，直径为50，所以半径为25
            },
            style: {
              fill: this.inner_color // 内层圆的填充色
            }
          },
          {
            type: 'circle',
            // left: '94.5', //内圆left - (外圆r - 内圆r) - 外圆线粗/2   =100-(15-11)-(3/2)
            left: this.inner_left -(this.outer_radius-this.inner_radius) - this.outer_line_width/2, //内圆left - (外圆r - 内圆r) - 外圆线粗/2   =100-(15-11)-(3/2)
            // top: '54.5', //内圆top -  (外圆r - 内圆r) - 外圆线粗/2    =60-(15-11)-(3/2)
            top: this.inner_top -(this.outer_radius-this.inner_radius) - this.outer_line_width/2, //内圆top -  (外圆r - 内圆r) - 外圆线粗/2    =60-(15-11)-(3/2)
            shape: {
              r: this.outer_radius // 外层圆的半径，比内层圆大
            },
            style: {
              // stroke: this.out_color, // 外层圆的边框颜色
              stroke: '#ffffff', // 外层圆的边框颜色
              fill: '#00000000', // 外层圆透明色
              lineWidth: this.outer_line_width // 线宽为2
            }
          },
          /*{
            type: 'text',
            left: this.inner_left - 25,
            top: this.inner_top - 22,
            style: {
              text: this.topText,
              fill: this.topTextColor,
              font: 'bold 14px Microsoft YaHei'
            }
          }*/
        ]
      }
    };

    // 使用刚指定的配置项和数据显示图表。
    myChart.setOption(option);
  }
}
</script>

<style scoped>

</style>
