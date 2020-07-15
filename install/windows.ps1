function InstallXipCli {
    # The install dir
    $installDir = 'C:\Program Files\XIP\bin'

    # Create install dir if it doesn't exist
    New-Item -Path $installDir -Force -ItemType Directory

    # Download the file
    Invoke-WebRequest 'https://github.com/xip-online-applications/xip-cli/releases/latest/download/x-ip_windows_amd64.exe' -Outfile "$installDir\x-ip.exe"

    # Update Path environment variable
    $regexAddPath = [regex]::Escape($installDir)
    $arrPath = $env:Path -split ';' | Where-Object {$_ -notMatch "^$regexAddPath\\?"}
    $env:Path = ($arrPath + $installDir) -join ';'

    # Write new path env variable
    [Environment]::SetEnvironmentVariable
    ("Path", $env:Path, [System.EnvironmentVariableTarget]::Machine)
} InstallXipCli
