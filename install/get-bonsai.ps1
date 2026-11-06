<#
get-bonsai.ps1: installs a Bonsai release on this Windows machine in one command, with every check the install page's
lines make (docs/install.md), in the same order, stopping at the first that fails. Written for Windows PowerShell 5.1.

What it does: checks this machine first (an amd64 or arm64 Windows; Windows' own curl.exe; gh, GitHub's command line,
2.97 or newer and logged in); takes the newest release (gh api repos/LastStep/Bonsai/releases/latest) or the version
given; in a new folder of its own under %TEMP% (its path printed and left for you to read), downloads the release's
archive for this machine (bonsai_<version>_windows_<amd64|arm64>.zip) and its checksums.txt; checks the archive's
SHA-256 against the one line checksums.txt holds for it; checks where it was built (gh attestation verify: Bonsai's own
release workflow, from that version's tag, on GitHub's machines); only then unpacks it and runs the release's own
install.ps1, which asks one Windows question ("allow this app to make changes") and puts bonsai at
C:\Program Files\Bonsai\bonsai.exe. It has no option that skips a check. curl.exe and powershell.exe are run from
Windows' own folder by their full paths; gh is the first the PATH names (where.exe gh).

Usage, in a normal PowerShell (not as administrator), fetch it, read it, run it:
  curl.exe -fsSLO https://raw.githubusercontent.com/LastStep/Bonsai/main/install/get-bonsai.ps1
  powershell -NoProfile -ExecutionPolicy Bypass -File .\get-bonsai.ps1                  # the newest release
  powershell -NoProfile -ExecutionPolicy Bypass -File .\get-bonsai.ps1 -Version 1.0.3   # one release (v1.0.3 too)
(Windows runs no downloaded script otherwise.) A pre-release is named as 1.0.0-rc.1.

Needs: Windows 10 or later (its curl.exe), Windows PowerShell 5.1, and gh 2.97 or newer, logged in (gh auth login).

Exit 0 when the release's install.ps1 is done; otherwise 1, with one plain sentence and its next step, and nothing
installed unless the failure came from install.ps1 itself (it says what it did).

The way back: run it again with the release you had, as: -Version 1.0.2; the way out: the release's own
install.ps1 -Remove (docs/install.md, "The way out").
#>
[CmdletBinding()]
param(
    [Parameter(Position = 0)]
    [string]$Version = ''
)
Set-StrictMode -Version 2
$ErrorActionPreference = 'Stop'

$Repo = 'LastStep/Bonsai'
$Workflow = "$Repo/.github/workflows/release.yml"
$Windows = Join-Path $env:SystemRoot 'System32'

function Stop-Get {
    param([string]$What, [string]$Next)
    [Console]::Error.WriteLine('get-bonsai.ps1: ' + $What)
    [Console]::Error.WriteLine('next: ' + $Next)
    exit 1
}

# Get-Version is its argument without a leading v when it is a version this script installs (X.Y.Z or X.Y.Z-rc.N,
# 1.0.0 or newer), else ''.
function Get-Version {
    param([string]$Text)
    $v = $Text
    if ($v.StartsWith('v')) { $v = $v.Substring(1) }
    if ($v -cmatch '\A[0-9]+\.[0-9]+\.[0-9]+(-rc\.[0-9]+)?\z' -and -not $v.StartsWith('0.')) { return $v }
    return ''
}

# Invoke-Native runs a program with its output shown and returns its exit code; Windows PowerShell's Stop would
# otherwise turn a line on its error stream into a failure.
function Invoke-Native {
    param([string]$Exe, [string[]]$Arguments)
    $old = $ErrorActionPreference
    $ErrorActionPreference = 'Continue'
    try {
        & $Exe @Arguments | Out-Host
        return $LASTEXITCODE
    } finally {
        $ErrorActionPreference = $old
    }
}

# Read-Native runs a program with its error stream dropped and returns its exit code (Code) and its output's lines
# (Lines, an array, empty for none).
function Read-Native {
    param([string]$Exe, [string[]]$Arguments)
    $old = $ErrorActionPreference
    $ErrorActionPreference = 'Continue'
    try {
        $out = @(& $Exe @Arguments 2>$null | ForEach-Object { [string]$_ })
        return New-Object PSObject -Property @{ Code = $LASTEXITCODE; Lines = $out }
    } finally {
        $ErrorActionPreference = $old
    }
}

