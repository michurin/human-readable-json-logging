Add-Type @"
using System;
using System.Runtime.InteropServices;

public class NativeMethods {
    [DllImport("kernel32.dll", SetLastError = true)]
    public static extern bool GenerateConsoleCtrlEvent(uint dwCtrlEvent, uint dwProcessGroupId);

    [DllImport("kernel32.dll", SetLastError = true)]
    public static extern bool FreeConsole();

    public const uint CTRL_C_EVENT = 0;
    public const uint CTRL_BREAK_EVENT = 1;
}
"@

function Send-CtrlBreakEvent {
    param (
        [int]$parentPid
    )

    [NativeMethods]::GenerateConsoleCtrlEvent([NativeMethods]::CTRL_C_EVENT, 0)
    [NativeMethods]::FreeConsole()
}

Write-Output 'Sleeping...'
Start-Sleep -Seconds 1
Write-Output 'Going to kill parent process...'

$parentPid = (Get-WmiObject Win32_Process -Filter "ProcessId=$PID").ParentProcessId
$null = Send-CtrlBreakEvent -parentPid $parentPid
