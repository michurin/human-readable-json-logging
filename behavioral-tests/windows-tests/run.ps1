$invoke_path = $(Get-Location)
$tests_path = Split-Path -Parent $MyInvocation.MyCommand.Path

function doTests
{
    param ($InvokePath)

    Set-Location -Path $InvokePath

    $directories = Get-ChildItem -Directory

    $failedTests = @()  # TODO remove after debug

    foreach ($cs in $directories)
    {
        Set-Location -Path $cs.FullName

        Write-Output "" | Tee-Object -FilePath output.log

        if (Test-Path .\custom-run.ps1 -PathType Leaf)
        {
            $cmd = {
                .\custom-run.ps1
            }
        }
        else
        {
            $cmd = {
                go run ..\..\..\cmd\... powershell ..\fake-server.ps1 | Tee-Object -FilePath output.log
            }
        }

        $P = Start-Process -PassThru -FilePath "powershell.exe" -ArgumentList "-Command", $cmd.ToString()
        $P.WaitForExit()

        Write-Output $cs.name
        Get-Content -Path output.log

        $diff = Compare-Object -ReferenceObject (Get-Content expected.log) -DifferenceObject (Get-Content output.log) -SyncWindow 0

        if ($null -ne $diff)
        {
            Write-Output "${cs}: difference detected in log files"
            Write-Output $diff
            $failedTests += $cs
#            TODO just for debug... failfast later
#            throw "Difference detected in log files."
        }
        else {
            Remove-Item -Path output.log -Force
        }
    }

    # TODO remove after debug
    if ($failedTests.Count -gt 0) {
        $cnt = $failedTests.Count
        Write-Output "Difference detected in log files in ${cnt} tests: $failedTests"
        throw "Difference detected in log files."
    }

}

try {
    doTests $tests_path
    Write-Output "OK"
} catch {
    Write-Output "FAIL"
    exit 1
} finally {
    Set-Location $invoke_path
}