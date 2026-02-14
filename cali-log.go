package main

// run after updating:
// go build -o ~/.local/bin/cali .

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type WorkoutEntry struct {
	Date     string
	Day      string
	Exercise string
	Level    string
	RepsSets string
	Goal     string
	Comment  string
	RowIndex int64
}

type Storage interface {
	Append(entry WorkoutEntry) error
	Recent(limit int) ([]WorkoutEntry, error)
	SearchByDate(date string) ([]WorkoutEntry, error)
	RemoveByDateIndex(date string, index int) error
	LastTrainingDay() (string, string, error)
}

const (
	defaultSheetName = "Log"
	dateLayout       = "2006-01-02"
)

// Goal map: Exercise -> Level -> Goal
var goals = map[string]map[string]string{
	"Pushups": {
		"Wall":         "50x3",
		"Incline":      "40x3",
		"Kneeling":     "30x3",
		"Half":         "25x2",
		"Full":         "20x2",
		"Close":        "20x2",
		"Uneven":       "20x2",
		"Half One-Arm": "20x2",
		"Lever":        "20x2",
		"One-Arm":      "100x1",
	},
	"Squats": {
		"Shoulderstand":    "50x3",
		"Jackknife":        "40x3",
		"Supported":        "30x3",
		"Half":             "50x2",
		"Full":             "30x2",
		"Close":            "20x2",
		"Uneven":           "20x2",
		"Half One-Leg":     "20x2",
		"Assisted One-Leg": "20x2",
		"One-Leg":          "50x2",
	},
	"Pullups": {
		"Vertical":         "40x3",
		"Horizontal":       "30x3",
		"Jackknife":        "20x3",
		"Half":             "15x2",
		"Full":             "10x2",
		"Close":            "10x2",
		"Uneven":           "9x2",
		"Half One-Arm":     "8x2",
		"Assisted One-Arm": "7x2",
		"One-Arm":          "6x2",
	},
	"Leg Raises": {
		"Knee Tuck":    "40x3",
		"Knee Raise":   "35x3",
		"Bent Leg":     "30x3",
		"Frog":         "25x3",
		"Flat":         "20x2",
		"Hanging Knee": "15x2",
		"Hanging Bent": "15x2",
		"Partial":      "15x2",
		"Hanging":      "30x2",
	},
	"Bridges": {
		"Short":          "50x3",
		"Straight":       "40x3",
		"Angled":         "30x3",
		"Head":           "25x2",
		"Half":           "20x2",
		"Full":           "15x2",
		"Wall Down":      "10x2",
		"Wall Up":        "8x2",
		"Closing":        "6x2",
		"Stand-to-Stand": "10-30x2",
	},
	"Handstand Push-ups": {
		"Wall Headstand": "2min",
		"Crow":           "1min",
		"Wall":           "2min",
		"Half":           "20x2",
		"Full":           "15x2",
		"Close":          "12x2",
		"Uneven":         "10x2",
		"Half One-Arm":   "8x2",
		"Lever":          "6x2",
		"One-Arm":        "5x2",
	},
}

