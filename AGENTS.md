# Astral Rhythm / 七曜工作法

Read `.agents/skills/astral-ui/SKILL.md` before UI changes.

Chinese name: 七曜工作法. English name: Astral Rhythm. Tagline: Work in rhythm. Executable: AstralRhythm.exe.
Keep the stable appID Astral and settings directory for backwards compatibility.
The upper heading is 曜星 / Day Star. Keep day-star and shichen content separate.
All added user-facing text must be bilingual. Exact-source English overrides must not silently apply stale translations to edited workbooks.
Preserve original source workbooks. Never add advice invented for a mockup to the source data.
Keep the transparent taskbar companion and approved compact spacing. No persistent footer or disclaimer.
Run `cd src && go test ./...`, cross-compile Windows x64, regenerate version resources when needed. Report actual tests honestly: layout estimates and Linux tests are not Windows desktop evidence.
Legacy class, mutex, settings and startup key names may stay only for compatibility.
Do not commit local settings, credentials, fonts, personal screenshots or executable build outputs. No open-source license has been selected by the owner.

Public releases start at 1.0.0. Settings UIVersion=3 is a migration schema; do not reset it. Documentation images must be exported from the native renderer, never presented as desktop screenshots.
