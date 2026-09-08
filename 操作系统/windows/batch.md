BAT脚本命令 是 Windows 系统中用于自动化执行一系列 DOS/Windows 命令的批处理脚本，文件扩展名为 `.bat` 或 `.cmd`。它基于 `cmd.exe` 解释器运行，适用于系统管理、文件操作、网络配置等自动化任务。

---

常用 BAT 脚本命令分类

### # 基础命令
- `@echo off`：关闭命令回显（通常放在脚本首行，避免显示命令本身）。
- `echo`：显示消息或控制回显状态。
  - `echo Hello`：输出文本。
  - `echo.`：输出空行。
- `pause`：暂停脚本，显示“请按任意键继续...”。
- `cls`：清屏。
- `rem` 或 `::`：添加注释（`::` 更简洁，但 `rem` 兼容性更好）。

### # 变量操作
- `set var=value`：定义变量。
- `%var%`：引用变量值。
- `set /p input=提示信息`：接收用户输入。
- `set /a result=10+20`：进行算术运算（需 `/a` 参数）。

### # 流程控制
- `if condition (command)`：条件判断。
  - 示例：`if exist file.txt (echo 存在) else (echo 不存在)`。
  - 支持比较运算符：`EQU`（等于）、`NEQ`（不等于）、`LSS`（小于）、`LEQ`（≤）、`GTR`（大于）、`GEQ`（≥）。
- `for %%i in (1,2,3) do echo %%i`：循环执行。
  - `/L` 参数用于数字范围：`for /l %%n in (1,1,5) do ...`。
  - `/F` 用于解析文件或命令输出。
- `goto label` + `:label`：跳转到指定标签，实现循环或分支。
- `call another.bat`：调用其他批处理文件，执行后返回当前脚本。

### # 文件与目录操作
- `dir`：列出目录内容。
- `cd /d 目录路径`：切换目录（`/d` 可同时切换盘符）。
- `md` / `mkdir`：创建目录。
- `rd` / `rmdir`：删除空目录；`/s /q` 可递归删除非空目录。
- `copy` / `xcopy` / `robocopy`：复制文件或目录（`robocopy` 功能更强大）。
- `del` / `erase`：删除文件；`/q` 为安静模式（不确认）。
- `ren` / `rename`：重命名文件或目录。

### # 系统与网络命令
- `tasklist`：列出当前进程。
- `taskkill /im 进程名 /f`：强制结束进程。
- `ipconfig`：查看网络配置。
- `ping`：测试网络连通性。
- `netstat`：显示网络连接状态。
- `start`：打开程序或文件，如 `start notepad.exe` 或 `start http://www.baidu.com`。

### # 高级技巧
- `%~dp0`：获取当前脚本所在目录。
- `%ERRORLEVEL%`：检查上一条命令的退出状态（0 表示成功）。
- `timeout /t 5`：等待指定秒数（Windows XP 及以上支持）。
- `color 0A`：设置命令行背景（0=黑）和文字颜色（A=绿）。
- `title 自定义标题`：更改命令窗口标题。

---

示例：简易备份脚本
```bat
@echo off
title 文件备份工具
set source=C:\重要文件
set backup=D:\Backup\%date:~0,4%%date:~5,2%%date:~8,2%
if not exist "%backup%" mkdir "%backup%"
xcopy "%source%" "%backup%" /s /e /y
echo 备份完成！
pause
```

> 提示：编辑 BAT 文件建议使用 记事本 或 Notepad++，并保存为 ANSI 编码 以避免中文乱码。

如需完整命令帮助，可在 CMD 中输入 `命令名 /?`，例如 `for /?` 或 `if /?`。