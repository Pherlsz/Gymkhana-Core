param(
    [ValidateSet('quick','full','security')]
    [string]$Mode = 'quick'
)

$bash = Get-Command bash -ErrorAction SilentlyContinue
if (-not $bash) {
    Write-Error 'Install Git Bash or WSL before running local verification.'
    exit 1
}

Push-Location (Resolve-Path (Join-Path $PSScriptRoot '..'))
try {
    & $bash.Source 'scripts/verify-local.sh' $Mode
    exit $LASTEXITCODE
}
finally {
    Pop-Location
}
