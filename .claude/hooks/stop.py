#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.11"
# dependencies = [
#     "python-dotenv",
# ]
# ///

import argparse
import json
import os
import sys
import random
import subprocess
from pathlib import Path
from datetime import datetime

try:
    from dotenv import load_dotenv
    load_dotenv()
except ImportError:
    pass  # dotenv is optional

# Import common utilities
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from common import (
    append_to_json_log,
    get_llm_completion_message,
    announce_via_tts,
    process_chat_transcript,
    save_chat_transcript,
    hook_main
)


def get_completion_messages():
    """Return list of friendly completion messages."""
    return [
        "Work complete!",
        "All done!",
        "Task finished!",
        "Job complete!",
        "Ready for next task!"
    ]


# Remove duplicate function - now using get_llm_completion_message from common.py

def announce_completion():
    """Announce completion using the best available TTS service."""
    # Get completion message (LLM-generated or fallback)
    completion_message = get_llm_completion_message(get_completion_messages())
    
    # Announce via TTS
    announce_via_tts(completion_message)


def main():
    try:
        # Parse command line arguments
        parser = argparse.ArgumentParser()
        parser.add_argument('--chat', action='store_true', help='Copy transcript to chat.json')
        args = parser.parse_args()
        
        # Read JSON input from stdin
        input_data = json.load(sys.stdin)

        # Extract required fields
        session_id = input_data.get("session_id", "")
        stop_hook_active = input_data.get("stop_hook_active", False)

        # Log the stop event
        append_to_json_log('stop.json', input_data)
        
        # Handle --chat switch
        if args.chat and 'transcript_path' in input_data:
            transcript_path = input_data['transcript_path']
            chat_data = process_chat_transcript(transcript_path)
            if chat_data:
                save_chat_transcript(chat_data)

        # Announce completion via TTS
        announce_completion()

        sys.exit(0)

    except json.JSONDecodeError:
        # Handle JSON decode errors gracefully
        sys.exit(0)
    except Exception:
        # Handle any other errors gracefully
        sys.exit(0)


if __name__ == "__main__":
    main()
