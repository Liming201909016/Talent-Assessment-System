import Vue from 'vue'
import Router from 'vue-router'

Vue.use(Router)

/* Layout */
import Layout from '@/layout'


/**
 * Note: 路由配置项
 *
 * hidden: true                     // 当设置 true 的时候该路由不会再侧边栏出现 如401，login等页面，或者如一些编辑页面/edit/1
 * alwaysShow: true                 // 当你一个路由下面的 children 声明的路由大于1个时，自动会变成嵌套的模式--如组件页面
 *                                  // 只有一个时，会将那个子路由当做根路由显示在侧边栏--如引导页面
 *                                  // 若你想不管路由下面的 children 声明的个数都显示你的根路由
 *                                  // 你可以设置 alwaysShow: true，这样它就会忽略之前定义的规则，一直显示根路由
 * redirect: noRedirect             // 当设置 noRedirect 的时候该路由在面包屑导航中不可被点击
 * name:'router-name'               // 设定路由的名字，一定要填写不然使用<keep-alive>时会出现各种问题
 * query: '{"id": 1, "name": "ry"}' // 访问路由的默认传递参数
 * roles: ['admin', 'common']       // 访问路由的角色权限
 * permissions: ['a:a:a', 'b:b:b']  // 访问路由的菜单权限
 * meta : {
    noCache: true                   // 如果设置为true，则不会被 <keep-alive> 缓存(默认 false)
    title: 'title'                  // 设置该路由在侧边栏和面包屑中展示的名字
    icon: 'svg-name'                // 设置该路由的图标，对应路径src/assets/icons/svg
    breadcrumb: false               // 如果设置为false，则不会在breadcrumb面包屑中显示
    activeMenu: '/system/user'      // 当路由设置了该属性，则会高亮相对应的侧边栏。
  }
 */

// 公共路由
export const constantRoutes = [
  {
    path: '/redirect',
    component: Layout,
    hidden: true,
    children: [
      {
        path: '/redirect/:path(.*)',
        component: () => import('@/views/redirect')
      }
    ]
  },
  {
    path: '/login',
    component: () => import('@/views/login'),
    hidden: true
  },
  {
    path: '/register',
    component: () => import('@/views/register'),
    hidden: true
  },
  {
    path: '/404',
    component: () => import('@/views/error/404'),
    hidden: true
  },
  {
    path: '/401',
    component: () => import('@/views/error/401'),
    hidden: true
  },
  {
    path: '',
    component: Layout,
    redirect: 'index',
    children: [
      {
        path: 'index',
        component: () => import('@/views/index'),
        name: 'Index',
        meta: { title: '首页', icon: 'dashboard', affix: true }
      }
    ]
  },
  {
    path: '/exam/competency-results',
    component: Layout,
    hidden: true,
    children: [
      {
        path: ':examId',
        component: () => import('@/views/exam/exam/competencyResults'),
        name: 'CompetencyResults',
        meta: { title: '胜任力测评结果', activeMenu: '/exam/exam' }
      }
    ]
  },
  {
    path: '/user',
    component: Layout,
    hidden: true,
    redirect: 'noredirect',
    children: [
      {
        path: 'profile',
        component: () => import('@/views/system/user/profile/index'),
        name: 'Profile',
        meta: { title: '个人中心', icon: 'user' }
      }
    ]
  },

  {
    path: '/my/exam',
    hidden: true,
    component: () => import('@/views/paper/exam/list'),
    name: 'ExamOnline',
    meta: { title: '在线测评', noCache: true, icon: 'guide' }
  },

  {
    path: '/exam/thank-you',
    hidden: true,
    component: () => import('@/views/paper/exam/thankYou'),
    name: 'ExamThankYou',
    meta: { title: '测评完成', noCache: true }
  },

  {
    path: '/my/exam/tester/:examId/:repoCode?',
    component: () => import('@/views/paper/exam/tester'),
    name: 'tester',
    meta: { title: '考生信息', noCache: false, activeMenu: '/my/exam' },
    hidden: true
  },

  {
    path: '/my/exam/candidate/:examId/:stuFlag/:repoCode/:testerId?',
    component: () => import('@/views/paper/exam/candidate'),
    name: 'candidateInfo',
    meta: { title: '考生信息', noCache: false, activeMenu: '/my/exam' },
    hidden: true
  },

  {
    path: '/my/exam/prepare/:examId/:id',
    component: () => import('@/views/paper/exam/preview'),
    name: 'PreExam',
    meta: { title: '准备测评', noCache: true, activeMenu: '/my/exam' },
    hidden: true
  },

  {
    path: '/exam/mbti/start/:id/:testerId',
    component: () => import('@/views/paper/exam/mbtiExam'),
    name: 'MbtiExam',
    meta: { title: 'MBTI测评', noCache: true, activeMenu: '/my/exam' },
    hidden: true
  },

  {
    path: '/exam/competency/start/:paperId',
    component: () => import('@/views/paper/exam/competencyExam'),
    name: 'CompetencyExam',
    meta: { title: '胜任力测评', noCache: true, activeMenu: '/my/exam' },
    hidden: true
  },
  {
    path: '/exam/competency/report/:paperId',
    component: () => import('@/views/paper/exam/competencyReport'),
    name: 'CompetencyReport',
    meta: { title: '胜任力测评报告', noCache: true },
    hidden: true
  },

  {
    path: '/exam/mbti/result/:id/:testerId',
    component: () => import('@/views/paper/exam/mbtiResult'),
    name: 'MbtiResult',
    meta: { title: 'MBTI结果', noCache: true, activeMenu: '/my/exam' },
    hidden: true
  },

  {
    path: '/my/exam/result/:id/:testerId',
    component: () => import('@/views/paper/exam/result.vue'),
    name: 'ShowExam',
    meta: { title: '测评结果', noCache: true, activeMenu: '/online/exam' },
    hidden: true
  },
  {
    path: '/my/exam/result2/:id/:testerId',
    component: () => import('@/views/paper/exam/result2.vue'),
    name: 'ShowMngExam',
    meta: { title: '测评结果', noCache: true, activeMenu: '/online/exam' },
    hidden: true
  },

  {
    path: '/my/exam/finish',
    component: () => import('@/views/paper/exam/finish.vue'),
    name: 'Finish',
    meta: { title: '感谢参与', noCache: true, activeMenu: '/online/exam' },
    hidden: true
  },

  {
    path: '/my/exam/records',
    hidden: true,
    component: () => import('@/views/user/exam/my'),
    name: 'ListMyExam',
    meta: { title: '我的成绩', noCache: true, icon: 'results' }
  },

  {
    path: '/exam/start/:id/:testerId',
    // roles: ['student', 'admin'],
    component: () => import('@/views/paper/exam/exam'),
    name: 'StartExam',
    meta: { title: '开始测评' },
    hidden: true
  },

  {
    path: '/exam/click/start/:id/:testerId',
    // roles: ['student', 'admin'],
    component: () => import('@/views/paper/exam/examClick.vue'),
    name: 'StartExamClick',
    meta: { title: '开始测评' },
    hidden: true
  },

]

