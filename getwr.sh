#!/usr/bin/env bash
t=$(mktemp)
python getCookies.py >"$t"
curl -L -b "$t" "https://docs.google.com/spreadsheets/d/1BKP3F13UTYMQZccQqDsOU-ck1Qi5x8qSucwMfIO5LnE/export?format=csv" -o spreadsheet.csv
rm "$t"
CSV_FILE="./spreadsheet.csv"
OUTPUT_FILE="./wrs"

if [ ! -f "$CSV_FILE" ]; then
  echo "Error: $CSV_FILE not found."
  exit 1
fi

# Ensure output file is clean
: >"$OUTPUT_FILE"

awk -v outfile="$OUTPUT_FILE" -f- "$CSV_FILE" <<'EOF'
BEGIN {
    FPAT = "([^,]*)|(\"[^\"]+\")"
}
{
    gsub(/\r/, "", $0)
}

# --- LINE 1 OF THE BLOCK (Row 1, 4, 7...) = Times ---
NR % 3 == 1 {
    level_num = (NR - 1) / 3
    if (level_num > 4) exit # We only need up to totalSplits (5 levels: 0-4)

    time_val = ""
    for (i = NF; i > 0; i--) {
        clean_val = $i
        gsub(/^[ \t,]+|[ \t,]+$/, "", clean_val)
        if (clean_val != "" && clean_val !~ /http/) {
            time_val = clean_val
            break
        }
    }

    # Convert the time string cleanly into milliseconds
    if (time_val != "") {
        ms = 0
        # Format can be MM:SS.mmm or SS.mmm
        if (split(time_val, parts, ":") == 2) {
            min = parts[1]
            sec_part = parts[2]
        } else {
            min = 0
            sec_part = time_val
        }

        gsub(/xxx/, "999", sec_part)

        split(sec_part, s_parts, ".")
        sec = s_parts[1]
        milli = s_parts[2]

        # Ensure milliseconds block is padded appropriately to 3 digits
        while (length(milli) < 3) milli = milli "0"
        milli = substr(milli, 1, 3)

        total_ms = (min * 60 * 1000) + (sec * 1000) + milli
        print total_ms >> outfile
    } else {
        print "0" >> outfile
    }
    next
}
EOF

rm spreadsheet.csv

echo "Successfully wrote raw ms to $OUTPUT_FILE"
