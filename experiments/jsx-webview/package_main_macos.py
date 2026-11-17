#!/usr/bin/env python3
"""Compatibility entrypoint: main packages now always use embedded moejs.

Accepts the original --out/--upstream-sha/--upstream-release-json arguments.
No fallback to the legacy Node runtime packaging is permitted.
"""
from pathlib import Path
import runpy

if __name__ == "__main__":
    runpy.run_path(str(Path(__file__).resolve().parents[1] / "moejs" / "package_main_macos.py"), run_name="__main__")
