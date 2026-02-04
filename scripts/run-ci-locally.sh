#!/bin/bash
set -e

echo "=== Run GitHub Actions CI Locally ==="
echo
echo "This script uses 'act' to run the exact CI workflow locally"
echo

# Check if act is installed
if ! command -v act &> /dev/null; then
    echo "Installing act (GitHub Actions local runner)..."

    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        curl --proto '=https' --tlsv1.2 -sSf https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        brew install act
    else
        echo "Please install act manually: https://github.com/nektos/act#installation"
        exit 1
    fi
fi

echo "✓ act is installed"
echo

# Check Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "Error: Docker is not running. Please start Docker first."
    exit 1
fi

echo "✓ Docker is running"
echo

# Run CI workflow locally
echo "Running CI workflow locally (this may take several minutes)..."
echo

# Use -j to run jobs in parallel, -v for verbose
# The workflow will use local Docker instead of pulling from Docker Hub if images are cached
act --container-architecture linux/amd64 \
    --workflows .github/workflows/ci.yml \
    --verbose \
    "$@"

echo
echo "=== CI Workflow Complete ==="
echo
echo "If all jobs passed above, your CI will be green on GitHub!"
