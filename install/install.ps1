<#
install.ps1: installs the bonsai.exe beside this script as this machine's Bonsai, on Windows (Windows PowerShell 5.1).

What it does: checks that the bonsai.exe in this script's folder runs (bonsai.exe --version) and takes its SHA-256.
At the real place, C:\Program Files\Bonsai\bonsai.exe, the administrator's steps then run in one elevated child (one
Windows prompt), or with no prompt when this PowerShell is already elevated: the folder made and its rights checked
(no right that writes for Users, Authenticated Users or Everyone); the file copied in as bonsai.exe.new, that copy's
SHA-256 checked against the one taken here (a source changed meanwhile is caught, and the copy removed) and its rights
checked the same way (a copy that fails is removed: nothing lands in a folder others can write); old copies set aside
by an earlier install removed; the old bonsai.exe, which a hook may be running, renamed aside to bonsai.exe.old-<n>
(Windows lets a running program be renamed, not overwritten), the new copy renamed into place (a hook starting between
the two renames finds no file and refuses, never runs another) and the old one deleted unless it is still running
(then the next install deletes it); C:\Program Files\Bonsai added once to the machine's PATH (its %...% entries kept as
they are) and the change announced to running programs. The child's code is inline (-EncodedCommand), never a file,
and it compiles and loads no code. Then, as you, it checks the installed file's SHA-256 again and writes the
installed copy's record, install.json, in the Bonsai home ($env:BONSAI_HOME, else %USERPROFILE%\.bonsai): its full path
(with forward slashes; a short name in it, such as RUNNER~1, made long), version and sha256, the record bonsai check
and bonsai status compare the bonsai on the PATH with. It prints the fingerprint, every bonsai on the PATH in order
and --version.

When to run it: once per Bonsai release, by the person, in a normal PowerShell, from the release's folder:
  powershell -NoProfile -ExecutionPolicy Bypass -File .\install.ps1
(Windows runs no downloaded script otherwise.) No agent runs it at the real place.

Options:
  -Target <file>  install to <file> instead of C:\Program Files\Bonsai\bonsai.exe: an absolute path with a drive
                  letter; no prompt, no PATH change and no rights check there, the same copy, hash and rename steps.
                  Example: powershell -NoProfile -ExecutionPolicy Bypass -File .\install.ps1 -Target C:\scratch\bonsai.exe
  -Remove         undo each step: the installed file and the copies set aside, the folder at the real place if it is
                  then empty, its PATH entry, and install.json. Example:
                    powershell -NoProfile -ExecutionPolicy Bypass -File .\install.ps1 -Remove
  -ForceChild     for tests only: run the steps in the elevated child even when this PowerShell is already elevated
                  (a test machine with no prompt), so the child's code is what runs. Refused unless this PowerShell is
                  elevated, so it never raises a prompt; it skips no step and no check.

Exit 0 when it is done; otherwise 1, with one plain sentence and its next step.

The way out: install.ps1 -Remove; or by hand, an administrator deletes C:\Program Files\Bonsai, takes it off the
machine's PATH (System Properties, Environment Variables) and deletes %USERPROFILE%\.bonsai\install.json.
#>
[CmdletBinding()]
param(
    [string]$Target = 'C:\Program Files\Bonsai\bonsai.exe',
    [switch]$Remove,
    [switch]$ForceChild
)
Set-StrictMode -Version 2
$ErrorActionPreference = 'Stop'

$Place = 'C:\Program Files\Bonsai\bonsai.exe'

# The steps the child's exit code names (0 is done).
$Steps = @{
    10 = 'making the folder'
    11 = 'copying bonsai.exe in as bonsai.exe.new'
    12 = 'checking the copy: its SHA-256 is not the checked file''s (the source changed between the check and the copy), so the copy was removed and nothing was installed'
    14 = 'renaming the old bonsai.exe aside'
    15 = 'renaming the new copy into place'
    16 = 'adding the folder to the machine''s PATH'
    17 = 'announcing the PATH change to running programs'
    18 = 'checking the folder''s rights: Users, Authenticated Users or Everyone can write the folder or the copy, so nothing was installed'
    20 = 'removing the installed file or a copy set aside (one may still be running)'
    21 = 'taking the folder off the machine''s PATH'
}

