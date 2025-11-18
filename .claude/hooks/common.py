#!/usr/bin/env python3
"""
Common utilities and constants for Claude hooks.
This module provides shared functionality used across different hook scripts.
"""

import json
import os
import sys
import random
import subprocess
from pathlib import Path
from typing import Any, Dict, List, Optional


# Constants
AI_BASE_DIR = os.path.join(os.getcwd(), 'ai')
LOGS_DIR = os.path.join(AI_BASE_DIR, 'logs')
CONTEXT_DIR = os.path.join(AI_BASE_DIR, 'context')
DOCS_DIR = os.path.join(AI_BASE_DIR, 'docs')


def ensure_log_dir() -> str:
    """Ensure the log directory exists and return its path."""
    os.makedirs(LOGS_DIR, exist_ok=True)
    return LOGS_DIR


def get_current_branch() -> Optional[str]:
    """Get the current git branch name."""
    try:
        result = subprocess.run(
            ['git', 'branch', '--show-current'],
            capture_output=True,
            text=True,
            check=True
        )
        return result.stdout.strip()
    except (subprocess.CalledProcessError, FileNotFoundError):
        return None


def get_ai_working_dir() -> Optional[str]:
    """Get the AI working directory path for the current git branch."""
    branch = get_current_branch()
    if not branch:
        return None
    
    # Split branch name by '/' and build path properly for cross-platform compatibility
    branch_parts = branch.split('/')
    return os.path.join(CONTEXT_DIR, *branch_parts)


def get_log_dir(ensure_exists: bool = False) -> str:
    """
    Get the appropriate log directory based on the current context.
    
    Args:
        ensure_exists: If True, create the directory if it doesn't exist.
    
    Returns:
        Path to the log directory (either AI working dir logs or default logs).
    """
    ai_working_dir = get_ai_working_dir()
    if ai_working_dir:
        log_dir = os.path.join(ai_working_dir, 'logs')
        if ensure_exists:
            os.makedirs(log_dir, exist_ok=True)
    else:
        # Fallback to default logs directory if no git branch
        if ensure_exists:
            ensure_log_dir()
        log_dir = LOGS_DIR
    
    return log_dir


def read_json_log(log_filename: str) -> List[Dict[str, Any]]:
    """Read and return JSON log data, handling errors gracefully."""
    log_dir = get_log_dir(ensure_exists=False)
    log_path = os.path.join(log_dir, log_filename)
    
    if os.path.exists(log_path):
        try:
            with open(log_path, 'r') as f:
                return json.load(f)
        except (json.JSONDecodeError, ValueError):
            return []
    return []


def write_json_log(log_filename: str, data: List[Dict[str, Any]]) -> None:
    """Write JSON data to log file with proper formatting."""
    log_dir = get_log_dir(ensure_exists=True)
    log_path = os.path.join(log_dir, log_filename)
    
    with open(log_path, 'w') as f:
        json.dump(data, f, indent=2)


def append_to_json_log(log_filename: str, new_data: Dict[str, Any]) -> None:
    """Append new data to existing JSON log file."""
    log_data = read_json_log(log_filename)
    log_data.append(new_data)
    write_json_log(log_filename, log_data)


def read_file_contents(file_path: str) -> Optional[str]:
    """Read and return file contents if file exists."""
    try:
        if os.path.exists(file_path):
            with open(file_path, 'r', encoding='utf-8') as f:
                return f.read()
    except Exception:
        pass
    return None


def get_tts_script_path() -> Optional[str]:
    """
    Determine which TTS script to use based on available API keys.
    Priority order: ElevenLabs > OpenAI > pyttsx3
    """
    # Get hooks directory and construct utils/tts path
    hooks_dir = Path(__file__).parent
    tts_dir = hooks_dir / "utils" / "tts"
    
    # Check for ElevenLabs API key (highest priority)
    if os.getenv('ELEVENLABS_API_KEY'):
        elevenlabs_script = tts_dir / "elevenlabs_tts.py"
        if elevenlabs_script.exists():
            return str(elevenlabs_script)
    
    # Check for OpenAI API key (second priority)
    if os.getenv('OPENAI_API_KEY'):
        openai_script = tts_dir / "openai_tts.py"
        if openai_script.exists():
            return str(openai_script)
    
    # Fall back to pyttsx3 (no API key required)
    pyttsx3_script = tts_dir / "pyttsx3_tts.py"
    if pyttsx3_script.exists():
        return str(pyttsx3_script)
    
    return None


