// 在 store 中创建一个模块（WebSocket 模块）
import {Notification} from 'element-ui';
import {getToken} from "@/utils/auth";

const state = {
  ws: null,
  heartCheckTimer: null,
  msg: null,
};

const mutations = {
  SET_WS(state, ws) {
    state.ws = ws;
    console.log("SET_WS", state.ws);
  },
  CLEAR_WS(state) {
    state.ws = null;
  },
  SET_HEART_CHECK_TIMER(state, timer) {
    state.heartCheckTimer = timer;
  },
  SET_MSG(state, msg) {
    state.msg = msg
  }
};

const actions = {
  startWebSocket({commit, dispatch, state}, payload) {
    // 防御 reconnect 时未传参数（payload 为 undefined）
    const {url, user} = payload || {};
    if (!url) {
      console.warn('[ws] startWebSocket called without url, skip');
      return;
    }
    if (url && (!state.ws || state.ws.readyState !== WebSocket.OPEN)) {
      console.log("SOCKET_PATH:", process.env);
      // const socketUrl = `${process.env.VUE_APP_SOCKET_PATH}/ws?token=${getToken()}`;
      const ws = new WebSocket(url);

      ws.onmessage = function (e) {
        console.log("用户：" + user + ` >>>>>>>> ${new Date().toLocaleString()} >>>>> 收到消息 ${e.data}`, state.ws);
        let data = JSON.parse(JSON.parse(e.data).data)
        let message = data.tester + " " + data.notAnswered + " " + "题未作答！"
        // if (e.data !== "ping") {
        //   Notification({
        //     type: "warning",
        //     title: '异常答题',
        //     message: message,
        //     position: "top-right",
        //     duration: 10000,
        //     showClose: true
        //   });
        // }

        // 自定义全局监听事件
        // window.dispatchEvent(new CustomEvent('onmessageWS', {
        //   detail: {
        //     data: e.data
        //   }
        // }))
      };

      ws.onclose = function () {
        console.log("用户：" + user + ` >>>>>>>> ${new Date().toLocaleString()} >>>>> 连接已关闭`);
        // 尝试重新连接
        dispatch('reconnectWebSocket');
      };

      ws.onopen = function () {
        console.log("用户：" + user + ` >>>>>>>> ${new Date().toLocaleString()} >>>>> 连接成功`, ws);
        // Notification({
        //   type: "success",
        //   title: '成功',
        //   message: '会话连接成功',
        //   position: "top-right",
        //   duration: 3000,
        //   showClose: true
        // });
        // 保存 WebSocket 连接信息
        commit('SET_WS', ws);
        // // 在这里调用 sendWebSocketMessage，确保 state.ws 已经被正确设置
        // 开始心跳检测
        dispatch('startHeartCheck', user);
      };

      ws.onerror = function (e) {
        console.log(`${new Date().toLocaleString()} >>>>> 数据传输发生异常`, e);
        // Notification({
        //   type: "error",
        //   title: '错误',
        //   message: '会话连接异常，服务已断开',
        //   position: "top-right",
        //   duration: 3000,
        //   showClose: true
        // });
      };
    }
  },

  sendWebSocketMessage({state}, {user, msg}) {
    console.log("用户：" + user + ` >>>>> ${new Date().toLocaleString()} >>>>> 发送消息：${msg}`, state.ws);
    state.ws.send(JSON.stringify(msg));
  },

  reconnectWebSocket({dispatch}) {
    dispatch('clearWebSocket');
    // 递归调用，一直尝试重连
    setTimeout(() => {
      dispatch('startWebSocket');
    }, 6000);
  },

  clearWebSocket({commit, state}) {
    if (state.ws) {
      state.ws.close();
      commit('CLEAR_WS');
    }
  },

  startHeartCheck({commit, dispatch, state}, user) {
    console.log("用户：" + user + ` >>>>> ${new Date().toLocaleString()} >>>>> 开始心跳检测`, state.ws);
    // 清除之前的计时器
    dispatch('clearHeartCheckTimer');

    // 创建新的计时器
    let roomId = state.ws.url[state.ws.url.length - 1]
    dispatch('sendWebSocketMessage', {user:user, msg:{type: 0, roomId: roomId, data: {user:user, msg:'ping'}}});
    const timer = setInterval(() => {
      if (!state.ws || state.ws.readyState !== WebSocket.OPEN) {
        console.log("用户：" + user + ` >>>>> ${new Date().toLocaleString()} >>>>> 心跳检测失败,触发重连`, state.ws);
        dispatch('reconnectWebSocket');
      } else {
        console.log("用户：" + user + ` >>>>> ${new Date().toLocaleString()} >>>>> 心跳正常,继续下一次心跳检测`, state.ws);
        dispatch('sendWebSocketMessage', {user:user, msg:{type: 0, roomId: 1, data: {user:user, msg:'ping'}}});
      }
    }, 1000 * 29);
    commit('SET_HEART_CHECK_TIMER', timer);
  },

  clearHeartCheckTimer({commit, state}) {
    const timer = state.heartCheckTimer;
    timer && clearInterval(timer);
    commit('SET_HEART_CHECK_TIMER', null);
  },
};

export default {
  state,
  mutations,
  actions,
};