// Ordered list of exercises
var exercises = []string{
	"Pushups",
	"Squats",
	"Pullups",
	"Leg Raises",
	"Bridges",
	"Handstand Push-ups",
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "open":
			if len(os.Args) < 3 {
				fmt.Println("Usage: cali open <workout-template>")
				os.Exit(1)
			}
			if err := openResource(os.Args[2]); err != nil {
				fmt.Fprintf(os.Stderr, "Error opening resource: %v\n", err)
				os.Exit(1)
			}
			return
		case "--template":
			if err := openResource("workout-template"); err != nil {
				fmt.Fprintf(os.Stderr, "Error opening resource: %v\n", err)
				os.Exit(1)
			}
			return
		case "-h", "--h", "--help":
			showHelp()
			return
		case "-p", "--print", "--history":
			storage, err := newStorage()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error configuring storage: %v\n", err)
				os.Exit(1)
			}
			showHistory(storage)
			return
		case "-s", "--search":
			if len(os.Args) < 3 {
				fmt.Println("Usage: cali -s <date>")
				fmt.Println("Example: cali -s 2026-01-24")
				os.Exit(1)
			}
			storage, err := newStorage()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error configuring storage: %v\n", err)
				os.Exit(1)
			}
			searchByDate(storage, os.Args[2])
			return
		case "-r", "--remove":
			storage, err := newStorage()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error configuring storage: %v\n", err)
				os.Exit(1)
			}
			removeEntry(storage)
			return
		}
	}

	storage, err := newStorage()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error configuring storage: %v\n", err)
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)

	printDayPlan()

	if day, date, err := storage.LastTrainingDay(); err == nil && day != "" {
		fmt.Printf("Previous training day: %s (%s)\n\n", day, date)
	}

	fmt.Print("Day (A/B/C): ")
	day, _ := reader.ReadString('\n')
	day = strings.TrimSpace(day)

	exercise := chooseExercise(reader)
	level := chooseLevel(reader, exercise)

	fmt.Print("Reps×Sets: ")
	repsSets, _ := reader.ReadString('\n')
	repsSets = strings.TrimSpace(repsSets)

	fmt.Print("Comment (optional): ")
	comment, _ := reader.ReadString('\n')
	comment = strings.TrimSpace(comment)

	goal := resolveGoal(exercise, level)
	date := time.Now().Format(dateLayout)

	entry := WorkoutEntry{
		Date:     date,
		Day:      day,
		Exercise: exercise,
		Level:    level,
		RepsSets: repsSets,
		Goal:     goal,
		Comment:  comment,
	}

	if err := storage.Append(entry); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing workout: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✓ Logged successfully")
}
func newStorage() (Storage, error) {
	if strings.EqualFold(os.Getenv("CALI_STORAGE"), "local") {
		return newFileStorage()
	}
	return newSheetsStorage()
}

func chooseExercise(reader *bufio.Reader) string {
	fmt.Println("\nChoose Exercise:")
	for i, ex := range exercises {
		fmt.Printf("  %d. %s\n", i+1, ex)
	}
	fmt.Print("Enter number: ")

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	choice, err := strconv.Atoi(input)

	if err != nil || choice < 1 || choice > len(exercises) {
		fmt.Println("Invalid choice, defaulting to Pushups")
		return exercises[0]
	}

	return exercises[choice-1]
}

func chooseLevel(reader *bufio.Reader, exercise string) string {
	levels := getLevelsForExercise(exercise)

	fmt.Printf("\nChoose Level for %s:\n", exercise)
	for i, lv := range levels {
		goal := goals[exercise][lv]
		fmt.Printf("  %d. %-20s (goal: %s)\n", i+1, lv, goal)
	}
	fmt.Print("Enter number: ")

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	choice, err := strconv.Atoi(input)

	if err != nil || choice < 1 || choice > len(levels) {
		fmt.Println("Invalid choice, defaulting to first level")
		return levels[0]
	}

	return levels[choice-1]
}

func printDayPlan() {
	fmt.Println("Day plan:")
	fmt.Println("  Day A")
	fmt.Println("    - Pushups")
	fmt.Println("    - Squats")
	fmt.Println("  Day B")
	fmt.Println("    - Pullups")
	fmt.Println("    - Leg Raises")
	fmt.Println("  Day C")
	fmt.Println("    - Bridges")
	fmt.Println("    - Handstand Push-ups")
	fmt.Println()
}

