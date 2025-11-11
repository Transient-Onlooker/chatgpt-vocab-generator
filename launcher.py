import subprocess
import os
import sys

# The absolute path to the directory where this launcher script is located.
launcher_dir = os.path.dirname(os.path.abspath(__file__))

# The path to the "소스코드" directory.
source_dir = os.path.join(launcher_dir, '소스코드')

# The path to the script to run.
script_path = os.path.join(source_dir, 'gui_vocab_maker.py')

# The command to run python. sys.executable is the path to the current python interpreter.
command = [sys.executable, script_path]

# Run the script with its working directory set to the "소스코드" directory.
# This ensures it can find api.json.
subprocess.run(command, cwd=source_dir)
