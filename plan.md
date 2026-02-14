---

# Calisthenics Logger – Implementation Plan

## Goal
Implement a minimal, terminal-based calisthenics workout logger using Go.

The application must:
- Run from terminal
- Prompt user for workout input
- Auto-resolve goals based on predefined exercise progressions
- Append records to a flat log file using | as delimiter
- Be installable as a single binary and callable via shell alias

No GUI, no TUI, no database.

---

## Core Constraints

- Language: Go
- Interface: stdin / stdout only
- Storage: append-only text file
- Delimiter: |
- No Python
- No external dependencies
- No background processes

---

## Data Format (Immutable Contract)

Each log entry MUST follow this exact format:

DATE | DAY | EXERCISE | LEVEL | REPS×SETS | GOAL | COMMENT

Rules:
- One line = one exercise
- Append-only (never modify existing lines)
- COMMENT may be empty
- REPS×SETS format example: 20x2
- GOAL is auto-filled by the program

Example:

2026-01-24|A|Pushups|Half|20x2|25x2|Solid control

---

## Application Behavior

### Execution Model
- Program runs once
- Prompts user for inputs
- Writes exactly one log line
- Exits immediately

No loops, no menus.

---

## User Inputs

Prompt the user for:
1. Day (A / B / C)
2. Exercise (mutiple choices)
3. Level (mutiple choices)
4. Reps * Sets (string)
5. Comment (optional string)

No advanced validation required in v1.

---

## Goal Resolution Logic

Goals MUST be predefined and resolved automatically.

### Data Structure
Use a nested map:

```go
map[string]map[string]string

Where:

Key 1 = Exercise

Key 2 = Level

Value = Goal (e.g. 20x2, 50x3, 2min)


Lookup Rules

If exercise exists AND level exists → use defined goal

Otherwise → goal is "-"



---

Goal Table (Convict Conditioning Style)

Implement the following predefined goals exactly:

Pushups

Wall → 50x3

Incline → 40x3

Kneeling → 30x3

Half → 25x2

Full → 20x2

Close → 20x2

Uneven → 20x2

Half One-Arm → 20x2

Lever → 20x2

One-Arm → 100x1


Squats

Shoulderstand → 50x3

Jackknife → 40x3

Supported → 30x3

Half → 50x2

Full → 30x2

Close → 20x2

Uneven → 20x2

Half One-Leg → 20x2

Assisted One-Leg → 20x2

One-Leg → 50x2


Pullups

Vertical → 40x3

Horizontal → 30x3

Jackknife → 20x3

Half → 15x2

Full → 10x2

Close → 10x2

Uneven → 9x2

Half One-Arm → 8x2

Assisted One-Arm → 7x2

One-Arm → 6x2


Leg Raises

Knee Tuck → 40x3

Knee Raise → 35x3

Bent Leg → 30x3

Frog → 25x3

Flat → 20x2

Hanging Knee → 15x2

Hanging Bent → 15x2

Partial → 15x2

Hanging → 30x2


Bridges

Short → 50x3

Straight → 40x3

Angled → 30x3

Head → 25x2

Half → 20x2

Full → 15x2

Wall Down → 10x2

Wall Up → 8x2

Closing → 6x2

Stand-to-Stand → 10-30x2


Handstand Push-ups

Wall Headstand → 2min

Crow → 1min

Wall → 2min

Half → 20x2

Full → 15x2

Close → 12x2

Uneven → 10x2

Half One-Arm → 8x2

Lever → 6x2

One-Arm → 5x2



---

File Output

Log Location

Directory: ~/workout/

Filename: workout-<YEAR>.log


Example:

~/workout/workout-2026.log

The program must:

Create directory if missing

Create file if missing

Append new lines only



---

Program Flow

1. Load goal map (compile-time)


2. Read user input via stdin


3. Resolve goal using Exercise + Level


4. Format log line using |


5. Append to yearly log file


6. Print confirmation


7. Exit




---

Error Handling Policy

Unknown exercise or level → goal = "-" (do not abort)

File write failure → print error and exit non-zero

Never prevent logging due to invalid input



---

Build & Installation

Build

go build -o cali-log

Install

mkdir -p ~/.local/bin
mv cali-log ~/.local/bin/
chmod +x ~/.local/bin/cali-log

Ensure PATH includes:

export PATH="$HOME/.local/bin:$PATH"


---

Terminal Alias

Add to shell config (.bashrc / .zshrc):
alias cali='cali-log'

User can now run:

cali


---

Explicit Non-Goals (Do NOT Implement)

Editing past logs

Deleting logs

Statistics

Menus

Auto-complete

Background services

Databases

Config files



---

Future Extensions (Out of Scope)

Stats reader

Goal completion checks

Auto-next-level suggestion

Shell auto-completion

TUI interface


These must NOT affect the existing log format.


---

Definition of Done

Single Go binary

Logs correct format

Goals auto-filled correctly

Works via terminal alias

No runtime dependencies
