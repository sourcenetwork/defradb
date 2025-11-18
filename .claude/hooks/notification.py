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
import subprocess
import random
from pathlib import Path

try:
    from dotenv import load_dotenv
    load_dotenv()
except ImportError:
    pass  # dotenv is optional

# Import common utilities
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from common import (
    append_to_json_log,
    announce_via_tts,
    read_json_input
)


def announce_notification():
    """Announce that the agent needs user input."""
    # Get engineer name if available
    engineer_name = os.getenv('ENGINEER_NAME', '').strip()
    
    # Create notification message with 30% chance to include name
    if engineer_name and random.random() < 0.3:
        notification_message = f"{engineer_name}, your agent needs your input"
    else:
        notification_message = "Your agent needs your input"
    
    # Announce via TTS
    announce_via_tts(notification_message)


def main():
    try:
        # Parse command line arguments
        parser = argparse.ArgumentParser()
        parser.add_argument('--notify', action='store_true', help='Enable TTS notifications')
        args = parser.parse_args()
        
        # Read JSON input from stdin
        input_data = read_json_input()
        
        # Log the notification event
        append_to_json_log('notification.json', input_data)
        
        # Announce notification via TTS only if --notify flag is set
        # Skip TTS for the generic "Claude is waiting for your input" message
        if args.notify and input_data.get('message') != 'Claude is waiting for your input':
            announce_notification()
        
        sys.exit(0)
        
    except json.JSONDecodeError:
        # Handle JSON decode errors gracefully
        sys.exit(0)
    except Exception:
        # Handle any other errors gracefully
        sys.exit(0)

if __name__ == '__main__':
    main()