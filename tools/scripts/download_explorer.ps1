# Download static assets from: `github.com/sourcenetwork/defradb-explorer`.
#
# Bump the release tag in the URL below to change versions.

$url = "https://github.com/sourcenetwork/defradb-explorer/releases/download/v1.0.0/dist.tar.gz"
$tarFile = "dist.tar.gz"

try {
    # Download the file
    Write-Host "Downloading explorer assets..."
    Invoke-WebRequest -Uri $url -OutFile $tarFile -ErrorAction Stop

    # Extract the tar.gz file
    Write-Host "Extracting assets..."
    tar -xzf $tarFile

    # Clean up the downloaded archive
    Remove-Item $tarFile

    Write-Host "Download complete!"
} catch {
    Write-Error "Failed to download or extract explorer assets: $_"
    exit 1
}
