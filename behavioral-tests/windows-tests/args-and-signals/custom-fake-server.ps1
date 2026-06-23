Write-Output "Arguments: $args"

$sendCtrlBreakScriptPath = Join-Path -Path $PSScriptRoot -ChildPath "custom-send-break.ps1"
$proc = Start-Process -FilePath "powershell.exe" -ArgumentList "-File `"$sendCtrlBreakScriptPath`"" -PassThru -NoNewWindow

try {
    while ($true) {
        Start-Sleep -Seconds 5
    }
} finally {
    # windows does not allow to write into pipe after interruption (Write-Output)
    # we can only write to console directly
    Write-Host "Getting SIGNAL SIGINT. Exiting"
}