# A version given is read before anything else.
$given = $PSBoundParameters.ContainsKey('Version')
if ($given) {
    $ver = Get-Version $Version
    if ($ver -eq '') {
        if ($Version -cnotmatch '\A[0-9a-z.v-]+\z') {
            Stop-Get 'the version given is not a release version such as 1.0.3 (a version holds only digits, dots and -rc.N)' "run it again with a version from https://github.com/$Repo/releases, or with no -Version for the newest"
        }
        if ($Version -cmatch '\Av?0\.[0-9]') {
            Stop-Get "'$Version' is the old product (0.x), which this script does not install" "read its own install notes at https://github.com/$Repo/tree/v0.4.3, or run it with no -Version for the newest release"
        }
        Stop-Get "'$Version' is not a release version such as 1.0.3 (or 1.0.0-rc.1)" "run it again with a version from https://github.com/$Repo/releases, or with no -Version for the newest"
    }
}

# A 32-bit PowerShell on a 64-bit Windows names the machine in PROCESSOR_ARCHITEW6432.
$machine = $env:PROCESSOR_ARCHITEW6432
if (-not $machine) { $machine = $env:PROCESSOR_ARCHITECTURE }
switch ($machine) {
    'AMD64' { $arch = 'amd64' }
    'ARM64' { $arch = 'arm64' }
    default { Stop-Get "there is no Bonsai build for this machine ($machine); builds are amd64 and arm64" 'build it from source (go build ./cmd/bonsai) and copy it by hand, or use an amd64 or arm64 machine' }
}

$curl = Join-Path $Windows 'curl.exe'
if (-not (Test-Path -LiteralPath $curl -PathType Leaf)) {
    Stop-Get "Windows' own curl.exe is not at $curl (it comes with Windows 10 and later)" 'run this on Windows 10 or later, or install with the lines in docs/install.md'
}
$powershell = Join-Path $Windows 'WindowsPowerShell\v1.0\powershell.exe'
if (-not (Test-Path -LiteralPath $powershell -PathType Leaf)) {
    Stop-Get "Windows PowerShell is not at $powershell" 'run this on Windows 10 or later, where Windows PowerShell 5.1 comes with Windows'
}
$found = @(Get-Command gh -CommandType Application -ErrorAction SilentlyContinue)
if ($found.Count -eq 0) {
    Stop-Get "gh, GitHub's command line, is not on the PATH; the provenance check needs it" 'install gh 2.97 or newer (https://cli.github.com), run gh auth login in a new PowerShell, then run this again'
}
$gh = $found[0].Source

# gh 2.97 or newer: --signer-workflow came in 2.51, --source-ref in 2.68, and 2.97 fixed GHSA-mm27-mwq9-fr5g.
$ghv = ''
$r = Read-Native $gh @('--version')
if ($r.Code -eq 0 -and $r.Lines.Count -gt 0 -and $r.Lines[0] -cmatch '\Agh version ([0-9]+)\.([0-9]+)') {
    $ghv = $Matches[1] + '.' + $Matches[2]
    $ok = ([int]$Matches[1] -gt 2) -or ([int]$Matches[1] -eq 2 -and [int]$Matches[2] -ge 97)
} else {
    $ok = $false
}
if (-not $ok) {
    if ($ghv -eq '') { $ghv = 'of a version gh --version does not name' }
    Stop-Get "gh is $ghv, older than 2.97, whose provenance check a lookalike workflow name could pass" 'update gh to 2.97 or newer (https://cli.github.com), then run this again'
}
if ((Read-Native $gh @('auth', 'status')).Code -ne 0) {
    Stop-Get 'gh is not logged in to GitHub; the provenance check needs it' 'run gh auth login, then run this again'
}