func getLevelsForExercise(exercise string) []string {
	levelOrder := map[string][]string{
		"Pushups": {
			"Wall", "Incline", "Kneeling", "Half", "Full",
			"Close", "Uneven", "Half One-Arm", "Lever", "One-Arm",
		},
		"Squats": {
			"Shoulderstand", "Jackknife", "Supported", "Half", "Full",
			"Close", "Uneven", "Half One-Leg", "Assisted One-Leg", "One-Leg",
		},
		"Pullups": {
			"Vertical", "Horizontal", "Jackknife", "Half", "Full",
			"Close", "Uneven", "Half One-Arm", "Assisted One-Arm", "One-Arm",
		},
		"Leg Raises": {
			"Knee Tuck", "Knee Raise", "Bent Leg", "Frog", "Flat",
			"Hanging Knee", "Hanging Bent", "Partial", "Hanging",
		},
		"Bridges": {
			"Short", "Straight", "Angled", "Head", "Half",
			"Full", "Wall Down", "Wall Up", "Closing", "Stand-to-Stand",
		},
		"Handstand Push-ups": {
			"Wall Headstand", "Crow", "Wall", "Half", "Full",
			"Close", "Uneven", "Half One-Arm", "Lever", "One-Arm",
		},
	}

	if levels, ok := levelOrder[exercise]; ok {
		return levels
	}
	return []string{}
}

func openResource(name string) error {
	if name != "workout-template" {
		return fmt.Errorf("unknown resource %q (use workout-template)", name)
	}

	const templateURL = "https://drive.google.com/file/d/19zXstmNsSoT6hmseO-nU-h2NNiIK-X2R/view?usp=drive_link"
	cmd := exec.Command("xdg-open", templateURL)
	return cmd.Start()
}

func resolveGoal(exercise, level string) string {
	if levels, ok := goals[exercise]; ok {
		if goal, ok := levels[level]; ok {
			return goal
		}
	}
	return "-"
}

func showHistory(storage Storage) {
	entries, err := storage.Recent(10)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading workout history: %v\n", err)
		os.Exit(1)
	}

	if len(entries) == 0 {
		fmt.Println("No workouts logged yet")
		return
	}

	fmt.Println("Last 10 workouts:")
	fmt.Println(strings.Repeat("-", 80))
	for _, entry := range entries {
		fmt.Printf("%s | Day %s | %s - %s | %s → %s | %s\n",
			entry.Date, entry.Day, entry.Exercise, entry.Level, entry.RepsSets, entry.Goal, entry.Comment)
	}
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("Total: %d workout(s)\n", len(entries))
}

func searchByDate(storage Storage, dateStr string) {
	if _, err := time.Parse(dateLayout, dateStr); err != nil {
		fmt.Println("Invalid date format. Use YYYY-MM-DD (e.g., 2026-01-24)")
		os.Exit(1)
	}

	entries, err := storage.SearchByDate(dateStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error searching workouts: %v\n", err)
		os.Exit(1)
	}

	if len(entries) == 0 {
		fmt.Printf("No workouts found for %s\n", dateStr)
		return
	}

	fmt.Printf("Workouts for %s:\n", dateStr)
	fmt.Println(strings.Repeat("-", 80))
	for i, entry := range entries {
		fmt.Printf("[%d] Day %s | %s - %s | %s → %s | %s\n",
			i+1, entry.Day, entry.Exercise, entry.Level, entry.RepsSets, entry.Goal, entry.Comment)
	}
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("Total: %d workout(s)\n", len(entries))
}

