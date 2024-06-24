try {
    go run ..\..\..\cmd\... powershell .\custom-fake-server.ps1 ARG1 ARG2 |
            Tee-Object -FilePath output.log
} finally {
    # there is windows limitation: Write-Output (write to pipe)
    # does not write anything after interruption.
    # we can write directly to console using Write-Host,
    # but Tee-Object does not receive this record too.
    # the only way to get record in output log file is
    # to write it manually:
    # $msg = 'some graceful shutdown logs..."'
    # $msg | Out-File -FilePath "output.log" -Append
}