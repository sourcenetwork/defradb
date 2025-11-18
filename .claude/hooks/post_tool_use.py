#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.8"
# ///

import json
import os
import sys
from pathlib import Path

# Import common utilities
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from common import append_to_json_log, hook_main

@hook_main
def main(input_data):
    # Log the tool use event
    append_to_json_log('post_tool_use.json', input_data)
    sys.exit(0)

if __name__ == '__main__':
    main()