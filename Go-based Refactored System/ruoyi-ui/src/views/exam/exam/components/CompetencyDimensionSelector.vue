<template>
  <div class="competency-selector" v-loading="loading">
    <div class="selector-summary">
      <span>已选择 <strong>{{ selectedCount }}</strong> 个维度</span>
      <span>共 <strong>{{ selectedQuestionCount }}</strong> 道启用题目</span>
      <span class="selector-hint">可跨类别自由组合；发布后不可修改</span>
    </div>

    <el-checkbox-group v-if="dimensions.length" v-model="selectedValues" class="dimension-groups">
      <section v-for="group in groupedDimensions" :key="group.name" class="dimension-group">
        <header class="group-header">
          <span>{{ group.name }}</span>
          <span class="group-count">{{ group.items.length }} 项</span>
        </header>
        <div class="dimension-grid">
          <el-checkbox
            v-for="dimension in group.items"
            :key="dimension.id"
            :label="dimension.id"
            :disabled="isDimensionDisabled(dimension)"
            border
            class="dimension-option"
          >
            <span class="dimension-code">{{ dimension.code }}</span>
            <span>{{ dimension.name }}</span>
            <span class="dimension-category">{{ dimension.applicableCategory }}</span>
            <span class="dimension-count">
              {{ dimension.questionCount || 0 }} 题{{ dimension.questionCount > 0 ? '' : '（不可选）' }}
            </span>
          </el-checkbox>
        </div>
      </section>
    </el-checkbox-group>

    <div v-else-if="!loading" class="empty-state">
      暂无可配置的胜任力维度，请先执行维度数据迁移。
    </div>
  </div>
</template>

<script>
export default {
  name: 'CompetencyDimensionSelector',
  props: {
    value: {
      type: Array,
      default: () => []
    },
    dimensions: {
      type: Array,
      default: () => []
    },
    disabled: {
      type: Boolean,
      default: false
    },
    loading: {
      type: Boolean,
      default: false
    }
  },
  computed: {
    selectedValues: {
      get() {
        return this.value
      },
      set(value) {
        this.$emit('input', value)
      }
    },
    selectedCount() {
      return this.value.length
    },
    selectedQuestionCount() {
      const selected = new Set(this.value)
      return this.dimensions.reduce((total, dimension) => {
        if (!selected.has(dimension.id) || dimension.status !== 0 || dimension.questionCount <= 0) {
          return total
        }
        return total + dimension.questionCount
      }, 0)
    },
    groupedDimensions() {
      const groups = []
      const byName = Object.create(null)
      this.dimensions.forEach(dimension => {
        const name = dimension.virdLevel || '其他'
        if (!byName[name]) {
          byName[name] = { name, items: [] }
          groups.push(byName[name])
        }
        byName[name].items.push(dimension)
      })
      return groups
    }
  },
  methods: {
    isDimensionDisabled(dimension) {
      return this.disabled || dimension.status !== 0 || dimension.questionCount <= 0
    }
  }
}
</script>

<style scoped>
.competency-selector {
  width: 100%;
}
.selector-summary {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  color: #606266;
}
.selector-summary strong {
  color: #409eff;
  font-size: 16px;
}
.selector-hint,
.group-count,
.dimension-category {
  color: #909399;
  font-size: 12px;
}
.dimension-group + .dimension-group {
  margin-top: 16px;
}
.group-header {
  display: flex;
  justify-content: space-between;
  padding-bottom: 8px;
  margin-bottom: 8px;
  border-bottom: 1px solid #ebeef5;
  color: #303133;
  font-weight: 600;
}
.dimension-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 8px;
}
.dimension-option {
  width: 100%;
  min-height: 40px;
  margin: 0;
  display: flex;
  align-items: center;
  box-sizing: border-box;
}
.dimension-code {
  display: inline-block;
  min-width: 32px;
  color: #409eff;
  font-weight: 600;
}
.dimension-category {
  margin-left: 6px;
}
.dimension-count {
  margin-left: auto;
  color: #909399;
  font-size: 12px;
}
.empty-state {
  padding: 24px;
  border: 1px dashed #dcdfe6;
  border-radius: 4px;
  color: #909399;
  text-align: center;
}
@media (max-width: 1200px) {
  .dimension-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}
@media (max-width: 768px) {
  .selector-summary {
    align-items: flex-start;
    flex-direction: column;
    gap: 4px;
  }
  .dimension-grid {
    grid-template-columns: 1fr;
  }
}
</style>
