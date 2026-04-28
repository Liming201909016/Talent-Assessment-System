#!/bin/bash
# 部署脚本 - 在生产服务器执行
set -e

echo "==> 1. 备份当前后端二进制"
sudo cp -f /opt/talent-assessment/server /opt/talent-assessment/server.bak.$(date +%Y%m%d_%H%M%S) 2>/dev/null || echo "  (无旧文件，跳过)"

echo "==> 2. 停止服务"
sudo systemctl stop talent-assessment

echo "==> 3. 部署新后端"
sudo cp /tmp/server-linux /opt/talent-assessment/server
sudo chmod +x /opt/talent-assessment/server
sudo chown root:root /opt/talent-assessment/server

echo "==> 4. 部署前端 dist"
sudo rm -rf /opt/talent-assessment/dist.bak
[ -d /opt/talent-assessment/dist ] && sudo mv /opt/talent-assessment/dist /opt/talent-assessment/dist.bak
sudo mv /tmp/dist-new /opt/talent-assessment/dist
sudo chmod -R 755 /opt/talent-assessment/dist
sudo chown -R root:root /opt/talent-assessment/dist

echo "==> 5. 启动服务"
sudo systemctl start talent-assessment
sleep 3

echo "==> 6. 服务状态"
sudo systemctl status talent-assessment --no-pager | head -10

echo "==> 7. nginx reload"
sudo nginx -t && sudo nginx -s reload

echo "==> 8. 健康检查"
echo "  HTTP /captchaImage: $(curl -s -o /dev/null -w '%{http_code}' http://localhost/prod-api/captchaImage)"
echo "  HTTP /: $(curl -s -o /dev/null -w '%{http_code}' http://localhost/)"
echo "  Backend port 8092: $(ss -tln | grep -c ':8092' || echo 0) listening"

echo ""
echo "✅ 部署完成"
