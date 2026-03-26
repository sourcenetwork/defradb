# Download static assets from: `github.com/sourcenetwork/defradb-playground`.
#
# Bump the release tag in the URL below to change versions.

$url = "https://github.com/sourcenetwork/defradb-playground/releases/download/v1.0.1/dist.tar.gz"
$tarFile = "dist.tar.gz"

try {
    # Download the file
    Write-Host "Downloading playground assets..."
    Invoke-WebRequest -Uri $url -OutFile $tarFile -ErrorAction Stop

    # Extract the tar.gz file
    Write-Host "Extracting assets..."
    tar -xzf $tarFile

    # Clean up the downloaded archive
    Remove-Item $tarFile

    Write-Host "Download complete!"
} catch {
    Write-Error "Failed to download or extract playground assets: $_"
    exit 1
}