func removeEntry(storage Storage) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter date to search (YYYY-MM-DD): ")
	dateStr, _ := reader.ReadString('\n')
	dateStr = strings.TrimSpace(dateStr)

	if _, err := time.Parse(dateLayout, dateStr); err != nil {
		fmt.Println("Invalid date format. Use YYYY-MM-DD (e.g., 2026-01-24)")
		os.Exit(1)
	}

	entries, err := storage.SearchByDate(dateStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error searching workouts: %v\n", err)
		os.Exit(1)
	}

	if len(entries) == 0 {
		fmt.Printf("No workouts found for %s\n", dateStr)
		return
	}

	fmt.Printf("\nWorkouts for %s:\n", dateStr)
	fmt.Println(strings.Repeat("-", 80))
	for i, entry := range entries {
		fmt.Printf("[%d] Day %s | %s - %s | %s → %s | %s\n",
			i+1, entry.Day, entry.Exercise, entry.Level, entry.RepsSets, entry.Goal, entry.Comment)
	}
	fmt.Println(strings.Repeat("-", 80))

	fmt.Print("\nEnter number to remove (0 to cancel): ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	choice, err := strconv.Atoi(input)
	if err != nil || choice < 0 || choice > len(entries) {
		fmt.Println("Invalid choice")
		return
	}
	if choice == 0 {
		fmt.Println("Cancelled")
		return
	}

	if err := storage.RemoveByDateIndex(dateStr, choice-1); err != nil {
		fmt.Fprintf(os.Stderr, "Error removing entry: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✓ Entry removed successfully")
}

func showHelp() {
	fmt.Println("Calisthenics Workout Logger")
	fmt.Println("\nUsage:")
	fmt.Println("  cali                    Log a new workout")
	fmt.Println("  cali -p, --print        Show last 10 workouts")
	fmt.Println("  cali -s <date>          Search workouts by date (YYYY-MM-DD)")
	fmt.Println("  cali -r, --remove       Remove a workout entry")
	fmt.Println("  cali --help             Show this help message")
	fmt.Println("  cali --template         Open workout template link")
	fmt.Println("  cali open workout-template  Open workout template link")
	fmt.Println("\nStorage backends:")
	fmt.Println("  Default: Google Sheets")
	fmt.Println("  Local files override: set CALI_STORAGE=local")
	fmt.Println("  Local path: ~/cali-logger/workout")
	fmt.Println("\nGoogle Sheets env vars:")
	fmt.Println("  CALI_SHEET_ID=<spreadsheet-id> (required)")
	fmt.Println("  CALI_SHEET_NAME=<tab-name>     (optional, default: Log)")
	fmt.Println("  CALI_GOOGLE_CREDENTIALS_JSON=<service-account-json-path>")
	fmt.Println("  or GOOGLE_APPLICATION_CREDENTIALS can be used instead")
	fmt.Println("\nExamples:")
	fmt.Println("  cali -s 2026-01-24")
	fmt.Println("  cali -p")
	fmt.Println("  CALI_STORAGE=local cali -p")
}

func parseLogLine(line string) (WorkoutEntry, bool) {
	parts := strings.Split(line, "|")
	if len(parts) < 7 {
		return WorkoutEntry{}, false
	}
	return WorkoutEntry{
		Date:     parts[0],
		Day:      parts[1],
		Exercise: parts[2],
		Level:    parts[3],
		RepsSets: parts[4],
		Goal:     parts[5],
		Comment:  parts[6],
	}, true
}

func serializeLogEntry(entry WorkoutEntry) string {
	return fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s\n",
		entry.Date, entry.Day, entry.Exercise, entry.Level, entry.RepsSets, entry.Goal, entry.Comment)
}

type fileStorage struct {
	logDir string
}

func newFileStorage() (Storage, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	return &fileStorage{
		logDir: filepath.Join(homeDir, "cali-logger", "workout"),
	}, nil
}

func (f *fileStorage) Append(entry WorkoutEntry) error {
	year := yearFromDate(entry.Date)
	logFile := filepath.Join(f.logDir, fmt.Sprintf("workout-%d.log", year))

	if err := os.MkdirAll(f.logDir, 0755); err != nil {
		return err
	}

	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(serializeLogEntry(entry))
	return err
}

func (f *fileStorage) Recent(limit int) ([]WorkoutEntry, error) {
	year := time.Now().Year()
	logFile := filepath.Join(f.logDir, fmt.Sprintf("workout-%d.log", year))

	file, err := os.Open(logFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []WorkoutEntry{}, nil
		}
		return nil, err
	}
	defer file.Close()

	var entries []WorkoutEntry
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		entry, ok := parseLogLine(strings.TrimSpace(scanner.Text()))
		if ok {
			entries = append(entries, entry)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if len(entries) <= limit {
		return entries, nil
	}
	return entries[len(entries)-limit:], nil
}

