#!/bin/bash
set -e

# SARIF to HTML Converter for CodeQL Results
# Converts CodeQL SARIF output to human-readable HTML reports

SARIF_DIR="${1:-sarif-results}"
OUTPUT_DIR="${2:-html-reports}"

echo "Converting SARIF results to HTML..."
echo "Input directory: $SARIF_DIR"
echo "Output directory: $OUTPUT_DIR"

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Function to convert SARIF to HTML
convert_sarif_to_html() {
    local sarif_file="$1"
    local output_file="$2"
    
    echo "Processing: $sarif_file"
    
    # Use jq to parse SARIF and generate HTML
    # If jq is not available, use a simple Python/Node script
    if command -v jq &> /dev/null; then
        generate_html_with_jq "$sarif_file" "$output_file"
    else
        generate_html_with_python "$sarif_file" "$output_file"
    fi
}

# Generate HTML using jq (preferred method)
generate_html_with_jq() {
    local sarif_file="$1"
    local output_file="$2"
    
    # Extract data from SARIF
    local tool_name=$(jq -r '.runs[0].tool.driver.name // "CodeQL"' "$sarif_file")
    local tool_version=$(jq -r '.runs[0].tool.driver.version // "unknown"' "$sarif_file")
    local results_count=$(jq '.runs[0].results | length' "$sarif_file")
    
    # Generate HTML
    cat > "$output_file" << 'HTML_HEADER'
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>CodeQL Security Analysis Results</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: #1a1a1a;
            color: #e0e0e0;
            line-height: 1.6;
        }
        .container {
            max-width: 1400px;
            margin: 0 auto;
            padding: 20px;
        }
        header {
            background: linear-gradient(135deg, #2a2a2a 0%, #1a1a1a 100%);
            padding: 30px;
            border-radius: 8px;
            margin-bottom: 30px;
            border: 2px solid #4CAF50;
        }
        h1 {
            color: #4CAF50;
            font-size: 2.5em;
            margin-bottom: 10px;
        }
        .header-info {
            color: #aaa;
            font-size: 1.1em;
        }
        .summary {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        .summary-card {
            background: #2a2a2a;
            padding: 20px;
            border-radius: 8px;
            border-left: 4px solid #4CAF50;
        }
        .summary-card h3 {
            color: #4CAF50;
            font-size: 0.9em;
            text-transform: uppercase;
            margin-bottom: 10px;
        }
        .summary-card .value {
            font-size: 2em;
            font-weight: bold;
        }
        .results-section {
            background: #2a2a2a;
            border-radius: 8px;
            padding: 30px;
        }
        .result-item {
            background: #1a1a1a;
            border: 1px solid #333;
            border-left: 4px solid #ff9800;
            border-radius: 4px;
            padding: 20px;
            margin-bottom: 20px;
            transition: all 0.3s ease;
        }
        .result-item:hover {
            border-left-color: #4CAF50;
            box-shadow: 0 2px 8px rgba(76, 175, 80, 0.2);
        }
        .result-item.error {
            border-left-color: #f44336;
        }
        .result-item.warning {
            border-left-color: #ff9800;
        }
        .result-item.note {
            border-left-color: #2196F3;
        }
        .result-header {
            display: flex;
            justify-content: space-between;
            align-items: flex-start;
            margin-bottom: 15px;
        }
        .result-title {
            color: #fff;
            font-size: 1.2em;
            font-weight: 600;
        }
        .severity-badge {
            padding: 4px 12px;
            border-radius: 4px;
            font-size: 0.85em;
            font-weight: bold;
            text-transform: uppercase;
        }
        .severity-error {
            background: #f44336;
            color: white;
        }
        .severity-warning {
            background: #ff9800;
            color: white;
        }
        .severity-note {
            background: #2196F3;
            color: white;
        }
        .result-message {
            color: #ccc;
            margin-bottom: 15px;
            line-height: 1.8;
        }
        .result-location {
            background: #0d0d0d;
            padding: 10px 15px;
            border-radius: 4px;
            font-family: 'Courier New', monospace;
            font-size: 0.9em;
            color: #4CAF50;
            margin-top: 10px;
        }
        .location-file {
            color: #2196F3;
        }
        .location-line {
            color: #ff9800;
        }
        .no-results {
            text-align: center;
            padding: 60px 20px;
            color: #4CAF50;
        }
        .no-results h2 {
            font-size: 2em;
            margin-bottom: 10px;
        }
        .no-results p {
            color: #888;
            font-size: 1.2em;
        }
        footer {
            text-align: center;
            margin-top: 40px;
            padding-top: 20px;
            border-top: 1px solid #333;
            color: #666;
        }
        .filter-controls {
            margin-bottom: 20px;
            display: flex;
            gap: 10px;
            flex-wrap: wrap;
        }
        .filter-btn {
            padding: 8px 16px;
            border: 2px solid #333;
            background: #2a2a2a;
            color: #e0e0e0;
            border-radius: 4px;
            cursor: pointer;
            transition: all 0.3s ease;
        }
        .filter-btn:hover {
            border-color: #4CAF50;
        }
        .filter-btn.active {
            background: #4CAF50;
            border-color: #4CAF50;
            color: white;
        }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>🔒 CodeQL Security Analysis</h1>
HTML_HEADER

    echo "            <div class=\"header-info\">Tool: $tool_name $tool_version | Generated: $(date -u +"%Y-%m-%d %H:%M:%S UTC")</div>" >> "$output_file"
    echo "        </header>" >> "$output_file"
    
    # Summary section
    echo "        <div class=\"summary\">" >> "$output_file"
    echo "            <div class=\"summary-card\">" >> "$output_file"
    echo "                <h3>Total Findings</h3>" >> "$output_file"
    echo "                <div class=\"value\">$results_count</div>" >> "$output_file"
    echo "            </div>" >> "$output_file"
    
    # Count by severity
    local error_count=$(jq '[.runs[0].results[] | select(.level == "error")] | length' "$sarif_file")
    local warning_count=$(jq '[.runs[0].results[] | select(.level == "warning")] | length' "$sarif_file")
    local note_count=$(jq '[.runs[0].results[] | select(.level == "note")] | length' "$sarif_file")
    
    echo "            <div class=\"summary-card\">" >> "$output_file"
    echo "                <h3>Errors</h3>" >> "$output_file"
    echo "                <div class=\"value\" style=\"color: #f44336;\">$error_count</div>" >> "$output_file"
    echo "            </div>" >> "$output_file"
    echo "            <div class=\"summary-card\">" >> "$output_file"
    echo "                <h3>Warnings</h3>" >> "$output_file"
    echo "                <div class=\"value\" style=\"color: #ff9800;\">$warning_count</div>" >> "$output_file"
    echo "            </div>" >> "$output_file"
    echo "            <div class=\"summary-card\">" >> "$output_file"
    echo "                <h3>Notes</h3>" >> "$output_file"
    echo "                <div class=\"value\" style=\"color: #2196F3;\">$note_count</div>" >> "$output_file"
    echo "            </div>" >> "$output_file"
    echo "        </div>" >> "$output_file"
    
    # Filter controls
    cat >> "$output_file" << 'FILTER_HTML'
        <div class="filter-controls">
            <button class="filter-btn active" onclick="filterResults('all')">All</button>
            <button class="filter-btn" onclick="filterResults('error')">Errors</button>
            <button class="filter-btn" onclick="filterResults('warning')">Warnings</button>
            <button class="filter-btn" onclick="filterResults('note')">Notes</button>
        </div>
FILTER_HTML
    
    echo "        <div class=\"results-section\">" >> "$output_file"
    
    if [ "$results_count" -eq 0 ]; then
        cat >> "$output_file" << 'NO_RESULTS_HTML'
            <div class="no-results">
                <h2>✅ No Issues Found</h2>
                <p>CodeQL analysis completed successfully with no security findings.</p>
            </div>
NO_RESULTS_HTML
    else
        # Generate result items
        jq -r '.runs[0].results[] | 
            @json' "$sarif_file" | while IFS= read -r result; do
            
            rule_id=$(echo "$result" | jq -r '.ruleId // "unknown"')
            message=$(echo "$result" | jq -r '.message.text // "No message"')
            level=$(echo "$result" | jq -r '.level // "warning"')
            file=$(echo "$result" | jq -r '.locations[0].physicalLocation.artifactLocation.uri // "unknown"')
            start_line=$(echo "$result" | jq -r '.locations[0].physicalLocation.region.startLine // "0"')
            end_line=$(echo "$result" | jq -r '.locations[0].physicalLocation.region.endLine // "0"')
            
            echo "            <div class=\"result-item $level\" data-level=\"$level\">" >> "$output_file"
            echo "                <div class=\"result-header\">" >> "$output_file"
            echo "                    <div class=\"result-title\">$rule_id</div>" >> "$output_file"
            echo "                    <span class=\"severity-badge severity-$level\">$level</span>" >> "$output_file"
            echo "                </div>" >> "$output_file"
            echo "                <div class=\"result-message\">$(echo "$message" | sed 's/&/\&amp;/g; s/</\&lt;/g; s/>/\&gt;/g; s/\"/\&quot;/g')</div>" >> "$output_file"
            
            if [ "$file" != "unknown" ]; then
                echo "                <div class=\"result-location\">" >> "$output_file"
                echo "                    📁 <span class=\"location-file\">$file</span>" >> "$output_file"
                if [ "$start_line" != "0" ]; then
                    echo "                    | 📍 Lines <span class=\"location-line\">$start_line-$end_line</span>" >> "$output_file"
                fi
                echo "                </div>" >> "$output_file"
            fi
            
            echo "            </div>" >> "$output_file"
        done
    fi
    
    cat >> "$output_file" << 'HTML_FOOTER'
        </div>

        <footer>
            <p>Generated by CodeQL Security Analysis | <a href="index.html" style="color: #4CAF50;">Back to Security Home</a></p>
        </footer>
    </div>

    <script>
        function filterResults(level) {
            const items = document.querySelectorAll('.result-item');
            const buttons = document.querySelectorAll('.filter-btn');
            
            buttons.forEach(btn => {
                btn.classList.remove('active');
                if (btn.textContent.toLowerCase().includes(level) || (level === 'all' && btn.textContent === 'All')) {
                    btn.classList.add('active');
                }
            });
            
            items.forEach(item => {
                if (level === 'all' || item.dataset.level === level) {
                    item.style.display = 'block';
                } else {
                    item.style.display = 'none';
                }
            });
        }
    </script>
</body>
</html>
HTML_FOOTER
}

# Generate HTML using Python (fallback method)
generate_html_with_python() {
    local sarif_file="$1"
    local output_file="$2"
    
    python3 << PYTHON_SCRIPT "$sarif_file" "$output_file"
import json
import sys
import html
from datetime import datetime

sarif_file = sys.argv[1]
output_file = sys.argv[2]

with open(sarif_file, 'r') as f:
    data = json.load(f)

run = data.get('runs', [{}])[0]
tool = run.get('tool', {}).get('driver', {})
tool_name = tool.get('name', 'CodeQL')
tool_version = tool.get('version', 'unknown')
results = run.get('results', [])

# Count by severity
error_count = sum(1 for r in results if r.get('level') == 'error')
warning_count = sum(1 for r in results if r.get('level') == 'warning')
note_count = sum(1 for r in results if r.get('level') == 'note')

html = f'''<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>CodeQL Security Analysis Results</title>
    <style>
        body {{ font-family: Arial, sans-serif; background: #1a1a1a; color: #e0e0e0; margin: 20px; }}
        .container {{ max-width: 1200px; margin: 0 auto; }}
        h1 {{ color: #4CAF50; }}
        .summary {{ background: #2a2a2a; padding: 20px; border-radius: 8px; margin-bottom: 20px; }}
        .result {{ background: #2a2a2a; padding: 15px; margin: 10px 0; border-left: 4px solid #ff9800; border-radius: 4px; }}
        .result.error {{ border-left-color: #f44336; }}
        .result.warning {{ border-left-color: #ff9800; }}
        .result.note {{ border-left-color: #2196F3; }}
    </style>
</head>
<body>
    <div class="container">
        <h1>🔒 CodeQL Security Analysis</h1>
        <div class="summary">
            <p><strong>Tool:</strong> {tool_name} {tool_version}</p>
            <p><strong>Generated:</strong> {datetime.utcnow().strftime("%Y-%m-%d %H:%M:%S UTC")}</p>
            <p><strong>Total Findings:</strong> {len(results)}</p>
            <p><strong>Errors:</strong> {error_count} | <strong>Warnings:</strong> {warning_count} | <strong>Notes:</strong> {note_count}</p>
        </div>
'''

if not results:
    html += '<div class="result"><h2>✅ No Issues Found</h2></div>'
else:
    for result in results:
        rule_id = result.get('ruleId', 'unknown')
        message = result.get('message', {}).get('text', 'No message')
        level = result.get('level', 'warning')
        
        locations = result.get('locations', [{}])
        if locations:
            loc = locations[0].get('physicalLocation', {})
            file_path = loc.get('artifactLocation', {}).get('uri', 'unknown')
            region = loc.get('region', {})
            line = region.get('startLine', 0)
        else:
            file_path = 'unknown'
            line = 0
        
        html += f'''
        <div class="result {html.escape(level)}">
            <h3>{html.escape(rule_id)}</h3>
            <p>{html.escape(message)}</p>
            <p><small>📁 {html.escape(file_path)}:{line}</small></p>
        </div>
'''

html += '''
    </div>
</body>
</html>
'''

with open(output_file, 'w') as f:
    f.write(html)

print(f"Generated {output_file}")
PYTHON_SCRIPT
}

# Find and convert all SARIF files
if [ -d "$SARIF_DIR" ]; then
    sarif_files=$(find "$SARIF_DIR" -name "*.sarif" -type f)
    
    if [ -z "$sarif_files" ]; then
        echo "No SARIF files found in $SARIF_DIR"
        # Create a placeholder HTML file
        cat > "$OUTPUT_DIR/codeql-results.html" << 'EOF'
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>CodeQL Results - No Data</title>
    <style>
        body { font-family: Arial, sans-serif; background: #1a1a1a; color: #e0e0e0; 
               display: flex; align-items: center; justify-content: center; height: 100vh; margin: 0; }
        .container { text-align: center; }
        h1 { color: #4CAF50; }
    </style>
</head>
<body>
    <div class="container">
        <h1>⏳ CodeQL Analysis Pending</h1>
        <p>Results will appear here after the first scan completes.</p>
    </div>
</body>
</html>
EOF
        echo "Created placeholder HTML file"
    else
        # Convert the first (and typically only) SARIF file
        for sarif in $sarif_files; do
            convert_sarif_to_html "$sarif" "$OUTPUT_DIR/codeql-results.html"
            break  # Only process the first file for now
        done
        echo "Conversion complete!"
    fi
else
    echo "Error: SARIF directory not found: $SARIF_DIR"
    exit 1
fi

echo "HTML reports generated in: $OUTPUT_DIR"
