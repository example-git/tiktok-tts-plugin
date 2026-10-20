#!/usr/bin/env python3
import argparse
import json
import os
from pathlib import Path
import shutil
import subprocess


def main():
    parser = argparse.ArgumentParser(description="Build an installable Haunt server extension.")
    parser.add_argument("--output", type=Path, default=Path("dist"))
    args = parser.parse_args()
    root = Path(__file__).resolve().parent
    manifest = json.loads((root / "manifest.json").read_text())
    package = args.output.resolve() / manifest["id"]
    package.mkdir(parents=True, exist_ok=True)
    subprocess.run(
        ["go", "build", "-trimpath", "-o", str(package / manifest["entrypoint"]), "./cmd/haunt-tiktok-tts"],
        cwd=root,
        env={**os.environ, "CGO_ENABLED": "0"},
        check=True,
    )
    shutil.copyfile(root / "manifest.json", package / "manifest.json")
    print(package)


if __name__ == "__main__":
    main()
