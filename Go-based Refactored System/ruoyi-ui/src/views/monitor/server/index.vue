<template>
  <div class="app-container">
    <el-row>
      <el-col :span="12" class="card-box">
        <el-card>
          <div slot="header"><span>CPU</span></div>
          <div class="el-table el-table--enable-row-hover el-table--medium">
            <table cellspacing="0" style="width: 100%;">
              <thead>
                <tr>
                  <th class="el-table__cell is-leaf"><div class="cell">属性</div></th>
                  <th class="el-table__cell is-leaf"><div class="cell">值</div></th>
                </tr>
              </thead>
              <tbody>
                <tr>
                  <td class="el-table__cell is-leaf"><div class="cell">核心数</div></td>
                  <td class="el-table__cell is-leaf"><div class="cell" v-if="server.cpu">{{ server.cpu.cpuNum }}</div></td>
                </tr>
                <tr>
                  <td class="el-table__cell is-leaf"><div class="cell">用户使用率</div></td>
                  <td class="el-table__cell is-leaf"><div class="cell" v-if="server.cpu">{{ server.cpu.used }}%</div></td>
                </tr>
                <tr>
                  <td class="el-table__cell is-leaf"><div class="cell">系统使用率</div></td>
                  <td class="el-table__cell is-leaf"><div class="cell" v-if="server.cpu">{{ server.cpu.sys }}%</div></td>
                </tr>
                <tr>
                  <td class="el-table__cell is-leaf"><div class="cell">当前空闲率</div></td>
                  <td class="el-table__cell is-leaf"><div class="cell" v-if="server.cpu">{{ server.cpu.free }}%</div></td>
                </tr>
              </tbody>
            </table>
          </div>
        </el-card>
      </el-col>

      <el-col :span="12" class="card-box">
        <el-card>
          <div slot="header"><span>内存</span></div>
          <div class="el-table el-table--enable-row-hover el-table--medium">
            <table cellspacing="0" style="width: 100%;">
              <thead>
                <tr>
                  <th class="el-table__cell is-leaf"><div class="cell">属性</div></th>
                  <th class="el-table__cell is-leaf"><div class="cell">内存</div></th>
                  <th class="el-table__cell is-leaf"><div class="cell">Go Runtime</div></th>
                </tr>
              </thead>
              <tbody>
                <tr>
                  <td class="el-table__cell is-leaf"><div class="cell">总内存</div></td>
                  <td class="el-table__cell is-leaf"><div class="cell" v-if="server.mem">{{ server.mem.total }}G</div></td>
                  <td class="el-table__cell is-leaf"><div class="cell" v-if="server.jvm">{{ server.jvm.total }}M</div></td>
                </tr>
                <tr>
                  <td class="el-table__cell is-leaf"><div class="cell">已用内存</div></td>
                  <td class="el-table__cell is-leaf"><div class="cell" v-if="server.mem">{{ server.mem.used}}G</div></td>
                  <td class="el-table__cell is-leaf"><div class="cell" v-if="server.jvm">{{ server.jvm.used}}M</div></td>
                </tr>
                <tr>
                  <td class="el-table__cell is-leaf"><div class="cell">剩余内存</div></td>
                  <td class="el-table__cell is-leaf"><div class="cell" v-if="server.mem">{{ server.mem.free }}G</div></td>
                  <td class="el-table__cell is-leaf"><div class="cell" v-if="server.jvm">{{ server.jvm.free }}M</div></td>
                </tr>
                <tr>
                  <td class="el-table__cell is-leaf"><div class="cell">使用率</div></td>
                  <td class="el-table__cell is-leaf"><div class="cell" v-if="server.mem" :class="{'text-danger': server.mem.usage > 80}">{{ server.mem.usage }}%</div></td>
                  <td class="el-table__cell is-leaf"><div class="cell" v-if="server.jvm" :class="{'text-danger': server.jvm.usage > 80}">{{ server.jvm.usage }}%</div></td>
                </tr>
              </tbody>
            </table>
          </div>
        </el-card>
      </el-col>

      <el-col :span="24" class="card-box">
        <el-card>
          <div slot="header">
            <span>服务器信息</span>
          </div>
          <div class="el-table el-table--enable-row-hover el-table--medium">
            <table cellspacing="0" style="width: 100%;">
              <tbody>
                <tr>
                  <td class="el-table__cell is-leaf"><div class="cell">服务器名称</div></td>
                  <td class="el-table__cell is-leaf"><div class="cell" v-if="server.sys">{{ server.sys.computerName }}</div></td>
                  <td class="el-table__cell is-leaf"><div class="cell">操作系统</div></td>
                  <td class="el-table__cell is-leaf"><div class="cell" v-if="server.sys">{{ server.sys.osName }}</div></td>
                </tr>
                <tr>
                  <td class="el-table__cell is-leaf"><div class="cell">服务器IP</div></td>
                  <td class="el-table__cell is-leaf"><div class="cell" v-if="server.sys">{{ server.sys.computerIp }}</div></td>
                  <td class="el-table__cell is-leaf"><div class="cell">系统架构</div></td>
                  <td class="el-table__cell is-leaf"><div class="cell" v-if="server.sys">{{ server.sys.osArch }}</div></td>
                </tr>
              </tbody>
            </table>
          </div>
        </el-card>
      </el-col>

      <el-col :span="24" class="card-box">
        <el-card>
          <div slot="header">
            <span>Go 运行时信息</span>
          </div>
          <div class="el-table el-table--enable-row-hover el-table--medium">
            <table cellspacing="0" style="width: 100%;table-layout:fixed;">
              <tbody>
                <tr>
                  <td class="el-table__cell is-leaf"><div class="cell">运行环境</div></td>
                  <td class="el-table__cell is-leaf"><div class="cell" v-if="server.jvm">{{ server.jvm.name }}</div></td>
                  <td class="el-table__cell is-leaf"><div class="cell">Go 版本</div></td>
                  <td class="el-table__cell is-leaf"><div class="cell" v-if="server.jvm">{{ server.jvm.version }}</div></td>
                </tr>
                <tr>
                  <td class="el-table__cell is-leaf"><div class="cell">启动时间</div></td>
                  <td class="el-table__cell is-leaf"><div class="cell" v-if="server.jvm">{{ server.jvm.startTime }}</div></td>
                  <td class="el-table__cell is-leaf"><div class="cell">运行时长</div></td>
                  <td class="el-table__cell is-leaf"><div class="cell" v-if="server.jvm">{{ server.jvm.runTime }}</div></td>
                </tr>
                <tr>
                  <td colspan="1" class="el-table__cell is-leaf"><div class="cell">可执行文件</div></td>
                  <td colspan="3" class="el-table__cell is-leaf"><div class="cell" v-if="server.jvm">{{ server.jvm.home }}</div></td>
                </tr>
                <tr>
                  <td colspan="1" class="el-table__cell is-leaf"><div class="cell">项目路径</div></td>
                  <td colspan="3" class="el-table__cell is-leaf"><div class="cell" v-if="server.sys">{{ server.sys.userDir }}</div></td>
                </tr>
                <tr>
                  <td colspan="1" class="el-table__cell is-leaf"><div class="cell">运行信息</div></td>
                  <td colspan="3" class="el-table__cell is-leaf"><div class="cell" v-if="server.jvm">{{ server.jvm.inputArgs }}</div></td>
                </tr>
              </tbody>
            </table>
          </div>
        </el-card>
      </el-col>

      <el-col :span="24" class="card-box">
        <el-card>
          <div slot="header">
            <span>磁盘状态</span>
          </div>
          <div class="el-table el-table--enable-row-hover el-table--medium">
            <table cellspacing="0" style="width: 100%;">
              <thead>
                <tr>
                  <th class="el-table__cell el-table__cell is-leaf"><div class="cell">盘符路径</div></th>
                  <th class="el-table__cell is-leaf"><div class="cell">文件系统</div></th>
                  <th class="el-table__cell is-leaf"><div class="cell">盘符类型</div></th>
                  <th class="el-table__cell is-leaf"><div class="cell">总大小</div></th>
                  <th class="el-table__cell is-leaf"><div class="cell">可用大小</div></th>
                  <th class="el-table__cell is-leaf"><div class="cell">已用大小</div></th>
                  <th class="el-table__cell is-leaf"><div class="cell">已用百分比</div></th>
                </tr>
              </thead>
              <tbody v-if="server.sysFiles">
                <tr v-for="(sysFile, index) in server.sysFiles" :key="index">
                  <td class="el-table__cell is-leaf"><div class="cell">{{ sysFile.dirName }}</div></td>
                  <td class="el-table__cell is-leaf"><div class="cell">{{ sysFile.sysTypeName }}</div></td>
                  <td class="el-table__cell is-leaf"><div class="cell">{{ sysFile.typeName }}</div></td>
                  <td class="el-table__cell is-leaf"><div class="cell">{{ sysFile.total }}</div></td>
                  <td class="el-table__cell is-leaf"><div class="cell">{{ sysFile.free }}</div></td>
                  <td class="el-table__cell is-leaf"><div class="cell">{{ sysFile.used }}</div></td>
                  <td class="el-table__cell is-leaf"><div class="cell" :class="{'text-danger': sysFile.usage > 80}">{{ sysFile.usage }}%</div></td>
                </tr>
              </tbody>
            </table>
          </div>
        </el-card>
      </el-col>

      <el-col :span="24" class="card-box" v-if="server.mysql">
        <el-card>
          <div slot="header"><span>MySQL 数据库</span></div>
          <el-row :gutter="16">
            <el-col :span="12">
              <div class="el-table el-table--enable-row-hover el-table--medium">
                <table cellspacing="0" style="width: 100%;">
                  <thead><tr>
                    <th class="el-table__cell is-leaf"><div class="cell">属性</div></th>
                    <th class="el-table__cell is-leaf"><div class="cell">值</div></th>
                  </tr></thead>
                  <tbody>
                    <tr>
                      <td class="el-table__cell is-leaf"><div class="cell">版本</div></td>
                      <td class="el-table__cell is-leaf"><div class="cell">{{ server.mysql.version }} {{ server.mysql.comment }}</div></td>
                    </tr>
                    <tr>
                      <td class="el-table__cell is-leaf"><div class="cell">字符集</div></td>
                      <td class="el-table__cell is-leaf"><div class="cell">{{ server.mysql.charset }} / {{ server.mysql.collation }}</div></td>
                    </tr>
                    <tr>
                      <td class="el-table__cell is-leaf"><div class="cell">运行时长</div></td>
                      <td class="el-table__cell is-leaf"><div class="cell">{{ server.mysql.uptime }}</div></td>
                    </tr>
                    <tr>
                      <td class="el-table__cell is-leaf"><div class="cell">数据目录</div></td>
                      <td class="el-table__cell is-leaf"><div class="cell">{{ server.mysql.datadir }}</div></td>
                    </tr>
                    <tr>
                      <td class="el-table__cell is-leaf"><div class="cell">连接数</div></td>
                      <td class="el-table__cell is-leaf">
                        <div class="cell" :class="{'text-danger': server.mysql.connUsage > 80}">
                          {{ server.mysql.curConn }} / {{ server.mysql.maxConn }}（{{ server.mysql.connUsage }}%）
                        </div>
                      </td>
                    </tr>
                    <tr>
                      <td class="el-table__cell is-leaf"><div class="cell">活跃线程</div></td>
                      <td class="el-table__cell is-leaf"><div class="cell">{{ server.mysql.threads }}</div></td>
                    </tr>
                    <tr>
                      <td class="el-table__cell is-leaf"><div class="cell">慢查询数</div></td>
                      <td class="el-table__cell is-leaf">
                        <div class="cell" :class="{'text-danger': parseInt(server.mysql.slowQueries) > 100}">{{ server.mysql.slowQueries }}</div>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </el-col>
            <el-col :span="12">
              <div class="el-table el-table--enable-row-hover el-table--medium">
                <table cellspacing="0" style="width: 100%;">
                  <thead><tr>
                    <th class="el-table__cell is-leaf"><div class="cell">性能指标</div></th>
                    <th class="el-table__cell is-leaf"><div class="cell">值</div></th>
                  </tr></thead>
                  <tbody>
                    <tr>
                      <td class="el-table__cell is-leaf"><div class="cell">QPS（平均）</div></td>
                      <td class="el-table__cell is-leaf"><div class="cell">{{ server.mysql.qps }}</div></td>
                    </tr>
                    <tr>
                      <td class="el-table__cell is-leaf"><div class="cell">InnoDB缓冲池</div></td>
                      <td class="el-table__cell is-leaf"><div class="cell">{{ server.mysql.bufferPool }} MB</div></td>
                    </tr>
                    <tr>
                      <td class="el-table__cell is-leaf"><div class="cell">缓冲池命中率</div></td>
                      <td class="el-table__cell is-leaf">
                        <div class="cell" :class="{'text-danger': server.mysql.hitRate < 95}">{{ server.mysql.hitRate }}%</div>
                      </td>
                    </tr>
                    <tr>
                      <td class="el-table__cell is-leaf"><div class="cell">总查询数</div></td>
                      <td class="el-table__cell is-leaf"><div class="cell">{{ formatNum(server.mysql.questions) }}</div></td>
                    </tr>
                    <tr>
                      <td class="el-table__cell is-leaf"><div class="cell">流量 (收/发)</div></td>
                      <td class="el-table__cell is-leaf"><div class="cell">{{ server.mysql.bytesRecv }} MB / {{ server.mysql.bytesSent }} MB</div></td>
                    </tr>
                    <tr>
                      <td class="el-table__cell is-leaf"><div class="cell">操作统计</div></td>
                      <td class="el-table__cell is-leaf">
                        <div class="cell">S:{{ server.mysql.comSelect }} I:{{ server.mysql.comInsert }} U:{{ server.mysql.comUpdate }} D:{{ server.mysql.comDelete }}</div>
                      </td>
                    </tr>
                    <tr>
                      <td class="el-table__cell is-leaf"><div class="cell">打开表 / 临时表</div></td>
                      <td class="el-table__cell is-leaf"><div class="cell">{{ server.mysql.tableOpen }} / {{ server.mysql.tmpTables }}</div></td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </el-col>
          </el-row>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script>
import { getServer } from "@/api/monitor/server";

export default {
  name: "Server",
  data() {
    return {
      // 服务器信息
      server: []
    };
  },
  created() {
    this.getList();
    this.openLoading();
  },
  methods: {
    /** 查询服务器信息 */
    getList() {
      getServer().then(response => {
        this.server = response.data;
        this.$modal.closeLoading();
      });
    },
    // 打开加载层
    openLoading() {
      this.$modal.loading("正在加载服务监控数据，请稍候！");
    },
    formatNum(n) {
      if (!n) return '0';
      return n.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',');
    }
  }
};
</script>