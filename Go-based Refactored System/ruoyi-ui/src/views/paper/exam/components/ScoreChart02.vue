<template>
  <div>
<!--    <h1>good</h1>-->
    <div ref="chart" style="width: 100%; height: 90px;">
<!--      class="pdf-group-item"-->
    </div>
  </div>
</template>

<script>
import * as echarts from 'echarts';
import scoreImg from '@/assets/report_images/score.png';
import el from "element-ui/src/locale/lang/el";

export default {
  name: 'TotalChart',
  props: ['traits', 'score'],
  data(){
    return {

      scoreImg: scoreImg,

      inner_radius: 8,
      // inner_left: 20,
      inner_left: 0,
      inner_top: 60,
      inner_color: '#FBB400',
      outer_radius: 12,
      out_color: '#FBB400',
      outer_line_width: 2,
      text_color: '#FBB400',
      text_left: 0,

      rect_width: 110,
      // rect_width: 140,
      rect_height: 40,
      rect_span:1,
      first_rect_left: 20,
      // first_rect_left: 170,
      first_rect_top: 100,
    }
  },
  mounted() {
    let traits = this.traits
    let testScore = 4;
    // 基于准备好的dom，初始化echarts实例
    let myChart = echarts.init(this.$refs.chart, null, { devicePixelRatio: 2 });
    // this.inner_left = this.inner_left + this.score * 57.4
    this.inner_left = 55.7 * this.score + 3
    this.text_left = this.inner_left - traits.length * 18 / 2 + 8
    if (this.text_left > 536 && traits.length === 4) {
      this.text_left = this.inner_left - traits.length * 18 / 2 - 14
    }
    // console.log(traits.length)

    if (this.score <= 2.5) {
      this.inner_color = '#16C063'
      this.out_color = '#16C063'
      this.text_color = '#16C063'
    } else if(this.score <= 5){
      this.inner_color = '#4BCEF0'
      this.out_color = '#4BCEF0'
      this.text_color = '#4BCEF0'
    }else if (this.score <= 7.5) {
      this.inner_color = '#FBB400'
      this.out_color = '#FBB400'
      this.text_color = '#FBB400'
    } else {
      this.inner_color = '#EF3E3E'
      this.out_color = '#EF3E3E'
      this.text_color = '#EF3E3E'
    }

    // if (this.inner_left <= 57.4 * 2) {
    //   this.inner_color = '#c6e160'
    //   this.out_color = '#c6e160'
    //   this.text_color = '#c6e160'
    // } else if (this.inner_left <= 57.4 * 4) {
    //   this.inner_color = '#54b9ab'
    //   this.out_color = '#54b9ab'
    //   this.text_color = '#54b9ab'
    // } else if (this.inner_left <= 57.4 * 6) {
    //   this.inner_color = '#ffd205'
    //   this.out_color = '#ffd205'
    //   this.text_color = '#ffd205'
    // } else if (this.inner_left <= 57.4 * 8) {
    //   this.inner_color = '#f48004'
    //   this.out_color = '#f48004'
    //   this.text_color = '#f48004'
    // } else {
    //   this.inner_color = '#c23809'
    //   this.out_color = '#c23809'
    //   this.text_color = '#c23809'
    // }

    // 指定图表的配置项和数据
    let option = {
      graphic: {
        elements: [
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
            left: this.text_left,
            top: this.inner_top - 25,
            style: {
              text: traits,
              fill: this.text_color,
              font: '18px Microsoft YaHei'
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
