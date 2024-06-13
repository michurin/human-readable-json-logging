powershell ..\fake-server.ps1 |
        go run ..\..\cmd\... |
        Tee-Object -FilePath output.log