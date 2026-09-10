# Astral Rhythm · 七曜工作法

**Work in rhythm.**

A compact, offline Windows taskbar companion: weekly direction from the seven day stars, and guidance for the current traditional two-hour period (*shichen*).

**English · [简体中文](README.zh-CN.md)**

## Start

Open [Windows builds](https://github.com/Lcub3d/Astral-Rhythm/actions/workflows/windows.yml), select a successful run, and download **Astral-Rhythm-Windows-x64-and-source** under **Artifacts**. GitHub requires sign-in for Actions artifact downloads. Unpack the **Windows x64 ZIP** inside. Published versions, when available, can also be downloaded from [Releases](https://github.com/Lcub3d/Astral-Rhythm/releases).

Extract the entire folder, close the previous version, and run **`AstralRhythm.exe`**. The executable needs no separately installed Go, Python, .NET or WebView runtime. **Code → Download ZIP** is source code, not the runnable app.

The transparent taskbar companion shows the **day star above** and **shichen below**. Hover to open the card; click to pin or close; drag to reposition; right-click for settings.

## A quiet interface

Two compact, independent cards separate weekly direction from the current time period. Day-star artwork changes with the weekday; shichen keeps its own mountain-and-mist artwork. Dark and light themes are included.

The taskbar button has **no visible background plate or border**. Its two text colors independently follow their five-element families. Long guidance wraps and scrolls. There is no persistent footer or toolbar.

## 中文 / English

Right-click → **Language → System / 简体中文 / English**. The change applies immediately and is saved. Chinese Windows interface languages select Simplified Chinese; other system languages select English.

`Start-English.cmd`, `Start-Chinese.cmd` and `Start-System.cmd` select a language directly. Menus, guidance cards, bundled advice, traditional explanations and copied text are bilingual. The upper heading is **Day Star / 曜星**.

## Make it yours

Edit either guidance workbook in **`表格/`**. Saved changes are reloaded in about five seconds. Keep filenames, headers and weekday/shichen identifiers. Invalid changes do not replace the last valid data. Public copies remove local-path and author metadata only; worksheet content and formatting are preserved.

Translations live in **`locales/en.json`**, keyed by exact Chinese source text. There is no translation service or network request. New text without a matching translation remains in its original language; add its translation to the local dictionary. Changing languages never rewrites the workbooks.

Data priority: workbooks → `data.json` when both workbooks are absent → embedded data.

## Upgrade and settings

The product name is **Astral Rhythm**. The settings directory remains **`%APPDATA%\Astral`** for compatibility with V3.1.0. Older 七曜时辰 preferences can be migrated when no Astral settings exist. Old settings are not deleted. Copy your customized `表格` folder into the new portable package when upgrading.

`Preview.cmd` opens the card. `Reset-Position.cmd` restores placement. `Self-Test.cmd` writes `self-test.txt`.

## Build

Go + Win32, no third-party Go dependencies, no CGO. Use Go 1.23 or newer, preferably a supported release for distribution.

```sh
cd src
go test -v ./...
go vet ./...
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-H=windowsgui -s -w" -o ../AstralRhythm.exe .
```

On Windows, run `src/build.cmd`. Regenerate icon/version/DPI resources with `python tools/build_resources.py`. Artwork is checked in; rebuilding it with `tools/build_assets.py` additionally needs Pillow and CairoSVG. These are development tools, not runtime requirements. No fonts are bundled; the app uses Windows system fonts.

`python tools/package.py` creates portable and source ZIPs plus SHA-256 checksums under `dist/`. [GitHub Actions](https://github.com/Lcub3d/Astral-Rhythm/actions) runs Linux tests/vet, native Windows tests and a Windows x64 build. Passing runs provide downloadable artifacts. Pushing a version tag matching the source version runs the tested release workflow; publication requires the corresponding repository permissions. Pull requests never receive release-write permissions.

## Scope

The widget is a separate native window attached to the **primary taskbar**; it does not reserve tray space or move other icons. Time follows Windows local civil time. Day stars change at midnight; Zi spans 23:00–01:00.

Automated layout tests estimate font widths. Passing tests and compilation do not establish visual verification on a real Windows desktop. The EXE is not digitally signed. No open-source license has been selected by the owner.
