#Requires -Version 5.1
<#
.SYNOPSIS
    Install simple-cli on Windows and register it on the user PATH.

.DESCRIPTION
    Downloads the latest simple-cli Windows binary, installs it to
    %LOCALAPPDATA%\simple-cli\bin, and idempotently registers that directory
    in the user-scope PATH environment variable.

    When run as Administrator the binary is installed to
    %ProgramFiles%\simple-cli\bin and registered in the machine-scope PATH.

    Constitution Principle III: installer MUST register ENV PATH and be
    idempotent (running twice MUST NOT create duplicate PATH entries).

.EXAMPLE
    # User install (no elevation required):
    irm https://github.com/your-org/simple-cli/releases/latest/download/install.ps1 | iex

    # Machine-wide install (run as Administrator):
    irm https://...install.ps1 | iex
#>

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$Repo    = 'your-org/simple-cli'
$Binary  = 'simple-cli.exe'

# ──────────────────────────────────────────────
# Helpers
# ──────────────────────────────────────────────

function Write-Log    { Write-Host "[install] $args" -ForegroundColor Green }
function Write-Warn   { Write-Host "[install] $args" -ForegroundColor Yellow }

function Get-LatestTag {
    $api = "https://api.github.com/repos/$Repo/releases/latest"
    (Invoke-RestMethod -Uri $api -UseBasicParsing).tag_name
}

function Get-InstallDir {
    if (([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole(
            [Security.Principal.WindowsBuiltInRole]::Administrator)) {
        return Join-Path $env:ProgramFiles 'simple-cli\bin'
    }
    return Join-Path $env:LOCALAPPDATA 'simple-cli\bin'
}

function Get-PathScope {
    if (([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole(
            [Security.Principal.WindowsBuiltInRole]::Administrator)) {
        return [System.EnvironmentVariableTarget]::Machine
    }
    return [System.EnvironmentVariableTarget]::User
}

# ──────────────────────────────────────────────
# PATH registration (idempotent)
# ──────────────────────────────────────────────

function Register-Path {
    param([string]$Dir, [System.EnvironmentVariableTarget]$Scope)

    $current = [System.Environment]::GetEnvironmentVariable('PATH', $Scope)
    $entries = $current -split ';' | Where-Object { $_ -ne '' }

    if ($entries -contains $Dir) {
        Write-Log "$Dir already in PATH ($Scope scope) — skipping"
        return
    }

    $newPath = ($entries + $Dir) -join ';'
    [System.Environment]::SetEnvironmentVariable('PATH', $newPath, $Scope)
    Write-Log "Registered $Dir in PATH ($Scope scope)"

    # Broadcast WM_SETTINGCHANGE so running Explorer / shells pick up the change.
    if ($Scope -eq [System.EnvironmentVariableTarget]::Machine) {
        $code = @'
using System;
using System.Runtime.InteropServices;
public class EnvBroadcast {
    [DllImport("user32.dll", SetLastError=true, CharSet=CharSet.Auto)]
    public static extern IntPtr SendMessageTimeout(IntPtr hWnd, uint Msg,
        UIntPtr wParam, string lParam, uint fuFlags, uint uTimeout, out UIntPtr lpdwResult);
    public static void Broadcast() {
        UIntPtr result;
        SendMessageTimeout((IntPtr)0xffff, 0x001A, UIntPtr.Zero, "Environment",
            0x0002, 5000, out result);
    }
}
'@
        Add-Type -TypeDefinition $code -Language CSharp -ErrorAction SilentlyContinue
        [EnvBroadcast]::Broadcast()
    }
}

# ──────────────────────────────────────────────
# Main
# ──────────────────────────────────────────────

$tag        = Get-LatestTag
$installDir = Get-InstallDir
$scope      = Get-PathScope
$dest       = Join-Path $installDir $Binary

Write-Log "Installing simple-cli $tag to $installDir ..."
New-Item -ItemType Directory -Force -Path $installDir | Out-Null

$url = "https://github.com/$Repo/releases/download/$tag/simple-cli_windows_amd64.zip"
$zip = Join-Path $env:TEMP 'simple-cli.zip'

Invoke-WebRequest -Uri $url -OutFile $zip -UseBasicParsing
Expand-Archive -Path $zip -DestinationPath $installDir -Force
Remove-Item $zip -ErrorAction SilentlyContinue

if (-not (Test-Path $dest)) {
    throw "Binary not found at $dest after extraction."
}
Write-Log "Installed $Binary → $dest"

Register-Path -Dir $installDir -Scope $scope

# Post-install validation in a fresh environment.
Write-Log "Validating installation..."
$env:PATH = "$installDir;$env:PATH"
$v = & $dest --version 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Log "✓ Installation successful: $v"
} else {
    Write-Warn "Binary installed but --version failed."
    Write-Warn "Open a new PowerShell session and run: simple-cli --version"
}