//动态路由，基于用户权限动态去加载
export const dynamicRoutes = [
  {
    path: '/system/user-auth',
    component: Layout,
    hidden: true,
    permissions: ['system:user:edit'],
    children: [
      {
        path: 'role/:userId(\\d+)',
        component: () => import('@/views/system/user/authRole'),
        name: 'AuthRole',
        meta: { title: '分配角色', activeMenu: '/system/user' }
      }
    ]
  },
  {
    path: '/system/role-auth',
    component: Layout,
    hidden: true,
    permissions: ['system:role:edit'],
    children: [
      {
        path: 'user/:roleId(\\d+)',
        component: () => import('@/views/system/role/authUser'),
        name: 'AuthUser',
        meta: { title: '分配用户', activeMenu: '/system/role' }
      }
    ]
  },
  {
    path: '/system/dict-data',
    component: Layout,
    hidden: true,
    permissions: ['system:dict:list'],
    children: [
      {
        path: 'index/:dictId(\\d+)',
        component: () => import('@/views/system/dict/data'),
        name: 'Data',
        meta: { title: '字典数据', activeMenu: '/system/dict' }
      }
    ]
  },
  {
    path: '/monitor/job-log',
    component: Layout,
    hidden: true,
    permissions: ['monitor:job:list'],
    children: [
      {
        path: 'index',
        component: () => import('@/views/monitor/job/log'),
        name: 'JobLog',
        meta: { title: '调度日志', activeMenu: '/monitor/job' }
      }
    ]
  },
  {
    path: '/tool/gen-edit',
    component: Layout,
    hidden: true,
    permissions: ['tool:gen:edit'],
    children: [
      {
        path: 'index/:tableId(\\d+)',
        component: () => import('@/views/tool/gen/editTable'),
        name: 'GenEdit',
        meta: { title: '修改生成配置', activeMenu: '/tool/gen' }
      }
    ]
  },

  {
    path: '/',
    component: Layout,
    redirect: '/dashboard',
    children: [
      {
        path: 'qu/view/:id',
        component: () => import('@/views/qu/qu/view'),
        name: 'ViewQu',
        meta: { title: '题目详情', noCache: true, activeMenu: '/manage/qu' },
        hidden: true
      }
    ]
  },

  {
    path: '/exam',
    component: Layout,
    roles: ['admin'],
    redirect: '/exam/repo',
    name: 'Manage',
    meta: {
      title: '测评管理',
      icon: 'example',
    },
    children: [

      {
        path: 'repo',
        component: () => import('@/views/qu/repo'),
        name: 'ListRepo',
        meta: { title: '题库管理', noCache: false, icon: 'repo' }
      },

      {
        path: 'repo/:repoId/questions',
        component: () => import('@/views/qu/qu'),
        name: 'RepoQuList',
        meta: { title: '题目管理', noCache: true, activeMenu: '/qu/repo' },
        hidden: true
      },

      {
        path: 'competency/questions',
        component: () => import('@/views/qu/competency'),
        name: 'CompetencyQuestionList',
        meta: { title: '胜任力测验题库', noCache: true, activeMenu: '/exam/repo' },
        hidden: true
      },

      {
        path: 'competency/dimensions',
        component: () => import('@/views/qu/competency/dimensions'),
        name: 'CompetencyDimensionMaintenance',
        meta: { title: '胜任力维度维护', noCache: true, activeMenu: '/exam/repo' },
        hidden: true
      },

      {
        path: 'repo/add',
        component: () => import('@/views/qu/repo/form'),
        name: 'AddRepo',
        meta: { title: '添加题库', noCache: true, activeMenu: '/exam/repo' },
        hidden: true
      },

      {
        path: 'repo/update/:id',
        component: () => import('@/views/qu/repo/form'),
        name: 'UpdateRepo',
        meta: { title: '题库详情', noCache: true, activeMenu: '/exam/repo' },
        hidden: true
      },

      {
        path: 'qu',
        component: () => import('@/views/qu/qu'),
        name: 'ListQu',
        meta: { title: '题目管理', noCache: false, icon: 'support' }
      },

      {
        path: 'qu/add',
        component: () => import('@/views/qu/qu/form'),
        name: 'AddQu',
        meta: { title: '添加题目', noCache: true, activeMenu: '/exam/qu' },
        hidden: true
      },

      {
        path: 'qu/update/:id',
        component: () => import('@/views/qu/qu/form'),
        name: 'UpdateQu',
        meta: { title: '修改题目', noCache: true, activeMenu: '/exam/qu' },
        hidden: true
      },

      {
        path: 'exam',
        component: () => import('@/views/exam/exam'),
        name: 'ListExam',
        meta: { title: '测试管理', noCache: false, icon: 'log' }
      },

      {
        path: 'exam/statistics/:examId/:title/:state/:isOpen',
        component: () => import('@/views/exam/exam/statistics'),
        name: 'StatisticsExam',
        meta: { title: '测试统计', noCache: true, activeMenu: '/exam/exam' },
        hidden: true
      },

      {
        path: 'exam/add',
        component: () => import('@/views/exam/exam/form'),
        name: 'AddExam',
        meta: { title: '添加测试', noCache: true, activeMenu: '/exam/exam' },
        hidden: true
      },

      {
        path: 'exam/update/:id',
        component: () => import('@/views/exam/exam/form'),
        name: 'UpdateExam',
        meta: { title: '修改测试', noCache: true, activeMenu: '/exam/exam' },
        hidden: true
      },
      {
        path: 'exam/users/:examId?/:isOpen?/:title?',
        component: () => import('@/views/user/exam'),
        name: 'ListExamUser',
        meta: { title: '测评详情', noCache: false, activeMenu: '/exam/exam' },
        hidden: true
      },
      {
        path: 'exam/paper/:examId',
        component: () => import('@/views/paper/paper'),
        name: 'ListPaper',
        meta: { title: '测试记录', noCache: true, activeMenu: '/exam/exam' },
        hidden: true
      }
    ]
  },
]

// 防止连续点击多次路由报错
let routerPush = Router.prototype.push;
Router.prototype.push = function push(location) {
  return routerPush.call(this, location).catch(err => err)
}

export default new Router({
  // mode: 'history', // 去掉url中的#
  mode: 'hash', // 去掉url中的#
  scrollBehavior: () => ({ y: 0 }),
  routes: constantRoutes
})
