# docs: images used by the main README

The main README links to these files. Add them after running the stack with `docker compose up --build`.

| File | What to capture |
|---|---|
| `dashboard.png` | http://localhost:3000 Dashboard page after sending a few test events (stats cards and charts visible) |
| `dlq.png` | http://localhost:3000/dlq after pressing "Send bad event" on the Events page, with one row expanded |
| `outage-demo.gif` | A terminal recording of `scripts/demo-db-outage.ps1` (or `.sh`) from start to the `PASS` line |

Tips: keep the GIF under about 10 MB (ScreenToGif on Windows works well), and crop the browser to the page content.
