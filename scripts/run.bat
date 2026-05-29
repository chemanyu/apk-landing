@echo off
REM 在 Windows VM 上启动 apk-landing 服务。
REM 双击本文件即可运行；服务监听 5001 端口（见 etc\config.yaml）。
cd /d %~dp0

echo ============================================
echo  启动 apk-landing （监听 127.0.0.1:5001）
echo  公网经 nginx /ulink 反代到本服务
echo  关闭本窗口即停止服务
echo ============================================

apk-landing.exe -f etc\config.yaml

REM 服务异常退出时窗口不立即关闭，便于看错误
echo.
echo 服务已退出，按任意键关闭窗口...
pause >nul
