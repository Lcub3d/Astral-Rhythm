# Look-ahead drawers / 未来预览

## Interaction

The day card contains **未来 3 日 / Next 3 days**. The shichen card contains **后续时辰 / Next 4 shichen**. Both controls stay visible, but their future rows are hidden until explicitly clicked. Hovering alone never expands either drawer.

Rows retain their identities; color is redundant to the name/keyword. Click a row to read its complete source advice. Drawer and row expansion is session-only, not written to settings. Closing/reopening, quitting or restarting the application returns to collapsed previews. Previously saved day/hour detail preferences are independent.

While a drawer is open, moving the mouse off the card does not dismiss it immediately. Clicking the taskbar companion closes it. Clicking outside dismisses an unpinned preview when detected; an explicitly pinned card stays pinned. Escape also dismisses the popup. None of these interactions activate the card or replace the foreground typing window. Long content uses the existing scrollbar and mouse wheel.

## Calendar

The day preview begins tomorrow and includes the next three calendar dates. The hour preview begins at the next two-hour boundary and includes four consecutive shichen, never repeating the current one. Zi spans 23:00–01:00; crossing-midnight intervals explicitly mark their ending day. Upcoming periods starting on the next date carry 明日 / Tomorrow.

The implementation uses the same displayed Windows civil clock as the rest of the app. It does not apply true-solar-time corrections or predict elapsed durations across daylight-saving changes. Clock/timezone changes refresh the lists; row selections reset when their relative calendar slot changes.

## Data provenance

`src/preview.go` contains reviewable Chinese/English short summaries, matched against the exact original advice cell. Night-time labels encourage sleeping; Wu encourages lunch/rest. Priority values are copied only from populated source cells. No new P-levels or auspiciousness claims are introduced.

When a workbook cell changes, the summary falls back to an excerpt of that current cell (or its exact-match English translation). The complete current cell remains available by expanding the row. Empty translation matches do not erase Chinese content. No network services are used; the source workbooks are not rewritten.

## Verification

`preview_test.go`: civil date/time boundaries, all 1,440 minute positions, sleep/lunch provenance, session isolation, custom text, overflow, and 16,800 portable preview layouts.

`preview_windows_test.go`: 672 expanded layouts using actual Windows GDI text metrics; optional exports for both languages and light/dark modes. Reproduce with `ASTRAL_DOCS_DIR` and `go test -v ./...` on Windows. Native exported renders are not desktop screenshots or full interaction certification.