def process_chat_transcript(transcript_path: str) -> List[Dict[str, Any]]:
    """Process a .jsonl transcript file and return as JSON array."""
    chat_data = []
    if os.path.exists(transcript_path):
        try:
            with open(transcript_path, 'r') as f:
                for line in f:
                    line = line.strip()
                    if line:
                        try:
                            chat_data.append(json.loads(line))
                        except json.JSONDecodeError:
                            pass  # Skip invalid lines
        except Exception:
            pass  # Fail silently
    return chat_data


def save_chat_transcript(chat_data: List[Dict[str, Any]]) -> None:
    """Save chat transcript data to logs/chat.json."""
    log_dir = get_log_dir(ensure_exists=True)
    chat_file = os.path.join(log_dir, 'chat.json')
    with open(chat_file, 'w') as f:
        json.dump(chat_data, f, indent=2)


def read_json_input() -> Dict[str, Any]:
    """Read and parse JSON input from stdin."""
    try:
        return json.load(sys.stdin)
    except json.JSONDecodeError:
        # Return empty dict on decode error
        return {}


def run_script_with_timeout(script_path: str, args: List[str], timeout: int = 10) -> Optional[str]:
    """
    Run an external script with timeout and return output.
    
    Args:
        script_path: Path to the script to run
        args: List of arguments to pass to the script
        timeout: Timeout in seconds (default: 10)
    
    Returns:
        stdout if successful, None otherwise
    """
    try:
        result = subprocess.run(
            ["uv", "run", script_path] + args,
            capture_output=True,
            text=True,
            timeout=timeout
        )
        if result.returncode == 0 and result.stdout.strip():
            return result.stdout.strip()
    except (subprocess.TimeoutExpired, subprocess.SubprocessError, FileNotFoundError):
        pass
    return None


def get_llm_script_path(service: str) -> Optional[str]:
    """Get the path to an LLM service script."""
    hooks_dir = Path(__file__).parent
    llm_dir = hooks_dir / "utils" / "llm"
    
    if service == "openai":
        return str(llm_dir / "oai.py") if (llm_dir / "oai.py").exists() else None
    elif service == "anthropic":
        return str(llm_dir / "anth.py") if (llm_dir / "anth.py").exists() else None
    return None


def get_llm_completion_message(fallback_messages: List[str]) -> str:
    """
    Generate completion message using available LLM services.
    Priority order: OpenAI > Anthropic > fallback to random message
    
    Args:
        fallback_messages: List of fallback messages to choose from
    
    Returns:
        Generated or fallback completion message
    """
    # Try OpenAI first
    if os.getenv('OPENAI_API_KEY'):
        script = get_llm_script_path("openai")
        if script:
            result = run_script_with_timeout(script, ["--completion"])
            if result:
                return result
    
    # Try Anthropic second
    if os.getenv('ANTHROPIC_API_KEY'):
        script = get_llm_script_path("anthropic")
        if script:
            result = run_script_with_timeout(script, ["--completion"])
            if result:
                return result
    
    # Fallback to random predefined message
    return random.choice(fallback_messages)


def announce_via_tts(message: str) -> None:
    """Announce a message using the best available TTS service."""
    try:
        tts_script = get_tts_script_path()
        if not tts_script:
            return  # No TTS scripts available
        
        subprocess.run(
            ["uv", "run", tts_script, message],
            capture_output=True,  # Suppress output
            timeout=10  # 10-second timeout
        )
    except (subprocess.TimeoutExpired, subprocess.SubprocessError, FileNotFoundError):
        # Fail silently if TTS encounters issues
        pass
    except Exception:
        # Fail silently for any other errors
        pass


def hook_main(func):
    """
    Decorator for hook main functions that provides standard error handling.
    The decorated function should accept the parsed JSON input data.
    """
    def wrapper():
        try:
            # Read JSON input
            input_data = read_json_input()
            
            # Call the actual hook function
            return func(input_data)
            
        except json.JSONDecodeError:
            # Handle JSON decode errors gracefully
            sys.exit(0)
        except Exception:
            # Handle any other errors gracefully
            sys.exit(0)
    
    return wrapper