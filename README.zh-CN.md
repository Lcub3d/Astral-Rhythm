# 七曜工作法 · Astral Rhythm

**Work in rhythm. · 循着节律工作。**

轻量、离线运行的 Windows 任务栏工具：用七曜安排一周方向，用十二时辰提示当下事项。

**简体中文 · [English](README.md)**

## 下载和启动

打开 [Windows 自动构建](https://github.com/Lcub3d/Astral-Rhythm/actions/workflows/windows.yml)，选择一次成功运行，在 **Artifacts** 中下载 **Astral-Rhythm-Windows-x64-and-source**，然后解开其中的 **Windows x64 ZIP**。GitHub 下载 Actions 产物需要登录。以后有正式发布时，也可从 [Releases](https://github.com/Lcub3d/Astral-Rhythm/releases) 下载。

先退出旧版，完整解压，再双击 **`AstralRhythm.exe`**。成品不需要安装 Go、Python、.NET 或 WebView。**Code → Download ZIP** 下载的是源码，不是可直接运行的软件。

任务栏第一行是**曜星**，第二行是**时辰**。鼠标悬停显示卡片；单击固定或收起；按住左键拖动；右键打开设置。

## 界面

保留 V3 的山岚双卡、紧凑文字和间距：曜星在上、时辰在下，解释分区独立。上卡底纹随曜日变化，下卡使用自己的时辰底纹，提供深浅两套主题。

任务栏按钮**没有可见底板和边框**。两行文字按各自五行着色，不把曜日和时辰的颜色混在一起。长内容自动换行并可滚动，没有常驻底部说明或工具栏。

## 中英文切换

右键 → **语言 / Language** → **跟随系统 / 简体中文 / English**，即时切换并记住选择。中文 Windows 界面默认使用简体中文，其他系统界面语言默认使用英文。

也可使用 `Start-English.cmd`、`Start-Chinese.cmd`、`Start-System.cmd`。菜单、悬浮卡、内置建议、传统解释和复制文本均支持中英文。上方标题为 **曜星 / Day Star**，不再显示“曜星／五行”。

## 修改工作建议

表格位于 **`表格`** 文件夹，保存后约 5 秒重新读取。请保留文件名、表头和星期／时辰标识；无效修改不会替换上一次有效数据，原表没有的工作等级不会补填。公开副本仅移除本机路径、作者等文件元数据，单元格和格式保持不变。

英文译文位于 **`locales/en.json`**，以中文原文精确匹配，不联网、不自动调用翻译服务。新改内容没有对应译文时保留原文，不会错误套用旧建议。补充本地译文即可；切换语言不会修改 Excel。

读取顺序：原始表格 → 两表均不存在时读取 `data.json` → 内置数据。

## 升级与设置

英文名称正式确定为 **Astral Rhythm**，中文仍为 **七曜工作法**，程序文件为 `AstralRhythm.exe`。

设置目录继续沿用 **`%APPDATA%\Astral`**，兼容 V3.1.0；尚无 Astral 设置时可迁移旧版七曜时辰的位置、字号、主题和分区展开选择，不删除旧设置。升级时，把自己改过的旧版“表格”文件夹复制到新版同名目录。

`Preview.cmd` / `打开悬浮卡.cmd` 打开详情；`Reset-Position.cmd` / `恢复默认位置.cmd` 恢复位置；`Self-Test.cmd` / `自检.cmd` 生成 `self-test.txt`。

## 构建与开发

原生 Go + Win32，无第三方 Go 依赖，不使用 CGO。源码位于 `src/`，需要 Go 1.23 或更新版本。Windows 可运行 `src/build.cmd`；普通使用者只需运行成品 EXE。

重建版本资源：`python tools/build_resources.py`，仅需 Python 标准库。美术素材已随源码提交，重建素材另需 Pillow、CairoSVG：`python tools/build_assets.py`。程序使用 Windows 系统字体，不分发字体文件。

`python tools/package.py` 在 `dist/` 生成便携包、源码包和 SHA-256 校验文件。[GitHub Actions](https://github.com/Lcub3d/Astral-Rhythm/actions) 包含 Linux 单元测试与 vet、Windows 原生单元测试、x64 编译和打包。成功运行后可下载构建产物；推送与源码版本一致的版本标签后，发布流程会在测试通过后尝试创建 Release，需要仓库允许相应发布权限。PR 不获得发布写权限。

## 验证范围

程序是贴附**主任务栏**的独立原生小窗，不会挤开现有图标，也不预留托盘空间。使用 Windows 本地时间，曜星在 00:00 换日，子时为 23:00—01:00。

自动排版检查使用估算字宽，测试和编译通过不等于 Windows 桌面视觉实测；原生字体、贴附和悬停效果仍需实机检查。EXE 尚未数字签名。仓库暂未选择开源许可证。
