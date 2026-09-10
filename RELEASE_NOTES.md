# Astral Rhythm v1.0.0 · 七曜工作法

**Work in rhythm. · 一周有章法，一日有节律。**

首个公开版本。用七曜给一周一个方向，用十二时辰给当下一个提示。原生 Windows、离线运行、解压即用。

## 下载 / Download

**普通用户请选择 `Astral-Rhythm_v1.0.0_Windows_x64.zip`**，不是 Source 包。先退出旧程序，完整解压，再运行 **`AstralRhythm.exe`**。无需安装 Go、Python、.NET 或 WebView。

For the ready-to-run application, download **`Astral-Rhythm_v1.0.0_Windows_x64.zip`**. Close an older instance, extract the entire folder and run **`AstralRhythm.exe`**.

## 本版 / Highlights

- 透明任务栏双行：曜星在上、时辰在下，文字分别按五行着色。
- 独立山岚双卡、七曜与十二时辰图标、深浅主题。
- 中文 / English / 跟随系统；右键即时切换。
- 两份 Excel 自定义建议，本地译文精确匹配；全程离线。
- 新的图文首页、原生界面图及可核对的传统典籍出处。

Transparent two-line taskbar companion, independent day-star and shichen cards, light/dark themes, bilingual interface, editable local guidance and an illustrated README with sourced cultural context.

## 升级 / Upgrading

公开版本从 **1.0.0** 开始；早期试用编号不作为正式发布序列。设置目录仍为 `%APPDATA%\Astral`。把自行修改过的“表格”文件夹复制到新版即可，不要覆盖自己的内容。

Public versioning begins at **1.0.0**. Existing settings remain compatible. Copy your customized workbooks into the new package.

## 验证与文件 / Verification and files

Windows CI runs tests, builds the x64 executable and exports bilingual native interface renders. Offscreen UI exports are not full interactive taskbar testing. The EXE is not code-signed.

- `..._Windows_x64.zip`: portable app, guidance, artwork, documentation and source.
- `..._Source.zip`: complete source and build tools.
- `Astral-Rhythm_SHA256.txt`: SHA-256 checksums for both ZIP files.