func (f *fileStorage) SearchByDate(date string) ([]WorkoutEntry, error) {
	year := date[:4]
	logFile := filepath.Join(f.logDir, fmt.Sprintf("workout-%s.log", year))

	file, err := os.Open(logFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []WorkoutEntry{}, nil
		}
		return nil, err
	}
	defer file.Close()

	var results []WorkoutEntry
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, date) {
			continue
		}
		entry, ok := parseLogLine(line)
		if ok {
			results = append(results, entry)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func (f *fileStorage) RemoveByDateIndex(date string, index int) error {
	year := date[:4]
	logFile := filepath.Join(f.logDir, fmt.Sprintf("workout-%s.log", year))

	file, err := os.Open(logFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("no workout log found for year %s", year)
		}
		return err
	}
	defer file.Close()

	var allLines []string
	var matchingLineIdx []int

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		line := scanner.Text()
		allLines = append(allLines, line)
		if strings.HasPrefix(line, date) {
			matchingLineIdx = append(matchingLineIdx, lineNum)
		}
		lineNum++
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	if index < 0 || index >= len(matchingLineIdx) {
		return fmt.Errorf("invalid remove index")
	}

	toRemove := matchingLineIdx[index]
	allLines = append(allLines[:toRemove], allLines[toRemove+1:]...)

	dst, err := os.Create(logFile)
	if err != nil {
		return err
	}
	defer dst.Close()

	for _, line := range allLines {
		if _, err := dst.WriteString(line + "\n"); err != nil {
			return err
		}
	}
	return nil
}

func (f *fileStorage) LastTrainingDay() (string, string, error) {
	year := time.Now().Year()
	logFile := filepath.Join(f.logDir, fmt.Sprintf("workout-%d.log", year))

	file, err := os.Open(logFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", "", nil
		}
		return "", "", err
	}
	defer file.Close()

	var last WorkoutEntry
	var found bool
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		entry, ok := parseLogLine(strings.TrimSpace(scanner.Text()))
		if ok {
			last = entry
			found = true
		}
	}
	if err := scanner.Err(); err != nil {
		return "", "", err
	}
	if !found {
		return "", "", nil
	}

	return last.Day, last.Date, nil
}

func yearFromDate(date string) int {
	if len(date) < 4 {
		return time.Now().Year()
	}
	year, err := strconv.Atoi(date[:4])
	if err != nil {
		return time.Now().Year()
	}
	return year
}

type sheetsStorage struct {
	ctx           context.Context
	svc           *sheets.Service
	spreadsheetID string
	sheetName     string
	sheetID       int64
}

