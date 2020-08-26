param (
    [switch]$clean = $false
)

$version = "1.4.0"

$filesToClean = '.\mcdolphin', '.\mcdolphin.exe', '.\vendor'
$ldflags = "-X main.Version=$version"

function Invoke-Build {
    Write-Host "Building '.\cmd\mcdolphin'..."
    Write-Host "Using LDFLAGS: `"$ldflags`""

    & go build -ldflags `"$ldflags`" -o mcdolphin.exe .\cmd\mcdolphin

    Write-Host 'Building complete!'
}

function Invoke-Clean {
    Write-Host 'Running go mod tidy...'
    & 'go' 'mod' 'tidy'

    # Remove unwanted files and directories
    foreach ($file in $filesToClean) {
        # Check if the file exists
        if (Test-Path -Path $file) {
            # Check if the file is actually a file and not a directory
            if (Test-Path -Path $file -PathType Leaf) {
                Write-Host "Removing '$file'"
                Remove-Item -Path $file
            }
            else {
                Write-Host "Removing '$file'"
                Remove-Item -Path $file -Recurse
            }
        }
    }
}

if ($clean) {
    Invoke-Clean
    exit
}

Invoke-Build
