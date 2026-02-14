# Calisthenics Logger (`cali`)

Terminal workout logger written in Go.

It supports two storage backends:
- `sheets` (default): reads/writes to Google Sheets
- `local` (optional override): writes to yearly files in `~/cali-logger/workout`

Each workout entry stores:

`DATE | DAY | EXERCISE | LEVEL | REPSxSETS | GOAL | COMMENT`

Example:

`2026-02-14|A|Pushups|Half|20x2|25x2|Solid form`

## Features

- Log one workout entry interactively
- Show last 10 entries (`-p`)
- Search entries by date (`-s YYYY-MM-DD`)
- Remove one entry from a date (`-r`)
- Open workout template link (`--template`)

## Commands

From terminal:

```bash
cali                    # log a new workout
cali -p                 # print last 10 workouts
cali -s 2026-02-14      # search by date
cali -r                 # remove one entry from a date
cali --help             # show help
cali --template         # open workout template link
```

`--template` opens the Google Drive template link. Local docs files are kept but not opened by CLI commands.

## Build and Install

### Linux / macOS

Build from repo root:

```bash
go build -o ~/.local/bin/cali .
```

If `~/.local/bin` is not in `PATH`, add this line to `~/.bashrc` or `~/.zshrc`:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

### Windows (PowerShell)

Build from repo root:

```powershell
go build -o "$HOME\bin\cali.exe" .
```

Run it:

```powershell
& "$HOME\bin\cali.exe" --help
```

Add the folder to PATH (current user):

```powershell
[Environment]::SetEnvironmentVariable(
  "Path",
  $env:Path + ";$HOME\bin",
  "User"
)
```

Open a new terminal after updating PATH.

## Storage Modes

### 1) Google Sheets mode (default)

`cali` uses Google Sheets by default.

Required:
- `CALI_SHEET_ID=<spreadsheet-id>`
- Credentials path:
  - `CALI_GOOGLE_CREDENTIALS_JSON=<path-to-service-account-json>`
  - or `GOOGLE_APPLICATION_CREDENTIALS=<path-to-service-account-json>`

Optional:
- `CALI_SHEET_NAME=<tab-name>` (default: `Log`)

The sheet tab should use columns `A:G` as:

`Date | Day | Exercise | Level | RepsxSets | Goal | Comment`

Header row is allowed.

### 2) Local mode (optional override)

If you want file logs instead of Google Sheets:

```bash
export CALI_STORAGE=local
```

In local mode, entries are written to:

`~/cali-logger/workout/workout-<year>.log`

## Google Sheets Mode Setup (Step-by-Step)

1. Create a new Google Sheet.
2. Create or rename one tab to `Log` (or choose a different tab name and set `CALI_SHEET_NAME`).
3. Add headers in row 1:
   - `Date | Day | Exercise | Level | RepsxSets | Goal | Comment`
4. Copy your spreadsheet ID from the URL:
   - URL format: `https://docs.google.com/spreadsheets/d/<SPREADSHEET_ID>/edit...`
5. Open Google Cloud Console and create a project (or select an existing one).
6. In that project, enable **Google Sheets API**:
   - APIs & Services -> Library -> search "Google Sheets API" -> Enable
7. Create a Service Account:
   - IAM & Admin -> Service Accounts -> Create Service Account
8. Create a JSON key for the service account:
   - Service account -> Keys -> Add key -> Create new key -> JSON
9. Save the JSON key file to your machine (for example: `$HOME/.config/cali/service-account.json`).
10. Share the Google Sheet with the service account email (ends with `iam.gserviceaccount.com`) as **Editor**.
11. Set environment variables on your OS (examples below).
12. Test read access:
   - `cali -p`
13. Test write access:
   - `cali` and submit one sample entry
14. Verify the new row appears in the Google Sheet.

### Linux / macOS (bash/zsh) env setup

```bash
export CALI_SHEET_ID="your_spreadsheet_id"
export CALI_SHEET_NAME="Log"
export CALI_GOOGLE_CREDENTIALS_JSON="$HOME/.config/cali/service-account.json"
```

### Windows PowerShell env setup

```powershell
$env:CALI_SHEET_ID="your_spreadsheet_id"
$env:CALI_SHEET_NAME="Log"
$env:CALI_GOOGLE_CREDENTIALS_JSON="$HOME\.config\cali\service-account.json"
```

To persist in PowerShell (new terminals):

```powershell
[Environment]::SetEnvironmentVariable("CALI_SHEET_ID","your_spreadsheet_id","User")
[Environment]::SetEnvironmentVariable("CALI_SHEET_NAME","Log","User")
[Environment]::SetEnvironmentVariable("CALI_GOOGLE_CREDENTIALS_JSON","C:\path\service-account.json","User")
```

### Windows CMD env setup

Current terminal only:

```cmd
set CALI_SHEET_ID=your_spreadsheet_id
set CALI_SHEET_NAME=Log
set CALI_GOOGLE_CREDENTIALS_JSON=C:\path\service-account.json
```

Persistent:

```cmd
setx CALI_SHEET_ID your_spreadsheet_id
setx CALI_SHEET_NAME Log
setx CALI_GOOGLE_CREDENTIALS_JSON C:\path\service-account.json
```

Open a new CMD/PowerShell window after `setx`.

## Quick Verification

After setup, run:

```bash
cali -p
```

If it returns history (or "No workouts logged yet") without auth errors, setup is correct.

## Troubleshooting

- `CALI_SHEET_ID is required`:
  - Set `CALI_SHEET_ID`.
- `set CALI_GOOGLE_CREDENTIALS_JSON or GOOGLE_APPLICATION_CREDENTIALS`:
  - Set one credentials env var to JSON key path.
- `sheet tab "Log" not found`:
  - Create the tab or set `CALI_SHEET_NAME`.
- Permission errors with Sheets:
  - Ensure the sheet is shared with service account email as Editor.
- Want local files temporarily:
  - Set `CALI_STORAGE=local`.
