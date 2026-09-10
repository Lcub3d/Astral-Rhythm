# 界面图从哪里来？ · Image provenance

`card-*.png` 与 `taskbar-*.png` 由 Windows CI 调用应用本身的 `paintCard`、`paintWidgetAt`、GDI 字体度量和逐像素透明合成代码导出。固定示例时间为 **2026-09-08 13:40**：火曜 / Mars，未时 / Wei。深浅两套、中英两种语言，DPI 设为 192（200%）。

这是**原生离屏绘制**，不是桌面截图、交互录像或 AI 概念图。它反映该次构建的真实排版和素材，不证明全部 Windows 版本、DPI 与任务栏交互组合已经实机测试。

`hero.png` 与 `identities.png` 由 `tools/build_docs.py` 将原生界面与项目原有图标合成；山岚、宣传语、展示底座属于文档设计，不是新增软件功能。透明任务栏控件在封面上放入示意底座，底座不属于软件按钮。

No personal desktop screenshot is used. No font file is committed, packaged or redistributed. Native exports use fonts installed on the Windows build machine; the composer also uses system fonts.

## Reproduce on Windows

```powershell
# Repository root; Pillow is a development-only dependency.
python -m pip install Pillow==12.3.0
cd src
$env:ASTRAL_DOCS_DIR = '../docs/images'
go test -run TestNativeDocumentationRenders -v
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
cd ..
python tools/build_docs.py
```

`render-metadata.json` records each native export's dimensions, fixture time, DPI and renderer. CI checks measured text against its allocated width and the transparent button's background alpha.
