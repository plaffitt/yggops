import subprocess
import sys

def run(command):
  print('>', ' '.join(command), flush=True)
  process = subprocess.Popen(command)
  process.communicate()
  if process.returncode != 0:
    sys.exit(process.returncode)
