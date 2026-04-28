#!/bin/bash
# Prometheus + Grafana 安装脚本（在生产服务器执行）
set -e

echo "==> 1. 安装 Docker（如未安装）"
if ! command -v docker &> /dev/null; then
  curl -fsSL https://get.docker.com | sudo sh
  sudo usermod -aG docker $USER
  echo "  Docker 已安装，需重新登录或运行 newgrp docker"
fi

echo "==> 2. 创建监控目录"
sudo mkdir -p /opt/monitoring
sudo chown -R $USER:$USER /opt/monitoring

echo "==> 3. 复制配置文件（已通过 scp 上传到 /tmp/monitoring/）"
cp /tmp/monitoring/* /opt/monitoring/

echo "==> 4. 启动监控栈"
cd /opt/monitoring
sudo docker compose up -d

echo "==> 5. 等待启动"
sleep 10

echo "==> 6. 验证"
echo "  Prometheus:        $(curl -s -o /dev/null -w '%{http_code}' http://localhost:9090/-/ready)"
echo "  Node Exporter:     $(curl -s -o /dev/null -w '%{http_code}' http://localhost:9100/metrics)"
echo "  Blackbox:          $(curl -s -o /dev/null -w '%{http_code}' http://localhost:9115/-/healthy)"
echo "  Grafana:           $(curl -s -o /dev/null -w '%{http_code}' http://localhost:3000/api/health)"

echo ""
echo "✅ 安装完成"
echo ""
echo "📊 访问："
echo "  Prometheus: http://20.200.136.133:9090"
echo "  Grafana:    http://20.200.136.133:3000  (admin/admin)"
echo ""
echo "⚠️  请在 Azure 网络安全组放行 9090/3000 端口（仅限你的 IP）"
echo ""
echo "📝 下一步在 Grafana 中导入官方面板："
echo "  Node Exporter Full: ID 1860"
echo "  Blackbox Exporter:  ID 7587"
