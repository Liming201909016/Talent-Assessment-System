<template>
  <div style="text-align:center;">
    <div ref="chart" class="bar-chart-inner pdf-group-item" style="width: 100%; height: 450px;"></div>
  </div>
</template>

<script>

import * as echarts from 'echarts';
import el from "element-ui/src/locale/lang/el";

export default {

  name: 'BarChart',
  props: ['scores'],

  data() {
    return {

      result_data: [
        {value: 3.2, name: '焦虑'},
        {value: 5.6, name: '抑郁'},
        {value: 6.4, name: '心理失衡'},
        {value: 4.1, name: '敌意'},
        {value: 2.1, name: '恐惧'},
        {value: 7.9, name: '身体不适'},
        {value: 8.0, name: '认知衰退'},
        {value: 6.8, name: '情绪化'},
        {value: 7.2, name: '挫折感'},
        {value: 7.6, name: '自我否定'},
        {value: 7.6, name: '怀疑感'},
        {value: 7.6, name: '职业倦怠'}
      ]

    }
  },

  mounted() {
    this.initCharts();
  },
  methods: {
    initCharts() {

      let newScores = []
      let keys = Object.keys(this.scores)
      keys.forEach((key) => {
        newScores.push({value: this.scores[key], name: key})
      })

      console.log(newScores)

      let dataSort = newScores.sort((a, b) => a.value - b.value)
      let valueList = [];
      let nameList = [];
      let lowScore = [];
      let midScore = [];
      let heightScore = [];

      dataSort.forEach((data) => {
        valueList.push(data.value);
        if (data.value < 4.5) {
          lowScore.push(data.value);
        } else if (data.value > 7) {
          heightScore.push(data.value);
        } else {
          midScore.push(data.value)
        }
        nameList.push(data.name)
      })
      // 新增：计算Y轴最大值（用于确定背景顶部位置）
      const yAxisMax = 10 - 0.2;


      let charts = echarts.init(this.$refs.chart, null, { devicePixelRatio: 2 });


      // 绘制左侧面
      const CubeLeft = echarts.graphic.extendShape({
        shape: {
          x: 0,
          y: 0,
        },
        buildPath: function (ctx, shape) {
          const xAxisPoint = shape.xAxisPoint;
          const c0 = [shape.x + 2, shape.y];
          const c1 = [shape.x - 7, shape.y - 1];
          const c2 = [xAxisPoint[0] - 7, xAxisPoint[1] - 3];
          const c3 = [xAxisPoint[0] + 2, xAxisPoint[1]];
          ctx
            .moveTo(c0[0], c0[1])
            .lineTo(c1[0], c1[1])
            .lineTo(c2[0], c2[1])
            .lineTo(c3[0], c3[1])
            .closePath();
        },
      });
      // 绘制右侧面
      const CubeRight = echarts.graphic.extendShape({
        shape: {
          x: 0,
          y: 0,
        },
        buildPath: function (ctx, shape) {
          const xAxisPoint = shape.xAxisPoint;
          const c1 = [shape.x + 1.5, shape.y];
          const c2 = [xAxisPoint[0] + 1.5, xAxisPoint[1]];
          const c3 = [xAxisPoint[0] + 10, xAxisPoint[1] - 5];
          const c4 = [shape.x + 10, shape.y - 5];
          ctx
            .moveTo(c1[0], c1[1])
            .lineTo(c2[0], c2[1])
            .lineTo(c3[0], c3[1])
            .lineTo(c4[0], c4[1]).closePath();
        },
      });
      // 绘制顶面
      const CubeTop = echarts.graphic.extendShape({
        shape: {
          x: 0,
          y: 0,
        },
        buildPath: function (ctx, shape) {
          const c1 = [shape.x + 2, shape.y + 2];
          const c2 = [shape.x + 10, shape.y - 5]; //右点
          const c3 = [shape.x - 0, shape.y - 7];
          const c4 = [shape.x - 7, shape.y - 1];
          ctx
            .moveTo(c1[0], c1[1])
            .lineTo(c2[0], c2[1])
            .lineTo(c3[0], c3[1])
            .lineTo(c4[0], c4[1])
            .closePath();
        },
      });
      // 注册三个面图形
      echarts.graphic.registerShape("CubeLeft", CubeLeft);
      echarts.graphic.registerShape("CubeRight", CubeRight);
      echarts.graphic.registerShape("CubeTop", CubeTop);

      let option = {
        title: {
          // text: "综合评价报告", // 标题
        },
        xAxis: {
          type: "category",
          data: nameList,
          axisLabel: {
            interval: 0,//代表显示所有x轴标签显示
            rotate: 45, //代表逆时针旋转45度
          },
          axisTick: {
            show: false // 不显示坐标轴刻度线
          },
          axisLine: {
            show: true,
            lineStyle: {
              color: "#0C315D",
            },
          },
        },
        yAxis: {
          type: "value",
          axisLine: {
            show: false // 不显示坐标轴线条
          },
          axisTick: {
            show: false // 显示刻度线
          },
          axisLabel: {
            show: true // 显示刻度标签（数值文本）
          }
        },
        series: [
          {
            type: "custom",
            renderItem: (params, api) => {
              const value = api.value(1); // api.value(1) 获取当前数据值
              let cubeLeftStyle = "";
              let cubeRightStyle = "";
              let cubeTopStyle = "";
              if (value < 4.5) {
                // 绿色系（<4.5）
                cubeLeftStyle = '#02B38F';       // 左侧面
                // cubeRightStyle = '#02A07A';      // 右侧面（稍暗）
                cubeRightStyle = '#02B38F';       // 右侧面
                cubeTopStyle = '#42C6AD';        // 顶面（稍亮）
              } else if (value <= 7) {
                // 橙色系（4.5-7）
                cubeLeftStyle = '#FFA300';       // 左侧面
                // cubeRightStyle = '#E08E00';      // 右侧面（稍暗）
                cubeRightStyle = '#FFA300';      // 右侧面
                cubeTopStyle = '#FFD791';        // 顶面（稍亮）
              } else {
                // 红色系（>7）
                cubeLeftStyle = '#E80000';       // 左侧面
                // cubeRightStyle = '#D00000';      // 右侧面（稍暗）
                cubeRightStyle = '#E80000';       // 右侧面
                cubeTopStyle = '#FF8C8C';        // 顶面（稍亮）
              }
              //颜色end
              const location = api.coord([api.value(0), api.value(1)]);
              const xAxisPoint = api.coord([api.value(0), 0]); // 柱子底部坐标（X轴位置）

              // 新增：柱子背景（撑满Y轴高度）
              const backgroundStyle = {
                type: 'rect',
                shape: {
                  x: location[0] - 7, // 左右位置保持不变
                  // 背景顶部
                  y: api.coord([api.value(0), yAxisMax])[1],
                  width: 17, // 宽度保持不变
                  // 背景高度：从Y轴顶端延伸至X轴（底部）
                  height: xAxisPoint[1] - api.coord([api.value(0), yAxisMax])[1]
                },
                style: {
                  fill: '#EFEFEF', // 背景色保持不变
                  opacity: 0.8     // 透明度保持不变
                }
              };
              // 新增：顶部菱形
              const diamondSize = 8; // 菱形大小
              const diamondStyle = {
                type: 'polygon',
                shape: {
                  // 菱形中心定位在背景顶部中心
                  points: [
                    [location[0] + 1.5, api.coord([api.value(0), yAxisMax])[1] - diamondSize + 2], // 上顶点
                    [location[0] + 1.5 + diamondSize, api.coord([api.value(0), yAxisMax])[1]], // 右顶点
                    [location[0] + 1.5, api.coord([api.value(0), yAxisMax])[1] + diamondSize - 2], // 下顶点
                    [location[0] + 1.5 - diamondSize, api.coord([api.value(0), yAxisMax])[1]]  // 左顶点
                  ]
                },
                style: {
                  fill: '#FBFBFB',
                  // fill: '#ea1515',
                }
              };
              // 新增：柱子顶部数值标签
              const valueLabelStyle = {
                type: 'text', // 绘制文本（数值）
                position: [location[0], location[1] - 15], // 设置文本的位置，这里稍微向上偏移一些以避免遮挡柱子顶部。具体偏移量根据实际效果调整。
                style: { // 设置文本样式，包括字体大小、颜色等。
                  text: value, // 设置显示的文本内容为数据值。
                  textAlign: 'center',  // 水平居中对齐
                  textVerticalAlign: 'middle',  // 垂直居中对齐
                  fill: '#111111', // 设置文本颜色为白色，以便于阅读。根据背景色调整颜色。
                },
                // state: 'emphasis' // 可以设置文本的高亮状态，以便在鼠标悬停时更加明显。
              };
              return {
                type: "group",
                children: [
                  backgroundStyle, // 背景层（放在最前面，显示在立方体下方）
                  diamondStyle,    // 新增：菱形放在背景上方
                  {
                    type: "CubeLeft",
                    shape: {
                      api,
                      xValue: api.value(0),
                      yValue: api.value(1),
                      x: location[0],
                      y: location[1],
                      xAxisPoint: api.coord([api.value(0), 0]),
                    },
                    style: {
                      fill: cubeLeftStyle,
                    },
                  },
                  {
                    type: "CubeRight",
                    shape: {
                      api,
                      xValue: api.value(0),
                      yValue: api.value(1),
                      x: location[0],
                      y: location[1],
                      xAxisPoint: api.coord([api.value(0), 0]),
                    },
                    style: {
                      fill: cubeRightStyle,
                    },
                  },
                  {
                    type: "CubeTop",
                    shape: {
                      api,
                      xValue: api.value(0),
                      yValue: api.value(1),
                      x: location[0],
                      y: location[1],
                      xAxisPoint: api.coord([api.value(0), 0]),
                    },
                    style: {
                      fill: cubeTopStyle,
                    },
                  },
                  valueLabelStyle  // 新增：数值标签（放在最上层）
                ],
              };
            },
            data: valueList,
            tooltip: {
              show: false // 禁用悬停提示框
            },
            silent: true //禁止系列响应鼠标事件（彻底避免交互触发）
          },

        ],
        tooltip: {
          // 提示框组件
        },
      };
      charts.setOption(option);
    },
  },
};
</script>

<style lang="scss" scoped></style>

