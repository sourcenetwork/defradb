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
from common import (
    get_ai_working_dir,
    get_current_branch,
    append_to_json_log,
    read_file_contents,
    read_json_input,
    DOCS_DIR
)

def main():
    try:
        # Read JSON input from stdin
        input_data = read_json_input()
        
        # Log the hook event
        append_to_json_log('pre_compact.json', input_data)
        
        # Get AI working directory for current branch
        ai_working_dir = get_ai_working_dir()
        if not ai_working_dir:
            # Cannot determine branch/directory, exit gracefully
            sys.exit(0)
        
        # Get current branch for display
        branch = get_current_branch()
        
        # Always include org general practices
        org_practices_path = os.path.join(DOCS_DIR, 'org-general-practices.md')
        
        # Build context messages
        context_messages = []
        
        # Add org general practices
        content = read_file_contents(org_practices_path)
        if content:
            context_messages.append(f"# Organization General Practices\n\n{content}")
        
        # Add all markdown files from AI working directory
        if os.path.exists(ai_working_dir):
            # Get all .md files in the directory
            md_files = []
            for filename in os.listdir(ai_working_dir):
                if filename.endswith('.md'):
                    md_files.append(filename)
            
            # Sort files for consistent ordering
            md_files.sort()
            
            # Read each file
            for filename in md_files:
                file_path = os.path.join(ai_working_dir, filename)
                content = read_file_contents(file_path)
                if content:
                    context_messages.append(f"# AI Working Directory: {filename}\n\n{content}")
        
        # If we have context to add, print it to stderr
        # This will be shown to the user as part of the compaction process
        if context_messages:
            full_context = "\n\n---\n\n".join(context_messages)
            print(f"Automatically including AI context from branch '{branch}':\n", file=sys.stderr)
            print(f"- Organization general practices", file=sys.stderr)
            if os.path.exists(ai_working_dir):
                print(f"- AI working directory files from: ai/context/{branch}/", file=sys.stderr)
            print("\nTo ensure this context is preserved in the compacted conversation.", file=sys.stderr)
        
        # Exit successfully
        sys.exit(0)
        
    except json.JSONDecodeError:
        # Handle JSON decode errors gracefully
        sys.exit(0)
    except Exception as e:
        # Log error but exit gracefully
        print(f"Error in pre_compact hook: {str(e)}", file=sys.stderr)
        sys.exit(0)

if __name__ == '__main__':
    main()