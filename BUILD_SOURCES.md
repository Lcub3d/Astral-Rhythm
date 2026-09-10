# Implementation references

These are official API/action references consulted for the build configuration and native localization. They are not sources for the workbook's work guidance.

- Windows user UI language: https://learn.microsoft.com/en-us/windows/win32/api/winnls/nf-winnls-getuserdefaultuilanguage
- Layered windows: https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-updatelayeredwindow
- Font creation: https://learn.microsoft.com/en-us/windows/win32/api/wingdi/nf-wingdi-createfontw
- Checkout action: https://github.com/actions/checkout
- Go setup action: https://github.com/actions/setup-go
- Build artifact action: https://github.com/actions/upload-artifact
- Python setup action: https://github.com/actions/setup-python

The data source remains the two user-supplied Excel workbooks; their original bytes are bundled unchanged.

Public workbook copies preserve worksheet and formatting parts. Only embedded local-path, author and document-instance metadata is removed before publication. Original conversation uploads are not modified.
