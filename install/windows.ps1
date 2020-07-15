function InstallXipCli {
    New-Item -Path 'C:\Program Files\XIP' -ItemType Directory
    New-Item -Path 'C:\Program Files\XIP\bin' -ItemType Directory

    Invoke-WebRequest https://github.com/xip-online-applications/xip-cli/releases/latest/download/x-ip_windows_amd64.exe -Outfile 'C:\Program Files\XIP\bin\x-ip.exe'

    $env:Path += 'C:\Program Files\XIP\bin'
    [Environment]::SetEnvironmentVariable
    ("Path", $env:Path, [System.EnvironmentVariableTarget]::Machine)
} InstallXipCli
