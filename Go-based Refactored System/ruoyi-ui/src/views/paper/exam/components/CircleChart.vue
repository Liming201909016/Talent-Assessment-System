<template>
  <div>
    <h1>good</h1>
    <div ref="chart" class="pdf-group-item" style="width: 100%; height: 200px;"></div>
  </div>
</template>

<script>
import * as echarts from 'echarts';

export default {
  name: 'CircleChart',
  data(){
    return {
      inner_radius: 11,
      inner_left: 200,
      inner_top: 60,
      inner_color: '#FFA500',
      outer_radius: 15,
      out_color: '#FFA500',
      outer_line_width: 2,

      rect_width: 60,
      rect_height: 20,
      rect_span:3,
      first_rect_left: 100,
      first_rect_top: 100,
    }
  },
  mounted() {
    // 基于准备好的dom，初始化echarts实例
    let myChart = echarts.init(this.$refs.chart, null, { devicePixelRatio: 2 });

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
              fill: '#90EE90' // 浅绿色填充
            }
          },
          {
            type: 'rect',
            left: this.first_rect_left + this.rect_width + this.rect_span,
            top: this.first_rect_top,
            shape: {
              width: this.rect_width,
              height: this.rect_height,
            },
            style: {
              fill: '#ADD8E6' // 浅蓝色
            }
          },
          {
            type: 'rect',
            left: this.first_rect_left + (this.rect_width + this.rect_span)*2,
            top: this.first_rect_top,
            shape: {
              width: this.rect_width,
              height: this.rect_height,
            },
            style: {
              fill: '#FFFF00' // 黄色填充
            }
          },
          {
            type: 'rect',
            left: this.first_rect_left + (this.rect_width + this.rect_span)*3,
            top: this.first_rect_top,
            shape: {
              width: this.rect_width,
              height: this.rect_height,
            },
            style: {
              fill: '#FFC0CB' // 红色
            }
          },
          {
            type: 'rect',
            left: this.first_rect_left + (this.rect_width + this.rect_span)*4,
            top: this.first_rect_top,
            shape: {
              width: this.rect_width,
              height: this.rect_height,
            },
            style: {
              fill: '#8B0000' // 深红色
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
              stroke: this.out_color, // 外层圆的边框颜色
              fill: '#00000000', // 外层圆透明色
              lineWidth: this.outer_line_width // 线宽为2
            }
          },
          {
            type: 'text',
            left: this.inner_left - 4,
            top: this.inner_top - 22,
            style: {
              text: '恐惧',
              fill: '#FFA500',
              font: 'bold 14px Microsoft YaHei'
            }
          }
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
