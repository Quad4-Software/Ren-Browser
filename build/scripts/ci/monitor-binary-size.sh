#!/bin/bash
set -euo pipefail

# Reference sizes in MB (from user query)
declare -A REF_SIZES
REF_SIZES["renbrowser-android-universal-signed.apk"]=11.7
REF_SIZES["renbrowser-android-universal-debug.apk"]=11.7 # Fallback for debug
REF_SIZES["renbrowser-ios.ipa"]=9.86
REF_SIZES["renbrowser-linux-amd64"]=26.9
REF_SIZES["renbrowser-linux-amd64.AppImage"]=85.2
REF_SIZES["renbrowser-linux-amd64.deb"]=10.4
REF_SIZES["renbrowser-linux-amd64.flatpak"]=7.92
REF_SIZES["renbrowser-linux-amd64.pkg.tar.zst"]=10.5
REF_SIZES["renbrowser-linux-amd64.rpm"]=10.8
REF_SIZES["renbrowser-linux-arm64"]=25.8
REF_SIZES["renbrowser-linux-arm64.AppImage"]=82.7
REF_SIZES["renbrowser-linux-arm64.deb"]=9.86
REF_SIZES["renbrowser-linux-arm64.pkg.tar.zst"]=9.84
REF_SIZES["renbrowser-linux-arm64.rpm"]=10.2
REF_SIZES["renbrowser-macos-universal.zip"]=21.1
REF_SIZES["renbrowser-micron-translator-backend.wasm"]=1.13
REF_SIZES["renbrowser-micron-translator.wasm"]=1.14
REF_SIZES["renbrowser-server-freebsd-amd64"]=26.6
REF_SIZES["renbrowser-server-freebsd-arm64"]=25.6
REF_SIZES["renbrowser-server-linux-amd64"]=26.8
REF_SIZES["renbrowser-server-linux-arm64"]=25.7
REF_SIZES["renbrowser-server-linux-armv6"]=24.8
REF_SIZES["renbrowser-server-netbsd-amd64"]=26.6
REF_SIZES["renbrowser-server-openbsd-amd64"]=26.6
REF_SIZES["renbrowser-server-windows-amd64.exe"]=27.6
REF_SIZES["renbrowser-server-windows7-amd64.exe"]=27.6
REF_SIZES["renbrowser-windows-amd64-installer.exe"]=12.0
REF_SIZES["renbrowser-windows-amd64.exe"]=27.2

THRESHOLD=1.5 # 50% increase is "massive" as per "massively rise" request. 
# Let's be a bit more strict, maybe 25%? User said "massively", let's use 30%.
THRESHOLD_RATIO=1.3

FAILED=0
DIRECTORY=${1:-release}

echo "Checking binary sizes in $DIRECTORY..."

for file in "$DIRECTORY"/*; do
    filename=$(basename "$file")
    
    # Skip checksums and SBOMs
    if [[ "$filename" == "SHA256SUMS.txt" || "$filename" == *"sbom"* ]]; then
        continue
    fi

    # Find matching reference
    ref_size=""
    for ref_name in "${!REF_SIZES[@]}"; do
        if [[ "$filename" == "$ref_name" ]]; then
            ref_size=${REF_SIZES[$ref_name]}
            break
        fi
    done

    if [ -n "$ref_size" ]; then
        size_bytes=$(stat -c%s "$file")
        size_mb=$(echo "scale=2; $size_bytes / 1048576" | bc)
        
        limit=$(echo "scale=2; $ref_size * $THRESHOLD_RATIO" | bc)
        
        echo -n "  $filename: $size_mb MB (ref: $ref_size MB, limit: $limit MB) "
        
        if (( $(echo "$size_mb > $limit" | bc -l) )); then
            echo "FAILED"
            FAILED=1
        else
            echo "OK"
        fi
    else
        echo "  $filename: No reference size found, skipping check."
    fi
done

if [ $FAILED -ne 0 ]; then
    echo "Error: One or more binaries have increased in size significantly!"
    exit 1
fi

echo "All binary sizes are within limits."