func newSheetsStorage() (Storage, error) {
	spreadsheetID := strings.TrimSpace(os.Getenv("CALI_SHEET_ID"))
	if spreadsheetID == "" {
		return nil, fmt.Errorf("CALI_SHEET_ID is required (Google Sheets is default; set CALI_STORAGE=local to use local files)")
	}

	sheetName := strings.TrimSpace(os.Getenv("CALI_SHEET_NAME"))
	if sheetName == "" {
		sheetName = defaultSheetName
	}

	credPath := strings.TrimSpace(os.Getenv("CALI_GOOGLE_CREDENTIALS_JSON"))
	if credPath == "" {
		credPath = strings.TrimSpace(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	}
	if credPath == "" {
		return nil, fmt.Errorf("set CALI_GOOGLE_CREDENTIALS_JSON or GOOGLE_APPLICATION_CREDENTIALS")
	}

	ctx := context.Background()
	svc, err := sheets.NewService(
		ctx,
		option.WithCredentialsFile(credPath),
		option.WithScopes(sheets.SpreadsheetsScope),
	)
	if err != nil {
		return nil, fmt.Errorf("creating sheets service: %w", err)
	}

	resp, err := svc.Spreadsheets.Get(spreadsheetID).Fields("sheets.properties").Do()
	if err != nil {
		return nil, fmt.Errorf("reading spreadsheet metadata: %w", err)
	}

	var foundSheetID int64 = -1
	for _, sh := range resp.Sheets {
		if sh.Properties != nil && sh.Properties.Title == sheetName {
			foundSheetID = sh.Properties.SheetId
			break
		}
	}
	if foundSheetID == -1 {
		return nil, fmt.Errorf("sheet tab %q not found in spreadsheet", sheetName)
	}

	return &sheetsStorage{
		ctx:           ctx,
		svc:           svc,
		spreadsheetID: spreadsheetID,
		sheetName:     sheetName,
		sheetID:       foundSheetID,
	}, nil
}

func (s *sheetsStorage) Append(entry WorkoutEntry) error {
	values := [][]interface{}{
		{entry.Date, entry.Day, entry.Exercise, entry.Level, entry.RepsSets, entry.Goal, entry.Comment},
	}
	_, err := s.svc.Spreadsheets.Values.Append(
		s.spreadsheetID,
		fmt.Sprintf("%s!A:G", s.sheetName),
		&sheets.ValueRange{Values: values},
	).ValueInputOption("RAW").InsertDataOption("INSERT_ROWS").Context(s.ctx).Do()
	return err
}

func (s *sheetsStorage) Recent(limit int) ([]WorkoutEntry, error) {
	entries, err := s.readAllEntries()
	if err != nil {
		return nil, err
	}
	if len(entries) <= limit {
		return entries, nil
	}
	return entries[len(entries)-limit:], nil
}

func (s *sheetsStorage) SearchByDate(date string) ([]WorkoutEntry, error) {
	entries, err := s.readAllEntries()
	if err != nil {
		return nil, err
	}

	var results []WorkoutEntry
	for _, entry := range entries {
		if entry.Date == date {
			results = append(results, entry)
		}
	}
	return results, nil
}

func (s *sheetsStorage) RemoveByDateIndex(date string, index int) error {
	entries, err := s.readAllEntries()
	if err != nil {
		return err
	}

	var matches []WorkoutEntry
	for _, entry := range entries {
		if entry.Date == date {
			matches = append(matches, entry)
		}
	}

	if index < 0 || index >= len(matches) {
		return fmt.Errorf("invalid remove index")
	}

	targetRow := matches[index].RowIndex
	req := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				DeleteDimension: &sheets.DeleteDimensionRequest{
					Range: &sheets.DimensionRange{
						SheetId:    s.sheetID,
						Dimension:  "ROWS",
						StartIndex: targetRow,
						EndIndex:   targetRow + 1,
					},
				},
			},
		},
	}

	_, err = s.svc.Spreadsheets.BatchUpdate(s.spreadsheetID, req).Context(s.ctx).Do()
	return err
}

func (s *sheetsStorage) LastTrainingDay() (string, string, error) {
	entries, err := s.readAllEntries()
	if err != nil {
		return "", "", err
	}
	if len(entries) == 0 {
		return "", "", nil
	}
	last := entries[len(entries)-1]
	return last.Day, last.Date, nil
}

func (s *sheetsStorage) readAllEntries() ([]WorkoutEntry, error) {
	resp, err := s.svc.Spreadsheets.Values.Get(
		s.spreadsheetID,
		fmt.Sprintf("%s!A:G", s.sheetName),
	).Context(s.ctx).Do()
	if err != nil {
		return nil, err
	}

	var entries []WorkoutEntry
	for rowIndex, row := range resp.Values {
		entry := WorkoutEntry{
			Date:     valueAt(row, 0),
			Day:      valueAt(row, 1),
			Exercise: valueAt(row, 2),
			Level:    valueAt(row, 3),
			RepsSets: valueAt(row, 4),
			Goal:     valueAt(row, 5),
			Comment:  valueAt(row, 6),
			RowIndex: int64(rowIndex),
		}

		if entry.Date == "" {
			continue
		}
		if strings.EqualFold(entry.Date, "date") {
			continue
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func valueAt(row []interface{}, idx int) string {
	if idx < 0 || idx >= len(row) {
		return ""
	}
	return fmt.Sprint(row[idx])
}
