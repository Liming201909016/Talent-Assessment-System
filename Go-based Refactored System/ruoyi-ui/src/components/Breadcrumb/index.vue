<template>
  <el-breadcrumb class="app-breadcrumb" separator="/">
    <transition-group name="breadcrumb">
      <el-breadcrumb-item v-for="(item,index) in levelList" :key="item.path">
        <span v-if="item.redirect === 'noRedirect' || index == levelList.length - 1" class="no-redirect">{{ item.meta.title }}</span>
        <a v-else @click.prevent="handleLink(item)">{{ item.meta.title }}</a>
      </el-breadcrumb-item>
    </transition-group>
  </el-breadcrumb>
</template>

<script>
export default {
  data() {
    return {
      levelList: null
    }
  },
  watch: {
    $route(route) {
      // if you go to the redirect page, do not update the breadcrumbs
      if (route.path.startsWith('/redirect/')) {
        return
      }
      this.getBreadcrumb()
    }
  },
  created() {
    this.getBreadcrumb()
  },
  methods: {
    getBreadcrumb() {
      // only show routes with meta.title
      let matched = this.$route.matched.filter(item => item.meta && item.meta.title)
      const first = matched[0]

      if (!this.isDashboard(first)) {
        matched = [{ path: '/index', meta: { title: '首页' }}].concat(matched)
      }

      this.levelList = matched.filter(item => item.meta && item.meta.title && item.meta.breadcrumb !== false)

      // FB-056: 若叶子页有 activeMenu，把父级分组的跳转改为 activeMenu，标题改为 activeMenu 对应路由的标题
      // 场景：测评详情(/exam/exam/users/...) activeMenu=/exam/exam → 父级"测评管理"应跳"测试管理"而非默认 redirect 的"题库管理"
      const leaf = this.levelList[this.levelList.length - 1]
      const activeMenu = leaf && leaf.meta && leaf.meta.activeMenu
      if (activeMenu && this.levelList.length >= 2) {
        const parentIdx = this.levelList.length - 2
        const parent = this.levelList[parentIdx]
        if (parent.redirect && parent.redirect !== activeMenu) {
          // 在父级 children 中找 activeMenu 对应的 sibling 路由的 title
          let siblingTitle = null
          if (parent.children && Array.isArray(parent.children)) {
            const target = parent.children.find(c => {
              const full = ('/' + (parent.path + '/' + c.path).replace(/^\/+|\/+$/g, '')).replace(/\/+/g, '/')
              return full === activeMenu
            })
            if (target && target.meta && target.meta.title) siblingTitle = target.meta.title
          }
          // 替换父级项的 path/redirect/title（仅本地副本，不影响 router 配置）
          this.$set(this.levelList, parentIdx, {
            path: activeMenu,
            redirect: activeMenu,
            meta: Object.assign({}, parent.meta, siblingTitle ? { title: siblingTitle } : {})
          })
        }
      }
    },
    isDashboard(route) {
      const name = route && route.name
      if (!name) {
        return false
      }
      return name.trim() === 'Index'
    },
    handleLink(item) {
      const { redirect, path } = item
      if (redirect) {
        this.$router.push(redirect)
        return
      }
      this.$router.push(path)
    }
  }
}
</script>

<style lang="scss" scoped>
.app-breadcrumb.el-breadcrumb {
  display: inline-block;
  font-size: 14px;
  line-height: 50px;
  margin-left: 8px;

  .no-redirect {
    color: #97a8be;
    cursor: text;
  }
}
</style>
