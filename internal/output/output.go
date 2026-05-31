package output

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const baseURL = "https://lists.calmcacil.dev"

type Show struct {
	TVDBID int    `json:"tvdbId"`
	Title  string `json:"title,omitempty"`
}

func WriteSeasonJSON(dir, category, season string, year int, shows []Show) error {
	yearDir := filepath.Join(dir, fmt.Sprintf("%d", year))
	filename := fmt.Sprintf("%s-%s.json", strings.ToLower(season), category)
	return writeJSON(yearDir, filename, shows)
}

func WriteYearJSON(dir, category string, year int, shows []Show) error {
	yearDir := filepath.Join(dir, fmt.Sprintf("%d", year))
	filename := fmt.Sprintf("%s.json", category)
	return writeJSON(yearDir, filename, shows)
}

func writeJSON(dir, filename string, shows []Show) error {
	if shows == nil {
		shows = []Show{}
	}
	data, err := json.Marshal(shows)
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}
	outPath := filepath.Join(dir, filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}
	if err := os.WriteFile(outPath, data, 0644); err != nil {
		return fmt.Errorf("write JSON file: %w", err)
	}
	return nil
}

func WriteAllJSON(outputDir, category string, seasonal map[string][]Show) error {
	byYear := map[int][]Show{}

	years := map[int]bool{}

	for key, shows := range seasonal {
		parts := strings.SplitN(key, "-", 2)
		if len(parts) != 2 {
			continue
		}
		season := parts[0]
		var year int
		if _, err := fmt.Sscanf(parts[1], "%d", &year); err != nil {
			continue
		}
		if err := WriteSeasonJSON(outputDir, category, season, year, shows); err != nil {
			return fmt.Errorf("write %s: %w", key, err)
		}
		byYear[year] = append(byYear[year], shows...)
		years[year] = true
	}

	for year, shows := range byYear {
		if err := WriteYearJSON(outputDir, category, year, shows); err != nil {
			return fmt.Errorf("write year %d: %w", year, err)
		}
	}

	if category == "series" {
		if err := WriteIndex(outputDir, years); err != nil {
			return fmt.Errorf("write index: %w", err)
		}
	}

	return nil
}

func WriteIndex(dir string, years map[int]bool) error {
	var yList []int
	for y := range years {
		yList = append(yList, y)
	}

	seasonLabels := []string{"Winter", "Spring", "Summer", "Fall"}
	seasonKeys := []string{"winter", "spring", "summer", "fall"}

	html := `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Seasonal Anime Lists</title>
<style>
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    background: #0d1117; color: #c9d1d9;
    display: flex; justify-content: center; align-items: center;
    min-height: 100vh; padding: 20px;
  }
  .card {
    background: #161b22; border: 1px solid #30363d; border-radius: 12px;
    padding: 40px; max-width: 520px; width: 100%;
  }
  h1 { font-size: 22px; margin-bottom: 6px; }
  p { color: #8b949e; font-size: 14px; margin-bottom: 24px; }
  .row { display: flex; gap: 12px; margin-bottom: 16px; }
  select {
    flex: 1; padding: 10px 12px; border-radius: 8px;
    background: #0d1117; color: #c9d1d9; border: 1px solid #30363d;
    font-size: 14px; cursor: pointer;
  }
  .url-box {
    display: flex; gap: 8px; margin-bottom: 20px;
  }
  .url-box input {
    flex: 1; padding: 10px 12px; border-radius: 8px;
    background: #0d1117; color: #8b949e; border: 1px solid #30363d;
    font-size: 13px; font-family: monospace;
  }
  .btn {
    padding: 10px 18px; border-radius: 8px; border: none;
    font-size: 14px; font-weight: 600; cursor: pointer;
    background: #238636; color: #fff; white-space: nowrap;
  }
  .btn:hover { background: #2ea043; }
  .btn:active { background: #1e7e34; }
  .badge {
    display: inline-block; padding: 3px 8px; border-radius: 4px;
    font-size: 11px; font-weight: 600; margin-left: 8px;
  }
  .badge-sonarr { background: #1f6feb; color: #fff; }
  .copied { background: #30363d; color: #8b949e; padding: 8px 12px;
    border-radius: 8px; font-size: 13px; text-align: center; display: none; }
</style>
</head>
<body>
<div class="card">
  <h1>Seasonal Anime Lists</h1>
  <p>Select a year and season, then copy the Sonarr import URL.</p>

  <div class="row">
    <select id="year">
`
	for _, y := range yList {
		html += fmt.Sprintf("      <option value=\"%d\">%d</option>\n", y, y)
	}
	html += `    </select>
    <select id="season">
      <option value="all">Entire Year</option>
`
	for i, s := range seasonLabels {
		html += fmt.Sprintf("      <option value=\"%s\">%s</option>\n", seasonKeys[i], s)
	}
	html += `    </select>
  </div>

  <div class="row" style="gap:8px; flex-wrap:wrap;">
    <span class="badge badge-sonarr">Sonarr</span>
  </div>

  <div class="url-box">
    <input type="text" id="url" readonly value="` + baseURL + `/2026/series.json">
    <button class="btn" onclick="copyUrl()">Copy</button>
  </div>
  <div class="copied" id="copied">URL copied to clipboard</div>
</div>

<script>
const base = '` + baseURL + `';
const yearSel = document.getElementById('year');
const seasonSel = document.getElementById('season');
const urlInput = document.getElementById('url');

function updateUrl() {
  const y = yearSel.value;
  const s = seasonSel.value;
  if (s === 'all') {
    urlInput.value = base + '/' + y + '/series.json';
  } else {
    urlInput.value = base + '/' + y + '/' + s + '-series.json';
  }
}

function copyUrl() {
  urlInput.select();
  navigator.clipboard.writeText(urlInput.value);
  document.getElementById('copied').style.display = 'block';
  setTimeout(() => document.getElementById('copied').style.display = 'none', 2000);
}

yearSel.addEventListener('change', updateUrl);
seasonSel.addEventListener('change', updateUrl);
</script>
</body>
</html>`

	indexPath := filepath.Join(dir, "index.html")
	return os.WriteFile(indexPath, []byte(html), 0644)
}