if (-not $given) {
    $r = Read-Native $gh @('api', "repos/$Repo/releases/latest", '--jq', '.tag_name')
    $text = ''
    if ($r.Code -eq 0 -and $r.Lines.Count -eq 1) { $text = $r.Lines[0].Trim() }
    $ver = Get-Version $text
    if ($ver -eq '') {
        Stop-Get "could not read the newest release's version from GitHub (gh api gave '$text')" "check gh can reach GitHub (gh api repos/$Repo/releases/latest), or run it with a version, as: -Version 1.0.3"
    }
}
$tag = 'v' + $ver
$archive = "bonsai_${ver}_windows_${arch}.zip"
$base = "https://github.com/$Repo/releases/download/$tag"

try {
    $dir = (New-Item -ItemType Directory (Join-Path $env:TEMP ('bonsai-' + [guid]::NewGuid()))).FullName
} catch {
    Stop-Get "could not make a folder under $env:TEMP" 'check your temporary folder can be written, then run this again'
}
Set-Location -LiteralPath $dir
Write-Host "Bonsai $ver for windows_$arch from $Repo, in $dir"

Write-Host "Downloading $archive and checksums.txt:"
if ((Invoke-Native $curl @('-fsSLO', "$base/$archive")) -ne 0) {
    Stop-Get "downloading $base/$archive failed; nothing was installed" "check the version is on https://github.com/$Repo/releases and has a windows_$arch archive, then run this again"
}
if ((Invoke-Native $curl @('-fsSLO', "$base/checksums.txt")) -ne 0) {
    Stop-Get "downloading $base/checksums.txt failed; nothing was installed" 'check your connection to github.com, then run this again'
}

Write-Host "Checking the archive's SHA-256:"
$want = @(Get-Content -LiteralPath (Join-Path $dir 'checksums.txt') | ForEach-Object {
        $f = $_ -split '\s+'
        if ($f.Count -ge 2 -and ($f[1] -ceq $archive -or $f[1] -ceq ('*' + $archive))) { $f[0].ToLowerInvariant() }
    })
if ($want.Count -ne 1) {
    Stop-Get "checksums.txt does not hold exactly one line for $archive; nothing was installed" "report it at https://github.com/$Repo/issues; do not install this archive"
}
$s = [IO.File]::OpenRead((Join-Path $dir $archive))
try { $h = [Security.Cryptography.SHA256]::Create().ComputeHash($s) } finally { $s.Dispose() }
$got = ([BitConverter]::ToString($h) -replace '-', '').ToLowerInvariant()
if ($got -cne $want[0]) {
    Stop-Get "$archive's SHA-256 is $got, not $($want[0]) as checksums.txt holds; nothing was installed" "run this again for a new download; if it fails again, report it at https://github.com/$Repo/issues"
}
Write-Host "${archive}: OK"

Write-Host 'Checking where it was built:'
if ((Invoke-Native $gh @('attestation', 'verify', $archive, '-R', $Repo, '--signer-workflow', $Workflow, '--source-ref', "refs/tags/$tag", '--deny-self-hosted-runners')) -ne 0) {
    Stop-Get "gh attestation verify did not pass for $archive; nothing was installed" "read gh's message above; if gh can reach GitHub and it fails again, report it at https://github.com/$Repo/issues"
}
Write-Host "${archive}: built by $Workflow from refs/tags/$tag"

Write-Host "Every check passed: unpacking $archive and running its install.ps1"
try {
    Expand-Archive -LiteralPath (Join-Path $dir $archive) -DestinationPath $dir -Force
} catch {
    Stop-Get "unpacking $archive failed ($($_.Exception.Message)); nothing was installed" "check $dir has room, then run this again"
}
$installer = Join-Path $dir 'install.ps1'
if (-not (Test-Path -LiteralPath $installer -PathType Leaf)) {
    Stop-Get "$archive holds no install.ps1; nothing was installed" "report it at https://github.com/$Repo/issues"
}
if ((Invoke-Native $powershell @('-NoProfile', '-ExecutionPolicy', 'Bypass', '-File', $installer)) -ne 0) {
    Stop-Get "the release's install.ps1 stopped (its message above says what it did)" 'follow its next step, or read docs/install.md, "When a line fails"'
}
exit 0
