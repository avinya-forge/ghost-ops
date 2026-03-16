import re

with open("docs/planning/backlog.md", "r") as f:
    lines = f.readlines()

epic_lines = []
new_lines = []
in_epic = False

for line in lines:
    if line.startswith("- **EPIC 500: Observer Agent & Metric Thresholds**"):
        in_epic = True
        epic_lines.append(line)
    elif in_epic and line.startswith("- **TASK 600: Rust Guest Sdk Design**"):
        in_epic = False
        new_lines.append(line)
    elif in_epic:
        epic_lines.append(line)
    else:
        new_lines.append(line)

with open("docs/planning/backlog.md", "w") as f:
    f.writelines(new_lines)

with open("docs/release/release-notes.md", "r") as f:
    release_content = f.read()

# insert right before "## v0.2.1 (Released)" or at the beginning of the file if not found, let's just append to top section
new_release_notes = "- **Observer Agent & Metric Thresholds**: EPIC 500 Complete. Implemented ObserverAgent to monitor P99 latency and error rates, and emit events for re-prompting.\n"

with open("docs/release/release-notes.md", "w") as f:
    f.write(new_release_notes + release_content)

print("Epic moved to release notes.")
