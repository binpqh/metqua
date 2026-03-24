; setup.nsi — NSIS installer for simple-cli on Windows.
; Constitution Principle III: builds signed NSIS .exe installer with machine-scope
; PATH registration and a proper uninstaller section.
;
; Prerequisites:
;   NSIS >= 3.09  (https://nsis.sourceforge.io/)
;   EnvVarUpdate.nsh  (bundled with NSIS Modern UI)
;
; Build:
;   makensis installer/windows/setup.nsi
;
; The binary for the current architecture (amd64) must be present at:
;   dist/simple-cli.exe

!include "MUI2.nsh"
!include "EnvVarUpdate.nsh"

; ──────────────────────────────────────────────
; Metadata
; ──────────────────────────────────────────────

!define APPNAME     "simple-cli"
!define APPVERSION  "1.0.0"
!define PUBLISHER   "your-org"
!define BINARY      "simple-cli.exe"
!define UNINST_KEY  "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}"

Name            "${APPNAME} ${APPVERSION}"
OutFile         "dist\simple-cli_windows_amd64_setup.exe"
InstallDir      "$PROGRAMFILES64\${APPNAME}"
InstallDirRegKey HKLM "${UNINST_KEY}" "InstallLocation"
RequestExecutionLevel admin

; ──────────────────────────────────────────────
; MUI pages
; ──────────────────────────────────────────────

!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES

!insertmacro MUI_LANGUAGE "English"

; ──────────────────────────────────────────────
; Install Section
; ──────────────────────────────────────────────

Section "Install" SecInstall
  SetOutPath "$INSTDIR"
  File "dist\${BINARY}"

  ; Register machine-scope PATH entry (idempotent via EnvVarUpdate).
  ${EnvVarUpdate} $0 "PATH" "A" "HKLM" "$INSTDIR"

  ; Write uninstaller and registry key.
  WriteUninstaller "$INSTDIR\Uninstall.exe"
  WriteRegStr HKLM "${UNINST_KEY}" "DisplayName"      "${APPNAME}"
  WriteRegStr HKLM "${UNINST_KEY}" "DisplayVersion"   "${APPVERSION}"
  WriteRegStr HKLM "${UNINST_KEY}" "Publisher"        "${PUBLISHER}"
  WriteRegStr HKLM "${UNINST_KEY}" "InstallLocation"  "$INSTDIR"
  WriteRegStr HKLM "${UNINST_KEY}" "UninstallString"  '"$INSTDIR\Uninstall.exe"'
  WriteRegDWORD HKLM "${UNINST_KEY}" "NoModify" 1
  WriteRegDWORD HKLM "${UNINST_KEY}" "NoRepair" 1

  ; Post-install validation.
  ExecWait '"$INSTDIR\${BINARY}" --version' $0
  StrCmp $0 "0" +2
    MessageBox MB_ICONINFORMATION "Installation complete, but --version validation failed.$\nPlease open a new terminal and run: simple-cli --version"
SectionEnd

; ──────────────────────────────────────────────
; Uninstall Section
; ──────────────────────────────────────────────

Section "Uninstall"
  ; Remove machine-scope PATH entry.
  ${un.EnvVarUpdate} $0 "PATH" "R" "HKLM" "$INSTDIR"

  Delete "$INSTDIR\${BINARY}"
  Delete "$INSTDIR\Uninstall.exe"
  RMDir  "$INSTDIR"

  DeleteRegKey HKLM "${UNINST_KEY}"
SectionEnd
