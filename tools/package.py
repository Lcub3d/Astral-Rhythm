"""Package reviewed project files. No credentials, local settings or fonts are included."""
from __future__ import annotations

import argparse
import hashlib
from pathlib import Path
import re
import zipfile

ROOT = Path(__file__).resolve().parents[1]
TOP_FILES = (
    "AstralRhythm.ico", "README.md", "README.en.md", "README.zh-CN.md", "CHANGELOG.md", "RELEASE_NOTES.md",
    "BUILD_SOURCES.md", "AGENTS.md", "先读我.txt", "使用说明.md", "data.json",
    ".gitignore", ".gitattributes", "release.json", "Start-English.cmd", "Start-Chinese.cmd", "Start-System.cmd",
    "Preview.cmd", "Reset-Position.cmd", "Self-Test.cmd", "打开悬浮卡.cmd", "恢复默认位置.cmd", "自检.cmd",
)
DIRECTORIES = ("src", "tools", ".agents", ".github", "locales", "表格", "docs")
ALLOWED = {".go", ".mod", ".sum", ".json", ".md", ".txt", ".cmd", ".py", ".png", ".svg", ".syso", ".yml", ".yaml", ".xlsx", ".ico"}


def selected_files(binary: bool) -> list[Path]:
    paths = [ROOT / name for name in TOP_FILES if (ROOT / name).is_file()]
    for directory in DIRECTORIES:
        paths += [p for p in (ROOT / directory).rglob("*")
                  if p.is_file() and not p.is_symlink() and p.suffix in ALLOWED
                  and "__pycache__" not in p.parts and not p.name.startswith("~$")]
    if binary:
        exe = ROOT / "AstralRhythm.exe"
        if not exe.is_file() or exe.read_bytes()[:2] != b"MZ":
            raise SystemExit("Build the Windows AstralRhythm.exe before creating a portable package.")
        paths.append(exe)
        if (ROOT / "validation").exists():
            paths += [p for p in (ROOT / "validation").iterdir()
                      if p.is_file() and p.suffix in {".log", ".txt", ".json"}]
    # Reject a symlink instead of silently following it outside the source tree.
    for p in paths:
        if p.is_symlink() or not p.resolve().is_relative_to(ROOT.resolve()):
            raise SystemExit(f"Unsafe package path: {p}")
    return sorted(set(paths), key=lambda p: p.relative_to(ROOT).as_posix())


def package(output: Path, binary: bool, version: str) -> Path:
    paths = selected_files(binary)
    label = "Windows_x64" if binary else "Source"
    target = output / f"Astral-Rhythm_v{version}_{label}.zip"
    prefix = "Astral-Rhythm"
    manifest = "".join(f"{hashlib.sha256(p.read_bytes()).hexdigest()}  {p.relative_to(ROOT).as_posix()}\n" for p in paths)
    with zipfile.ZipFile(target, "w", zipfile.ZIP_DEFLATED, compresslevel=9) as z:
        for p in paths:
            z.write(p, f"{prefix}/{p.relative_to(ROOT).as_posix()}")
        if binary:
            z.writestr(f"{prefix}/SHA256.txt", manifest)
    return target


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", type=Path, default=ROOT / "dist")
    parser.add_argument("--source-only", action="store_true")
    args = parser.parse_args()
    match = re.search(r'const appVersion = "([0-9]+\.[0-9]+\.[0-9]+)"', (ROOT / "src/locale.go").read_text(encoding="utf-8"))
    if not match:
        raise SystemExit("Cannot find a semantic appVersion in src/locale.go")
    args.output.mkdir(parents=True, exist_ok=True)
    outputs = [package(args.output, False, match[1])]
    if not args.source_only:
        outputs.insert(0, package(args.output, True, match[1]))
    hashes = ""
    for p in outputs:
        digest = hashlib.sha256(p.read_bytes()).hexdigest()
        hashes += f"{digest}  {p.name}\n"
        print(f"{p.name}: {p.stat().st_size:,} bytes; SHA256 {digest}")
    (args.output / "Astral-Rhythm_SHA256.txt").write_text(hashes, encoding="ascii")


if __name__ == "__main__":
    main()
