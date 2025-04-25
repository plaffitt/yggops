import subprocess
import sys


def run(command):
    print(">", " ".join(command), flush=True)
    with subprocess.Popen(command) as process:
        process.communicate()
        if process.returncode != 0:
            sys.exit(process.returncode)
