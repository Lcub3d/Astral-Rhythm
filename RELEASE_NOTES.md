# Astral Rhythm v1.1.0 · 只在需要时，看见下一步

**Work in rhythm. · 当下专注，未来可见。**

本版保留原有山岚双卡、透明任务栏和中英双语，新增可折叠的未来预览。

## 下载 / Download

下载 **`Astral-Rhythm_v1.1.0_Windows_x64.zip`**，退出旧程序、完整解压，运行 **`AstralRhythm.exe`**。无需安装运行环境。

Download the **Windows_x64.zip**, close the previous instance, extract the entire folder and run **AstralRhythm.exe**.

## 新功能 / What's new

- 曜星卡：**未来 3 日**，名称、日期、五行色与简短工作标签。
- 时辰卡：**后续 4 个时辰**，名称、准确时段、跨日标识与原表填写的工作等级。
- 两组独立开关，默认收起；关闭悬浮卡后重置，不打扰下一次查看当下。
- 点击预览行查看完整原始建议，不替换当前曜星、时辰或任务栏显示。
- 展开后可安心阅读；点任务栏按钮关闭，长内容滚轮浏览。
- 标签依据真实原表：午时是午餐休息，寅时是继续睡眠；没有填写的等级不会补造。
- 修改 Excel 后用当前文本替代旧摘要，中英文同步适配。

Two independent look-ahead drawers, original identities and element colors, date-aware time ranges, source-based keywords and click-to-expand guidance. Previews reset whenever the popup closes. Current guidance and the transparent taskbar remain unchanged.

## 升级 / Upgrading

设置目录仍为 `%APPDATA%\Astral`。复制自行修改的“表格”文件夹到新版本；请保留自己的文件备份。

Existing settings remain compatible. Copy customized workbooks into the new package. No configuration reset is required.

## 验证 / Verification

Portable tests cover every minute of the day, calendar boundaries, short-label provenance, independent session state and 16,800 preview layouts. Windows CI additionally checks actual GDI font metrics and exports native interface images. These are not full desktop taskbar-interaction tests. The executable is not code-signed.

附带完整源码及 SHA-256 校验文件 / Source and SHA-256 checksums are included.
