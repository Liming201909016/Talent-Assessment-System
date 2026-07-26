<template>
  <div style="text-align: center;">
<!--    <h1>good</h1>-->
    <!-- FB-051: width 580 → 760 与页面内容同宽 -->
    <!-- FB-055: width 760 → 700，与下方 score-rs-bg 灰底框对齐（减去两边 padding）-->
    <div ref="chart" class="pdf-group-item score-chart-inner" style="width: 700px; height: 100px; margin: 0 auto;"></div>
  </div>
</template>

<script>
import * as echarts from 'echarts';
import el from "element-ui/src/locale/lang/el";

export default {
  name: 'TotalChart',
  props: ['traits', 'score'],
  data(){
    return {
      inner_radius: 14,
      inner_left: 10,
      inner_top: 42,
      inner_color: '#FFA500',
      outer_radius: 15,
      out_color: '#FFA500',
      outer_line_width: 2,
      text_color: '#FFA500',

      // rect_width: 110,
      rect_width: 670, // FB-055: 730 → 670，与灰底框对齐
      // rect_width: 140,
      rect_height: 12,
      rect_span:1,
      first_rect_left: 10,
      // first_rect_left: 170,
      first_rect_top: 50,
    }
  },
  methods:{
    mixColors(color1, color2, factor) {
      // 将颜色值转换为RGB
      let r1 = parseInt(color1.slice(1, 3), 16);
      let g1 = parseInt(color1.slice(3, 5), 16);
      let b1 = parseInt(color1.slice(5, 7), 16);
      let r2 = parseInt(color2.slice(1, 3), 16);
      let g2 = parseInt(color2.slice(3, 5), 16);
      let b2 = parseInt(color2.slice(5, 7), 16);

      // 混合颜色
      let r = Math.round(r1 * (1 - factor) + r2 * factor);
      let g = Math.round(g1 * (1 - factor) + g2 * factor);
      let b = Math.round(b1 * (1 - factor) + b2 * factor);

      // 将RGB值转换回十六进制并格式化
      return `#${((1 << 24) + (r << 16) + (g << 8) + b).toString(16).slice(1)}`;
    },
    getColorFromGradient(position) {
      // 创建渐变色标（与矩形渐变保持一致）
      const colorStops = [
        { offset: 0, color: '#16C063' },    // 绿色
        { offset: 0.33, color: '#4BCEF0' },  // 蓝色
        { offset: 0.66, color: '#FBB400' },  // 黄色
        { offset: 1, color: '#EF3E3E' }      // 红色
      ];

      // 找到当前位置所在的渐变段
      let startStop, endStop;
      for (let i = 0; i < colorStops.length - 1; i++) {
        if (position >= colorStops[i].offset && position <= colorStops[i + 1].offset) {
          startStop = colorStops[i];
          endStop = colorStops[i + 1];
          break;
        }
      }

      // 计算在当前渐变段内的相对位置（0-1）
      const segmentPosition = (position - startStop.offset) / (endStop.offset - startStop.offset);

      // 颜色混合（复用mixColors方法）
      return this.mixColors(startStop.color, endStop.color, segmentPosition);
    },
  },
  mounted() {
    let traits = this.traits
    // 基于准备好的dom，初始化echarts实例
    let myChart = echarts.init(this.$refs.chart, null, { devicePixelRatio: 2 });
    let step = 65  // FB-055: 71 → 65，匹配 rect_width 670 ((670-20)/10)
    this.inner_left = this.inner_left + this.score * step

    if (this.inner_left <= step * 2.5) {
      this.inner_color = '#16C063'
      this.out_color = '#16C063'
      this.text_color = '#16C063'
    } else if (this.inner_left <= step * 5) {
      this.inner_color = '#4BCEF0'
      this.out_color = '#4BCEF0'
      this.text_color = '#4BCEF0'
    } else if (this.inner_left <= step * 7.5) {
      this.inner_color = '#FBB400'
      this.out_color = '#FBB400'
      this.text_color = '#FBB400'
    } else {
      this.inner_color = '#EF3E3E'
      this.out_color = '#EF3E3E'
      this.text_color = '#EF3E3E'
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
              r: 20
            },
            style: {
              // fill: '#16C063' // 浅绿色填充
              fill: new echarts.graphic.LinearGradient(0, 0, 1, 0, [{
                offset: 0, color: '#16C063' // 开始颜色
              }, {
                offset: 0.33, color: '#4BCEF0'
              }, {
                offset: 0.66, color: '#FBB400'
              },{
                offset: 1, color: '#EF3E3E' // 结束颜色
              }])
            }
          },
          {
            type: 'text',
            left: this.first_rect_left,
            top: this.first_rect_top + this.rect_height + this.rect_span * 4 + 10,
            style: {
              text: '0分',
              fill: '#333333',
              fontSize:11,
            }
          },
          {
            type: 'text',
            left: this.first_rect_left + (this.rect_width - 20) / 5 ,
            top: this.first_rect_top + this.rect_height + this.rect_span * 4 + 10,
            style: {
              text: '2分',
              fill: '#333333',
              fontSize:11,
            }
          },
          {
            type: 'text',
            left: this.first_rect_left + (this.rect_width - 20) / 5 * 2 ,
            top: this.first_rect_top + this.rect_height + this.rect_span * 4 + 10,
            style: {
              text: '4分',
              fill: '#333333',
              fontSize:11,
            }
          },
          {
            type: 'text',
            left: this.first_rect_left + (this.rect_width - 20) / 5 * 3 ,
            top: this.first_rect_top + this.rect_height + this.rect_span * 4 + 10,
            style: {
              text: '6分',
              fill: '#333333',
              fontSize:11,
            }
          },
          {
            type: 'text',
            left: this.first_rect_left + (this.rect_width - 20) / 5 * 4 ,
            top: this.first_rect_top + this.rect_height + this.rect_span * 4 + 10,
            style: {
              text: '8分',
              fill: '#333333',
              fontSize:11,
            }
          },
          {
            type: 'text',
            left: this.first_rect_left + (this.rect_width - 20) / 5 * 5 ,
            top: this.first_rect_top + this.rect_height + this.rect_span * 4 + 10,
            style: {
              text: '10分',
              fill: '#333333',
              fontSize:11,
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
              fill: this.getColorFromGradient(this.score/10) // 内层圆的填充色
            }
          },
          {
            type: 'text',
            left: this.inner_left - 4,
            top: this.inner_top - 22,
            style: {
              text: traits,
              fill: '#333333',
              fontSize: 12,
              fontWeight:500
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