function Stop-Install {
    param([string]$What, [string]$Next)
    [Console]::Error.WriteLine('install.ps1: ' + $What)
    [Console]::Error.WriteLine('next: ' + $Next)
    exit 1
}

function Get-Sha256 {
    param([string]$Path)
    $s = [IO.File]::OpenRead($Path)
    try { $h = [Security.Cryptography.SHA256]::Create().ComputeHash($s) } finally { $s.Dispose() }
    return ([BitConverter]::ToString($h) -replace '-', '').ToLowerInvariant()
}

# Merge-MachinePath is the machine's raw PATH value with Folder added once at its end, or with every entry naming it
# taken out (Remove): the other entries kept as written (%SystemRoot% and the like unexpanded), empty entries dropped
# only when the value changes, and the value returned unchanged when the folder is already there once (or, removing,
# not there). Entries compare without case and without a trailing backslash.
function Merge-MachinePath {
    param([string]$Current, [string]$Folder, [bool]$Remove)
    $want = $Folder.Trim().TrimEnd('\')
    $kept = New-Object System.Collections.Generic.List[string]
    $found = $false
    $changed = $false
    foreach ($e in ($Current -split ';')) {
        if ($e.Trim() -eq '') { continue }
        if ($e.Trim().TrimEnd('\') -ieq $want) {
            if ($Remove -or $found) { $changed = $true; continue }
            $found = $true
        }
        $kept.Add($e)
    }
    if (-not $Remove -and -not $found) { $kept.Add($Folder); $changed = $true }
    if (-not $changed) { return $Current }
    return ($kept.ToArray() -join ';')
}

# New-ChildText is the elevated child's code: fixed text with four values put in as literals, each checked first (the
# paths absolute with a drive letter, the hash 64 lowercase hex) and single-quoted (its quotes doubled). Real says
# whether the target is the real place (the PATH and the rights steps), Remove whether the child undoes the install.
function New-ChildText {
    param([string]$Source, [string]$Target, [string]$Hash, [bool]$Real, [bool]$Remove)
    foreach ($p in @($Source, $Target)) {
        if ($p -notmatch '^[A-Za-z]:\\[^\r\n]+$') { throw "the path '$p' is not an absolute path with a drive letter" }
    }
    if ($Hash -cnotmatch '^[0-9a-f]{64}$') { throw "the expected SHA-256 '$Hash' is not 64 lowercase hex characters" }
    $q = { param([string]$s) "'" + [Management.Automation.Language.CodeGeneration]::EscapeSingleQuotedStringContent($s) + "'" }
    $head = '$src = ' + (& $q $Source) + "`n" + '$dst = ' + (& $q $Target) + "`n" + '$want = ' + (& $q $Hash) + "`n" +
        '$real = $' + $Real.ToString().ToLowerInvariant() + "`n" + '$remove = $' + $Remove.ToString().ToLowerInvariant() + "`n"
    $merge = 'function Merge-MachinePath {' + ${function:Merge-MachinePath}.ToString() + "}`n"
    $body = @'
$ErrorActionPreference = 'Stop'
$step = 10
function Move-Busy([string]$a, [string]$b) {
    for ($i = 0; ; $i++) {
        try { [IO.File]::Move($a, $b); return } catch { if ($i -ge 50) { throw }; Start-Sleep -Milliseconds 100 }
    }
}
function Remove-Busy([string]$a) {
    for ($i = 0; ; $i++) {
        try { [IO.File]::Delete($a); return } catch { if ($i -ge 50) { throw }; Start-Sleep -Milliseconds 100 }
    }
}
function Set-MachinePath([bool]$out) {
    $key = [Microsoft.Win32.Registry]::LocalMachine.OpenSubKey('SYSTEM\CurrentControlSet\Control\Session Manager\Environment', $true)
    try {
        $raw = [string]$key.GetValue('Path', '', [Microsoft.Win32.RegistryValueOptions]::DoNotExpandEnvironmentNames)
        $new = Merge-MachinePath -Current $raw -Folder $dir -Remove $out
        if ($new -cne $raw) { $key.SetValue('Path', $new, [Microsoft.Win32.RegistryValueKind]::ExpandString) }
    } finally { $key.Close() }
}
# Test-Rights says whether Users, Authenticated Users or Everyone holds any right that writes on the folder or file:
# write, append, its extended attributes or attributes, delete (its own or a child's), its rights or its owner.
function Test-Rights([string]$p) {
    $bad = @('S-1-5-32-545', 'S-1-5-11', 'S-1-1-0')
    $mask = 0x2 -bor 0x4 -bor 0x10 -bor 0x40 -bor 0x100 -bor 0x10000 -bor 0x40000 -bor 0x80000 -bor 0x10000000 -bor 0x40000000
    if ([IO.Directory]::Exists($p)) { $acl = [IO.Directory]::GetAccessControl($p) } else { $acl = [IO.File]::GetAccessControl($p) }
    foreach ($a in $acl.GetAccessRules($true, $true, [Security.Principal.SecurityIdentifier])) {
        if ($a.AccessControlType -ne 'Allow') { continue }
        if (($a.PropagationFlags -band [Security.AccessControl.PropagationFlags]::InheritOnly) -ne 0) { continue }
        if ($bad -contains $a.IdentityReference.Value -and (([long][int]$a.FileSystemRights) -band $mask) -ne 0) { return $true }
    }
    return $false
}
# Send-SettingChange announces the environment's change (WM_SETTINGCHANGE "Environment" to every window) through .NET's
# own call, compiling nothing: SetEnvironmentVariable at the machine's level sends that message itself after it writes
# the key, and here it deletes a value never made, so the key is left as it was (the PATH is written by
# Set-MachinePath, raw, never through this call, which would write it expanded).
function Send-SettingChange {
    [Environment]::SetEnvironmentVariable('BONSAI_INSTALL_NOTICE', [NullString]::Value, [EnvironmentVariableTarget]::Machine)
}
try {
    $dir = [IO.Path]::GetDirectoryName($dst)
    $name = [IO.Path]::GetFileName($dst)
    if ($remove) {
        $step = 20
        if ([IO.Directory]::Exists($dir)) {
            [Environment]::CurrentDirectory = $dir
            Set-Location -LiteralPath $dir
            foreach ($f in [IO.Directory]::GetFiles($dir)) {
                $n = [IO.Path]::GetFileName($f)
                if ($n -ieq $name -or $n -ieq ($name + '.new') -or $n.StartsWith($name + '.old-', [StringComparison]::OrdinalIgnoreCase)) { Remove-Busy $f }
            }
            Set-Location -LiteralPath ([Environment]::SystemDirectory)
            [Environment]::CurrentDirectory = [Environment]::SystemDirectory
            if ($real -and [IO.Directory]::GetFileSystemEntries($dir).Length -eq 0) { [IO.Directory]::Delete($dir) }
        }
        if ($real) { $step = 21; Set-MachinePath $true; $step = 17; Send-SettingChange }
        exit 0
    }
    [void][IO.Directory]::CreateDirectory($dir)
    [Environment]::CurrentDirectory = $dir
    Set-Location -LiteralPath $dir
    # At the real place the rights are checked before anything lands: the folder before the copy, the copy (with what
    # it inherits) before the rename, so a folder others can write gets no file and no PATH entry.
    if ($real) { $step = 18; if (Test-Rights $dir) { exit 18 } }
    $step = 11
    $new = $dst + '.new'
    [IO.File]::Copy($src, $new, $true)
    $step = 12
    $s = [IO.File]::OpenRead($new)
    try { $h = [Security.Cryptography.SHA256]::Create().ComputeHash($s) } finally { $s.Dispose() }
    if (([BitConverter]::ToString($h) -replace '-', '').ToLowerInvariant() -cne $want) { Remove-Busy $new; exit 12 }
    if ($real) { $step = 18; if (Test-Rights $new) { Remove-Busy $new; exit 18 } }
    foreach ($f in [IO.Directory]::GetFiles($dir)) {
        if ([IO.Path]::GetFileName($f).StartsWith($name + '.old-', [StringComparison]::OrdinalIgnoreCase)) {
            try { [IO.File]::Delete($f) } catch { }  # one still running stays until the next install
        }
    }
    $step = 14
    $aside = ''
    if ([IO.File]::Exists($dst)) {
        $n = 1
        while ([IO.File]::Exists($dst + '.old-' + $n)) { $n++ }
        $aside = $dst + '.old-' + $n
        Move-Busy $dst $aside
    }
    $step = 15
    Move-Busy $new $dst
    if ($aside -ne '') { try { [IO.File]::Delete($aside) } catch { } }  # running: it stays until the next install
    if ($real) {
        $step = 16
        Set-MachinePath $false
        $step = 17
        Send-SettingChange
    }
    exit 0
} catch {
    try { [Console]::Error.WriteLine($_.Exception.Message) } catch { }
    exit $step
}
'@
    return ($head + $merge + $body)
}

# Invoke-Child runs the child's code in Windows PowerShell, by its full path, elevated through the Windows prompt
# (RunAs) or as a plain child of this one, and returns its exit code. A prompt declined is a refusal.
function Invoke-Child {
    param([string]$Text, [bool]$RunAs)
    $ps = Join-Path ([Environment]::SystemDirectory) 'WindowsPowerShell\v1.0\powershell.exe'
    $enc = [Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($Text))
    $argv = @('-NoProfile', '-NonInteractive', '-EncodedCommand', $enc)
    if ($RunAs) {
        try {
            $p = Start-Process -FilePath $ps -ArgumentList $argv -Verb RunAs -Wait -PassThru -WindowStyle Hidden
        } catch {
            Stop-Install ('the Windows prompt was declined, so nothing was installed (' + $_.Exception.Message + ')') 'run the last line again and choose Yes'
        }
        return [int]$p.ExitCode
    }
    $eap = $ErrorActionPreference
    $ErrorActionPreference = 'Continue'
    try { $out = & $ps @argv 2>&1 } finally { $ErrorActionPreference = $eap }
    $code = $LASTEXITCODE
    if ($code -ne 0 -and $out) { foreach ($l in $out) { [Console]::Error.WriteLine([string]$l) } }
    return [int]$code
}

function Get-StepText {
    param([int]$Code)
    if ($Steps.ContainsKey($Code)) { return $Steps[$Code] }
    return "an unknown step (exit $Code)"
}

# Main. An error no step expected is a refusal too.
trap { Stop-Install ('it stopped on an error: ' + $_.Exception.Message) 'run it again; if it stops again, run it with -Remove and then again' }
if ($Target -notmatch '^[A-Za-z]:\\') {
    Stop-Install "the target $Target is not an absolute path with a drive letter" 'run it again with one, as: -Target C:\scratch\bonsai.exe'
}
try { $Target = [IO.Path]::GetFullPath($Target) } catch { Stop-Install "the target $Target is not a path Windows reads" 'run it again with another -Target' }
if ([IO.Directory]::Exists($Target)) {
    Stop-Install "the target $Target is a folder, not a file's path" "run it again with the file's path in it, as: -Target $Target\bonsai.exe, or another target"
}
$real = $Target -ieq $Place
$admin = ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if ($ForceChild -and -not $admin) {
    Stop-Install '-ForceChild is for tests on a machine that is already elevated, and this PowerShell is not' 'run it again without -ForceChild'
}
$runAs = [bool]$ForceChild -or ($real -and -not $admin)
if ($env:BONSAI_HOME) { $bonsaiHome = $env:BONSAI_HOME } else { $bonsaiHome = Join-Path $env:USERPROFILE '.bonsai' }
$record = Join-Path $bonsaiHome 'install.json'
$zero = '0' * 64

if ($Remove) {
    try { $text = New-ChildText -Source $Target -Target $Target -Hash $zero -Real $real -Remove $true } catch { Stop-Install $_.Exception.Message 'run it again with another -Target' }
    $code = Invoke-Child $text $runAs
    if ($code -ne 0) {
        Stop-Install ('the remove stopped at ' + (Get-StepText $code)) 'close every program that may run bonsai.exe (Claude Code sessions), then run it again'
    }
    if ([IO.File]::Exists($record)) { [IO.File]::Delete($record) }
    [Console]::Out.WriteLine("Removed $Target, the copies set aside and $record" + $(if ($real) { ', and its folder''s PATH entry.' } else { '.' }))
    exit 0
}

$src = Join-Path $PSScriptRoot 'bonsai.exe'
if (-not [IO.File]::Exists($src)) {
    Stop-Install "there is no bonsai.exe beside this script ($src)" 'run the install.ps1 that came in the release''s folder, beside its bonsai.exe'
}
$line = ''
try {
    $eap = $ErrorActionPreference
    $ErrorActionPreference = 'Continue'
    $out = @(& $src --version 2>$null)
    $ok = $LASTEXITCODE -eq 0
    $ErrorActionPreference = $eap
    if ($out.Count -gt 0) { $line = [string]$out[0] }
} catch { $ok = $false }
if (-not $ok) {
    Stop-Install "the bonsai.exe beside this script ($src) does not run (bonsai.exe --version failed)" 'download the release for Windows again and run its install.ps1'
}
if (-not $line.StartsWith('bonsai ')) {
    Stop-Install "the file beside this script ($src) does not answer --version as bonsai does" 'download Bonsai''s release again and run its install.ps1'
}
$want = Get-Sha256 $src

try { $text = New-ChildText -Source $src -Target $Target -Hash $want -Real $real -Remove $false } catch { Stop-Install $_.Exception.Message 'run it from a folder with an ordinary path, as: C:\Users\<you>\Downloads\bonsai' }
$code = Invoke-Child $text $runAs
if ($code -ne 0) {
    Stop-Install ('the install stopped at ' + (Get-StepText $code)) 'run it again; if it stops at the same step, run it with -Remove and then again'
}

$sum = Get-Sha256 $Target
if ($sum -cne $want) {
    Stop-Install "the installed $Target is not the file that was checked (its SHA-256 differs)" 'run it again'
}
$eap = $ErrorActionPreference
$ErrorActionPreference = 'Continue'
$installed = @(& $Target --version 2>$null)
$ErrorActionPreference = $eap
$vline = ''
if ($installed.Count -gt 0) { $vline = [string]$installed[0] }
$version = ''
if ($vline.StartsWith('bonsai ')) { $version = ($vline.Substring(7) -split ' ')[0] }

[void][IO.Directory]::CreateDirectory($bonsaiHome)
$json = '{"path": ' + (ConvertTo-Json ($Target -replace '\\', '/')) + ', "version": ' + (ConvertTo-Json $version) + ', "sha256": "' + $sum + '"}' + "`n"
$part = $record + '.new'
[IO.File]::WriteAllText($part, $json, (New-Object System.Text.UTF8Encoding $false))
if ([IO.File]::Exists($record)) { [IO.File]::Replace($part, $record, [NullString]::Value) } else { [IO.File]::Move($part, $record) }

if ($real) {
    $env:Path = [Environment]::GetEnvironmentVariable('Path', 'Machine') + ';' + [Environment]::GetEnvironmentVariable('Path', 'User')
}
[Console]::Out.WriteLine("Installed $Target, bonsai $version.")
[Console]::Out.WriteLine("SHA-256: $sum")
[Console]::Out.WriteLine("The record: $record")
[Console]::Out.WriteLine('Every bonsai on the PATH, in order (Get-Command bonsai -All):')
$found = @(Get-Command bonsai -All -CommandType Application -ErrorAction SilentlyContinue)
if ($found.Count -eq 0) {
    [Console]::Out.WriteLine('  none: its folder is not on the PATH, so hooks and commands that call bonsai by name will not find it')
} else {
    foreach ($c in $found) { [Console]::Out.WriteLine('  ' + $c.Source) }
    if ($found[0].Source -ine $Target) {
        [Console]::Out.WriteLine('  ' + $found[0].Source + " comes first on the PATH, so a command that calls bonsai by name runs it, not $Target`: take it off the PATH")
    }
}
[Console]::Out.WriteLine($vline)
exit 0
